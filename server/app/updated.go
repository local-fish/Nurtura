package app

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func (a *App) patientActivity(w http.ResponseWriter, r *http.Request) {
	pid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "patient not found")
		return
	}
	var at *time.Time
	err := a.db.QueryRow(r.Context(), `
		SELECT pa.last_activity_at
		FROM patients p
		LEFT JOIN patient_activity pa ON pa.patient_id = p.id
		WHERE p.id = $1 AND p.user_id = $2`, pid, userId(r)).Scan(&at)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "patient not found")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"patient_id": pid, "last_activity_at": at})
}
