package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	db *pgxpool.Pool
	jwtSecret []byte
	dummyHash []byte // used to equalize login timing for unknown emails
}

func New() (http.Handler, error) {
	dbURL := os.Getenv("DATABASE_URL")
	secret := os.Getenv("JWT_SECRET")
	if dbURL == "" || len(secret) < 32 {
		return nil, errors.New("DATABASE_URL and JWT_SECRET (>= 32 chars) must be set")
	}

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	cfg.MaxConns = 3

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcryptCost)
	if err != nil {
		return nil, err
	}

	app := &App{
		db: pool,
		jwtSecret: []byte(secret),
		dummyHash: dummy,
	}
	return app.routes(), nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	auth := a.requireAuth

	mux.HandleFunc("POST /auth/register", a.register)
	mux.Handle("POST /auth/login", http.HandlerFunc(a.login))

	mux.Handle("POST /patients", auth(a.createPatient))
	mux.Handle("GET /patients", auth(a.listPatients))

	mux.Handle("POST /patients/{id}/records", auth(a.addRecord))
	mux.Handle("PATCH /records/{id}", auth(a.editRecord))
	mux.Handle("GET /patients/{id}/records", auth(a.listRecords))
	mux.Handle("GET /records/{id}", auth(a.getRecord))
	mux.Handle("GET /patients/{id}/normals", auth(a.normals))
	mux.Handle("GET /patients/{id}/growth", auth(a.growth))
	mux.Handle("POST /patients/{id}/logs", auth(a.addLog))
	mux.Handle("GET /patients/{id}/logs", auth(a.listLogs))
	mux.Handle("POST /patients/{id}/vaccines/{vaccine_id}/complete", auth(a.completeVaccine))
	mux.Handle("GET /patients/{id}/vaccines", auth(a.listVaccines))

	mux.Handle("GET /patients/{id}/activity", auth(a.patientActivity))

	return mux
}

