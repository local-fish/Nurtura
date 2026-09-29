package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Patient struct {
	Id string `json:"id"`
	Name string `json:"name"`
	DateOfBirth time.Time `json:"-"`
	DOB string `json:"date_of_birth"`
	Sex string `json:"sex"`
	CreatedAt time.Time `json:"created_at"`
}

const patientCols = `id, name, date_of_birth, sex, created_at`

func scanPatient(row pgx.Row) (*Patient, error) {
	var p Patient
	if err := row.Scan(&p.Id, &p.Name, &p.DateOfBirth, &p.Sex, &p.CreatedAt); err != nil {
		return nil, err
	}
	p.DOB = p.DateOfBirth.Format(time.DateOnly)
	return &p, nil
}

// loadPatient fetches a patient only if it belongs to uid. Ownership is in the query.
func (a *App) loadPatient(ctx context.Context, uid, pid string) (*Patient, error) {
	return scanPatient(a.db.QueryRow(ctx,
		`SELECT `+patientCols+` FROM patients WHERE id = $1 AND user_id = $2`, pid, uid))
}

// ownedPatient resolves {id} to an owned patient, writing 400/404/500 on failure.
func (a *App) ownedPatient(w http.ResponseWriter, r *http.Request) (*Patient, bool) {
	pid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "patient not found")
		return nil, false
	}
	p, err := a.loadPatient(r.Context(), userId(r), pid)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "patient not found")
		return nil, false
	}
	if err != nil {
		serverErr(w, r, err)
		return nil, false
	}
	return p, true
}

func (a *App) createPatient(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string `json:"name"`
		DateOfBirth string `json:"date_of_birth"`
		Sex         string `json:"sex"`
	}
	if err := decode(w, r, &in, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Sex = strings.ToUpper(strings.TrimSpace(in.Sex))
	dob, err := time.Parse(time.DateOnly, in.DateOfBirth)
	switch {
	case in.Name == "" || len(in.Name) > 200:
		writeErr(w, http.StatusBadRequest, "name is required (max 200 chars)")
		return
	case err != nil || dob.After(time.Now()):
		writeErr(w, http.StatusBadRequest, "date_of_birth must be YYYY-MM-DD and not in the future")
		return
	case in.Sex != "M" && in.Sex != "F":
		writeErr(w, http.StatusBadRequest, "sex must be M or F")
		return
	}
	p, err := scanPatient(a.db.QueryRow(r.Context(),
		`INSERT INTO patients (user_id, name, date_of_birth, sex) VALUES ($1, $2, $3, $4)
		 RETURNING `+patientCols, userId(r), in.Name, in.DateOfBirth, in.Sex))
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (a *App) listPatients(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(),
		`SELECT `+patientCols+` FROM patients WHERE user_id = $1 ORDER BY created_at`, userId(r))
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer rows.Close()
	out := []*Patient{}
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

