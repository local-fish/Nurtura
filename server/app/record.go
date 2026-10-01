package app

import (
	"context"
	"fmt"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Record struct {
	Id string `json:"id"`
	PatientId string `json:"patient_id"`
	Metric string `json:"metric"`
	Value float64 `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const recordCols = `r.id, r.patient_id, r.metric, r.value, r.measured_at, r.created_at, r.updated_at`

func scanRecord(row pgx.Row) (*Record, error) {
	var rec Record
	err := row.Scan(&rec.Id, &rec.PatientId, &rec.Metric, &rec.Value,
		&rec.MeasuredAt, &rec.CreatedAt, &rec.UpdatedAt)
	return &rec, err
}

func (a *App) checkValue(ctx context.Context, metric string, v float64) (string, error) {
	var lo, hi float64
	err := a.db.QueryRow(ctx,
		`SELECT min_value, max_value FROM metrics WHERE code = $1`, metric).Scan(&lo, &hi)
	if errors.Is(err, pgx.ErrNoRows) {
		return "unknown metric", nil
	}
	if err != nil {
		return "", err
	}
	if v < lo || v > hi {
		return fmt.Sprintf("%s must be between %g and %g", metric, lo, hi), nil
	}
	return "", nil
}

func validMeasuredAt(t *time.Time) bool {
	return t == nil || !t.After(time.Now().Add(time.Hour))
}

func (a *App) addRecord(w http.ResponseWriter, r *http.Request) {
	pid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "patient not found")
		return
	}
	var in struct {
		Metric string `json:"metric"`
		Value *float64 `json:"value"`
		MeasuredAt *time.Time `json:"measured_at"`
	}
	if err := decode(w, r, &in, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.Value == nil {
		writeErr(w, http.StatusBadRequest, "value is required")
		return
	}
	msg, err := a.checkValue(r.Context(), in.Metric, *in.Value)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	if msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	if !validMeasuredAt(in.MeasuredAt) {
		writeErr(w, http.StatusBadRequest, "measured_at cannot be in the future")
		return
	}
	limit, offset, msg := pagination(r)
	if msg != "" {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	rec, err := scanRecord(a.db.QueryRow(r.Context(), `
		INSERT INTO records AS r (patient_id, metric, value, measured_at)
		SELECT p.id, $2, $3::numeric, COALESCE($4::timestamptz, now())
		FROM patients p WHERE p.id = $1 AND p.user_id = $5
		ORDER BY l.created_at DESC, l.id
		LIMIT $3 OFFSET $4
		RETURNING `+recordCols,
		pid, in.Metric, *in.Value, in.MeasuredAt, userId(r), limit, offset))
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "patient not found")
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // FK: metric not in metrics table
		writeErr(w, http.StatusBadRequest, "unknown metric")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

func (a *App) editRecord(w http.ResponseWriter, r *http.Request) {
	rid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "record not found")
		return
	}
	var in struct {
		Value *float64 `json:"value"`
		MeasuredAt *time.Time `json:"measured_at"`
	}
	if err := decode(w, r, &in, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validMeasuredAt(in.MeasuredAt) {
		writeErr(w, http.StatusBadRequest, "measured_at cannot be in the future")
		return
	}
	if in.Value != nil {
		var metric string
		err := a.db.QueryRow(r.Context(), `
			SELECT r.metric FROM records r JOIN patients p ON p.id = r.patient_id
			WHERE r.id = $1 AND p.user_id = $2`, rid, userId(r)).Scan(&metric)
		if errors.Is(err, pgx.ErrNoRows) {
			writeErr(w, http.StatusNotFound, "record not found")
			return
		}
		if err != nil {
			serverErr(w, r, err)
			return
		}
		msg, err := a.checkValue(r.Context(), metric, *in.Value)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		if msg != "" {
			writeErr(w, http.StatusBadRequest, msg)
			return
		}
	}
	rec, err := scanRecord(a.db.QueryRow(r.Context(), `
		UPDATE records r SET
			value       = COALESCE($1::numeric, r.value),
			measured_at = COALESCE($2::timestamptz, r.measured_at),
			updated_at  = now()
		FROM patients p
		WHERE r.id = $3 AND r.patient_id = p.id AND p.user_id = $4
		RETURNING `+recordCols,
		in.Value, in.MeasuredAt, rid, userId(r)))
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "record not found")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (a *App) getRecord(w http.ResponseWriter, r *http.Request) {
	rid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "record not found")
		return
	}
	rec, err := scanRecord(a.db.QueryRow(r.Context(), `
		SELECT `+recordCols+` FROM records r JOIN patients p ON p.id = r.patient_id
		WHERE r.id = $1 AND p.user_id = $2`, rid, userId(r)))
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "record not found")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (a *App) listRecords(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	rows, err := a.db.Query(r.Context(), `
		SELECT `+recordCols+` FROM records r JOIN patients p ON p.id = r.patient_id
		WHERE r.patient_id = $1 AND p.user_id = $2 AND ($3 = '' OR r.metric = $3)
		ORDER BY r.measured_at DESC`, p.Id, userId(r), r.URL.Query().Get("metric"))
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer rows.Close()
	out := []*Record{}
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
