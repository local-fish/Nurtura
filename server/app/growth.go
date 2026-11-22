package app

import (
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// RefPoint is one age step of a reference curve. Unlike Reference (normals.go),
// which is the single row nearest the child's current age, these are returned as
// a full series so the growth chart can draw the percentile bands.
type RefPoint struct {
	AgeMonths int `json:"age_months"`
	P3 float64 `json:"p3"`
	P50 float64 `json:"p50"`
	P97 float64 `json:"p97"`
}

// GrowthPoint is one of the patient's own measurements, carrying the age it was
// taken at so the client can plot it against the reference curve's x axis
// without needing the date of birth.
type GrowthPoint struct {
	Id string `json:"id"`
	Value float64 `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	AgeMonths int `json:"age_months"`
}

// growth returns the reference curve for the patient's sex plus the patient's own
// measurements, for a single metric. Defaults to weight, the chart's first tab.
func (a *App) growth(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "weight"
	}
	var unit string
	err := a.db.QueryRow(ctx, `SELECT unit FROM metrics WHERE code = $1`, metric).Scan(&unit)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusBadRequest, "unknown metric")
		return
	}
	if err != nil {
		serverErr(w, r, err)
		return
	}

	refRows, err := a.db.Query(ctx, `
		SELECT age_months, p3, p50, p97 FROM growth_reference
		WHERE sex = $1 AND metric = $2
		ORDER BY age_months`, p.Sex, metric)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer refRows.Close()
	reference := []RefPoint{}
	for refRows.Next() {
		var pt RefPoint
		if err := refRows.Scan(&pt.AgeMonths, &pt.P3, &pt.P50, &pt.P97); err != nil {
			serverErr(w, r, err)
			return
		}
		reference = append(reference, pt)
	}
	if err := refRows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}

	// Ascending, unlike listRecords: the chart plots left-to-right over time.
	recRows, err := a.db.Query(ctx, `
		SELECT id, value, measured_at FROM records
		WHERE patient_id = $1 AND metric = $2
		ORDER BY measured_at`, p.Id, metric)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	defer recRows.Close()
	records := []GrowthPoint{}
	for recRows.Next() {
		var pt GrowthPoint
		if err := recRows.Scan(&pt.Id, &pt.Value, &pt.MeasuredAt); err != nil {
			serverErr(w, r, err)
			return
		}
		pt.AgeMonths = ageMonths(p.DateOfBirth, pt.MeasuredAt)
		records = append(records, pt)
	}
	if err := recRows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"patient_id": p.Id, "sex": p.Sex, "age_months": ageMonths(p.DateOfBirth, time.Now()),
		"metric": metric, "unit": unit,
		"reference": reference, "records": records,
	})
}
