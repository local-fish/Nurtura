package app

import (
	"net/http"
	"strings"
	"time"
)

type Log struct {
	Id string `json:"id"`
	PatientId string `json:"patient_id"`
	Note string `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *App) addLog(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	var in struct {
		Note string `json:"note"`
	}
	if err := decode(w, r, &in, false); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	if in.Note == "" || len(in.Note) > 10000 {
		writeErr(w, http.StatusBadRequest, "note is required (max 10000 chars)")
		return
	}
	var l Log
	err := a.db.QueryRow(r.Context(), `
		INSERT INTO logs (patient_id, note)
		SELECT p.id, $2 FROM patients p WHERE p.id = $1 AND p.user_id = $3
		RETURNING id, patient_id, note, created_at`,
		p.Id, in.Note, userId(r)).Scan(&l.Id, &l.PatientId, &l.Note, &l.CreatedAt)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

func (a *App) listLogs(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	rows, err := a.db.Query(r.Context(), `
		SELECT l.id, l.patient_id, l.note, l.created_at
		FROM logs l JOIN patients p ON p.id = l.patient_id
		WHERE l.patient_id = $1 AND p.user_id = $2
		ORDER BY l.created_at DESC`, p.Id, userId(r))
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer rows.Close()
	out := []Log{}
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.Id, &l.PatientId, &l.Note, &l.CreatedAt); err != nil {
			serverErr(w, r, err)
			return
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

