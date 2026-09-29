package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type Reference struct {
	AgeMonths int `json:"reference_age_months"`
	P3 float64 `json:"p3"`
	P50 float64 `json:"p50"`
	P97 float64 `json:"p97"`
}

type Comparison struct {
	Value float64 `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	AgeAtMeasure int `json:"age_months_at_measurement"`
	Reference *Reference `json:"reference"`
	Band string `json:"band"` // below_p3 | p3_to_p97 | above_p97
}

func (a *App) lookupRef(ctx context.Context, sex, metric string, age int) (*Reference, error) {
	var ref Reference
	err := a.db.QueryRow(ctx, `
		SELECT age_months, p3, p50, p97 FROM growth_reference
		WHERE sex = $1 AND metric = $2
		ORDER BY abs(age_months - $3::int), age_months
		LIMIT 1`, sex, metric, age).Scan(&ref.AgeMonths, &ref.P3, &ref.P50, &ref.P97)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &ref, err
}

func band(v float64, ref *Reference) string {
	switch {
	case v < ref.P3:
		return "below_p3"
	case v > ref.P97:
		return "above_p97"
	default:
		return "p3_to_p97"
	}
}

type latestVal struct {
	value float64
	at    time.Time
}

func (a *App) normals(w http.ResponseWriter, r *http.Request) {
	p, ok := a.ownedPatient(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	age := ageMonths(p.DateOfBirth, time.Now())

	mrows, err := a.db.Query(ctx,
		`SELECT DISTINCT metric FROM growth_reference WHERE sex = $1 ORDER BY metric`, p.Sex)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	var metrics []string
	for mrows.Next() {
		var m string
		if err := mrows.Scan(&m); err != nil {
			mrows.Close()
			serverErr(w, r, err)
			return
		}
		metrics = append(metrics, m)
	}
	mrows.Close()
	if err := mrows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}

	lrows, err := a.db.Query(ctx, `
		SELECT DISTINCT ON (metric) metric, value, measured_at
		FROM records WHERE patient_id = $1
		ORDER BY metric, measured_at DESC`, p.Id)
	if err != nil {
		serverErr(w, r, err)
		return
	}
	latestBy := map[string]latestVal{}
	for lrows.Next() {
		var m string
		var lv latestVal
		if err := lrows.Scan(&m, &lv.value, &lv.at); err != nil {
			lrows.Close()
			serverErr(w, r, err)
			return
		}
		latestBy[m] = lv
	}
	lrows.Close()
	if err := lrows.Err(); err != nil {
		serverErr(w, r, err)
		return
	}

	reference := map[string]*Reference{}
	latest := map[string]*Comparison{}
	for _, metric := range metrics {
		ref, err := a.lookupRef(ctx, p.Sex, metric, age)
		if err != nil {
			serverErr(w, r, err)
			return
		}
		reference[metric] = ref

		lv, has := latestBy[metric]
		if !has {
			latest[metric] = nil
			continue
		}
		c := &Comparison{Value: lv.value, MeasuredAt: lv.at, AgeAtMeasure: ageMonths(p.DateOfBirth, lv.at)}
		if c.Reference, err = a.lookupRef(ctx, p.Sex, metric, c.AgeAtMeasure); err != nil {
			serverErr(w, r, err)
			return
		}
		if c.Reference != nil {
			c.Band = band(c.Value, c.Reference)
		}
		latest[metric] = c
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"patient_id": p.Id, "sex": p.Sex, "age_months": age,
		"reference": reference, "latest": latest,
	})
}
