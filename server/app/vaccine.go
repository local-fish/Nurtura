package app

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type VaccineStatus struct {
	VaccineId int `json:"vaccine_id"`
	Name string `json:"name"`
	DoseNumber int `json:"dose_number"`
	RecommendedAgeMonths int `json:"recommended_age_months"`
	Completed bool `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
	Status string `json:"status"` // completed | overdue | upcoming
}

func (a *App) completeVaccine(w http.ResponseWriter, r *http.Request) {
	pid, ok := pathUUID(r, "id")
	if !ok {
		writeErr(w, http.StatusNotFound, "patient not found")
		return
	}
	vid, err := strconv.Atoi(r.PathValue("vaccine_id"))
	if err != nil || vid <= 0 {
		writeErr(w, http.StatusNotFound, "vaccine not found")
		return
	}
	var in struct {
		CompletedAt *time.Time `json:"completed_at"`
	}
	if err := decode(w, r, &in, true); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.CompletedAt != nil && in.CompletedAt.After(time.Now().Add(time.Hour)) {
		writeErr(w, http.StatusBadRequest, "completed_at cannot be in the future")
		return
	}
	var completedAt time.Time
	err = a.db.QueryRow(r.Context(), `
		INSERT INTO patient_vaccines (patient_id, vaccine_id, completed_at)
		SELECT p.id, v.id, COALESCE($3::timestamptz, now())
		FROM patients p, vaccines v
		WHERE p.id = $1 AND p.user_id = $4 AND v.id = $2
		ON CONFLICT (patient_id, vaccine_id) DO UPDATE SET completed_at = EXCLUDED.completed_at
		RETURNING completed_at`,
		pid, vid, in.CompletedAt, userId(r)).Scan(&completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "patient or vaccine not found")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"patient_id": pid, "vaccine_id": vid, "completed": true, "completed_at": completedAt,
	})
}

func (a *App) listVaccines(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	age := ageMonths(p.DateOfBirth, time.Now())
	rows, err := a.db.Query(r.Context(), `
		SELECT v.id, v.name, v.dose_number, v.recommended_age_months, pv.completed_at
		FROM vaccines v
		LEFT JOIN patient_vaccines pv ON pv.vaccine_id = v.id AND pv.patient_id = $1
		ORDER BY v.recommended_age_months, v.name, v.dose_number`, p.Id)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer rows.Close()
	out := []VaccineStatus{}
	for rows.Next() {
		var v VaccineStatus
		if err := rows.Scan(&v.VaccineId, &v.Name, &v.DoseNumber, &v.RecommendedAgeMonths, &v.CompletedAt); err != nil {
			serverErr(w, r, err)
			return
		}
		v.Completed = v.CompletedAt != nil
		switch {
		case v.Completed:
			v.Status = "completed"
		case age >= v.RecommendedAgeMonths:
			v.Status = "overdue"
		default:
			v.Status = "upcoming"
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"patient_id": p.Id, "age_months": age, "vaccines": out})
}

