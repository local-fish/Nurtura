package tests

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

var base = func() string {
	if v := os.Getenv("BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:3000"
}()

const pw = "correcthorse-1"

type resp struct {
	Status int
	Raw    string
	JSON   map[string]any
	List   []any
}

func call(t *testing.T, method, path, token string, body any) resp {
	t.Helper()
	var rd io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rd = strings.NewReader(b)
	default:
		buf, _ := json.Marshal(b)
		rd = bytes.NewReader(buf)
	}
	req, _ := http.NewRequest(method, base+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v (is the server running at %s?)", method, path, err, base)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	r := resp{Status: res.StatusCode, Raw: string(raw)}
	_ = json.Unmarshal(raw, &r.JSON)
	_ = json.Unmarshal(raw, &r.List)
	return r
}

func want(t *testing.T, r resp, status int) {
	t.Helper()
	if r.Status != status {
		t.Fatalf("want HTTP %d, got %d: %s", status, r.Status, r.Raw)
	}
}

func str(m map[string]any, k string) string { s, _ := m[k].(string); return s }

func randEmail() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "t-" + hex.EncodeToString(b) + "@example.com"
}

// login retries once after a rate-limit (429), since every test logs in from
// the same IP and the API limits /auth/login to ~10/min.
func login(t *testing.T, email string) resp {
	t.Helper()
	body := map[string]string{"email": email, "password": pw}
	r := call(t, "POST", "/auth/login", "", body)
	if r.Status == 429 {
		t.Log("login rate-limited; waiting 61s for the window to reset")
		time.Sleep(61 * time.Second)
		r = call(t, "POST", "/auth/login", "", body)
	}
	return r
}

// newUser registers + logs in, returning email and token.
func newUser(t *testing.T) (string, string) {
	t.Helper()
	email := randEmail()
	want(t, call(t, "POST", "/auth/register", "", map[string]string{"email": email, "password": pw}), 201)
	r := login(t, email)
	want(t, r, 200)
	tok := str(r.JSON, "token")
	if tok == "" {
		t.Fatal("login returned no token")
	}
	return email, tok
}

// sharedToken is one user reused by tests that don't need a fresh account,
// which keeps the number of logins under the rate limit.
var shared struct {
	once sync.Once
	tok  string
}

func sharedToken(t *testing.T) string {
	t.Helper()
	shared.once.Do(func() { _, shared.tok = newUser(t) })
	if shared.tok == "" {
		t.Fatal("shared user setup failed")
	}
	return shared.tok
}

func newPatient(t *testing.T, tok string) string {
	t.Helper()
	r := call(t, "POST", "/patients", tok, map[string]string{"name": "Kid", "date_of_birth": "2024-03-15", "sex": "F"})
	want(t, r, 201)
	return str(r.JSON, "id")
}

func TestAuth(t *testing.T) {
	email := randEmail()

	t.Run("register ok", func(t *testing.T) {
		want(t, call(t, "POST", "/auth/register", "", map[string]string{"email": email, "password": pw}), 201)
	})
	t.Run("duplicate email 409 (case-insensitive)", func(t *testing.T) {
		want(t, call(t, "POST", "/auth/register", "", map[string]string{"email": strings.ToUpper(email), "password": pw}), 409)
	})
	t.Run("bad email 400", func(t *testing.T) {
		want(t, call(t, "POST", "/auth/register", "", map[string]string{"email": "nope", "password": pw}), 400)
	})
	t.Run("short password 400", func(t *testing.T) {
		want(t, call(t, "POST", "/auth/register", "", map[string]string{"email": randEmail(), "password": "short"}), 400)
	})
	t.Run("malformed JSON 400", func(t *testing.T) {
		want(t, call(t, "POST", "/auth/register", "", "{not json"), 400)
	})
	t.Run("login ok returns token", func(t *testing.T) {
		r := login(t, email)
		want(t, r, 200)
		if str(r.JSON, "token") == "" {
			t.Fatal("no token")
		}
	})
	t.Run("wrong password and unknown email look identical", func(t *testing.T) {
		a := call(t, "POST", "/auth/login", "", map[string]string{"email": email, "password": "wrong-password-1"})
		b := call(t, "POST", "/auth/login", "", map[string]string{"email": randEmail(), "password": "wrong-password-1"})
		want(t, a, 401)
		want(t, b, 401)
		if a.Raw != b.Raw {
			t.Fatalf("responses differ (leaks account existence):\n%s\n%s", a.Raw, b.Raw)
		}
	})
	t.Run("protected routes reject missing/garbage tokens", func(t *testing.T) {
		want(t, call(t, "GET", "/patients", "", nil), 401)
		want(t, call(t, "GET", "/patients", "garbage.token.here", nil), 401)
	})
	t.Run("alg=none token rejected", func(t *testing.T) {
		// header {"alg":"none","typ":"JWT"} . {"sub":"00000000-0000-0000-0000-000000000000"} .
		tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAifQ."
		want(t, call(t, "GET", "/patients", tok, nil), 401)
	})
}

func TestPatients(t *testing.T) {
	tok := sharedToken(t)

	t.Run("validation", func(t *testing.T) {
		bad := []map[string]string{
			{"name": "", "date_of_birth": "2024-01-01", "sex": "F"},
			{"name": "K", "date_of_birth": "not-a-date", "sex": "F"},
			{"name": "K", "date_of_birth": "2999-01-01", "sex": "F"},
			{"name": "K", "date_of_birth": "2024-01-01", "sex": "X"},
		}
		for i, b := range bad {
			if r := call(t, "POST", "/patients", tok, b); r.Status != 400 {
				t.Errorf("case %d: want 400, got %d: %s", i, r.Status, r.Raw)
			}
		}
	})
	t.Run("create + list", func(t *testing.T) {
		id := newPatient(t, tok)
		r := call(t, "GET", "/patients", tok, nil)
		want(t, r, 200)
		found := false
		for _, p := range r.List {
			if m, _ := p.(map[string]any); str(m, "id") == id {
				found = true
			}
		}
		if !found {
			t.Fatalf("created patient %s not in list: %s", id, r.Raw)
		}
	})
	t.Run("list is scoped to caller", func(t *testing.T) {
		_, other := newUser(t)
		r := call(t, "GET", "/patients", other, nil)
		want(t, r, 200)
		if len(r.List) != 0 {
			t.Fatalf("new user sees %d patients, want 0", len(r.List))
		}
	})
}

func TestRecords(t *testing.T) {
	tok := sharedToken(t)
	pid := newPatient(t, tok)
	var rid string

	t.Run("add height", func(t *testing.T) {
		r := call(t, "POST", "/patients/"+pid+"/records", tok, map[string]any{"metric": "height", "value": 84.2})
		want(t, r, 201)
		rid = str(r.JSON, "id")
		if str(r.JSON, "metric") != "height" || r.JSON["value"].(float64) != 84.2 {
			t.Fatalf("unexpected body: %s", r.Raw)
		}
	})
	t.Run("add weight with explicit measured_at", func(t *testing.T) {
		r := call(t, "POST", "/patients/"+pid+"/records", tok,
			map[string]any{"metric": "weight", "value": 11.1, "measured_at": "2026-01-01T10:00:00Z"})
		want(t, r, 201)
	})
	t.Run("rejects bad input", func(t *testing.T) {
		cases := map[string]any{
			"missing value":  map[string]any{"metric": "height"},
			"unknown metric": map[string]any{"metric": "nonsense", "value": 5},
			"too high":       map[string]any{"metric": "height", "value": 9999},
			"too low":        map[string]any{"metric": "height", "value": 0.001},
			"negative":       map[string]any{"metric": "weight", "value": -3},
			"future date":    map[string]any{"metric": "height", "value": 80, "measured_at": "2999-01-01T00:00:00Z"},
			"unknown field":  map[string]any{"metric": "height", "value": 80, "bogus": 1},
		}
		for name, body := range cases {
			if r := call(t, "POST", "/patients/"+pid+"/records", tok, body); r.Status != 400 {
				t.Errorf("%s: want 400, got %d: %s", name, r.Status, r.Raw)
			}
		}
	})
	t.Run("list newest first", func(t *testing.T) {
		r := call(t, "GET", "/patients/"+pid+"/records", tok, nil)
		want(t, r, 200)
		if len(r.List) != 2 {
			t.Fatalf("want 2 records, got %d", len(r.List))
		}
		first, _ := r.List[0].(map[string]any)
		if str(first, "metric") != "height" { // height=now, weight=Jan 2026
			t.Fatalf("expected newest (height) first, got %s", r.Raw)
		}
	})
	t.Run("filter by metric", func(t *testing.T) {
		r := call(t, "GET", "/patients/"+pid+"/records?metric=weight", tok, nil)
		want(t, r, 200)
		if len(r.List) != 1 {
			t.Fatalf("want 1 weight record, got %d", len(r.List))
		}
	})
	t.Run("get one", func(t *testing.T) {
		r := call(t, "GET", "/records/"+rid, tok, nil)
		want(t, r, 200)
		if str(r.JSON, "id") != rid {
			t.Fatalf("wrong record: %s", r.Raw)
		}
	})
	t.Run("edit value", func(t *testing.T) {
		r := call(t, "PATCH", "/records/"+rid, tok, map[string]any{"value": 85.0})
		want(t, r, 200)
		if r.JSON["value"].(float64) != 85.0 {
			t.Fatalf("value not updated: %s", r.Raw)
		}
	})
	t.Run("edit rejects out-of-range and metric change", func(t *testing.T) {
		want(t, call(t, "PATCH", "/records/"+rid, tok, map[string]any{"value": 9999}), 400)
		want(t, call(t, "PATCH", "/records/"+rid, tok, map[string]any{"metric": "weight"}), 400)
	})
	t.Run("bad ids 404", func(t *testing.T) {
		want(t, call(t, "GET", "/records/not-a-uuid", tok, nil), 404)
		want(t, call(t, "GET", "/records/00000000-0000-0000-0000-000000000000", tok, nil), 404)
		want(t, call(t, "POST", "/patients/not-a-uuid/records", tok, map[string]any{"metric": "height", "value": 80}), 404)
	})
}

func TestLogs(t *testing.T) {
	tok := sharedToken(t)
	pid := newPatient(t, tok)

	want(t, call(t, "POST", "/patients/"+pid+"/logs", tok, map[string]string{"note": ""}), 400)
	want(t, call(t, "POST", "/patients/"+pid+"/logs", tok, map[string]string{"note": "   "}), 400)
	want(t, call(t, "POST", "/patients/"+pid+"/logs", tok, map[string]string{"note": "first"}), 201)
	want(t, call(t, "POST", "/patients/"+pid+"/logs", tok, map[string]string{"note": "second"}), 201)

	r := call(t, "GET", "/patients/"+pid+"/logs", tok, nil)
	want(t, r, 200)
	if len(r.List) != 2 {
		t.Fatalf("want 2 logs, got %d", len(r.List))
	}
	if first, _ := r.List[0].(map[string]any); str(first, "note") != "second" {
		t.Fatalf("expected newest first: %s", r.Raw)
	}
}

func TestVaccines(t *testing.T) {
	tok := sharedToken(t)
	pid := newPatient(t, tok)

	list := func() []map[string]any {
		r := call(t, "GET", "/patients/"+pid+"/vaccines", tok, nil)
		want(t, r, 200)
		arr, _ := r.JSON["vaccines"].([]any)
		out := make([]map[string]any, 0, len(arr))
		for _, v := range arr {
			m, _ := v.(map[string]any)
			out = append(out, m)
		}
		return out
	}

	before := list()
	if len(before) == 0 {
		t.Fatal("no vaccines returned; did you run the seed file?")
	}
	for _, v := range before {
		if v["completed"] != false {
			t.Fatalf("fresh patient has a completed vaccine: %v", v)
		}
	}
	id := int(before[0]["vaccine_id"].(float64))

	t.Run("mark complete", func(t *testing.T) {
		want(t, call(t, "POST", fmt.Sprintf("/patients/%s/vaccines/%d/complete", pid, id), tok, nil), 200)
		done := 0
		for _, v := range list() {
			if int(v["vaccine_id"].(float64)) == id {
				if v["completed"] != true || v["status"] != "completed" {
					t.Fatalf("not marked complete: %v", v)
				}
				done++
			}
		}
		if done != 1 {
			t.Fatalf("vaccine %d appeared %d times", id, done)
		}
	})
	t.Run("idempotent", func(t *testing.T) {
		want(t, call(t, "POST", fmt.Sprintf("/patients/%s/vaccines/%d/complete", pid, id), tok, nil), 200)
	})
	t.Run("explicit completed_at", func(t *testing.T) {
		id2 := int(before[1]["vaccine_id"].(float64))
		want(t, call(t, "POST", fmt.Sprintf("/patients/%s/vaccines/%d/complete", pid, id2), tok,
			map[string]string{"completed_at": "2025-06-01T00:00:00Z"}), 200)
	})
	t.Run("rejects bad input", func(t *testing.T) {
		want(t, call(t, "POST", "/patients/"+pid+"/vaccines/999999/complete", tok, nil), 404)
		want(t, call(t, "POST", "/patients/"+pid+"/vaccines/abc/complete", tok, nil), 404)
		want(t, call(t, "POST", fmt.Sprintf("/patients/%s/vaccines/%d/complete", pid, id), tok,
			map[string]string{"completed_at": "2999-01-01T00:00:00Z"}), 400)
	})
}

func TestNormals(t *testing.T) {
	tok := sharedToken(t)
	pid := newPatient(t, tok)

	t.Run("works with no records", func(t *testing.T) {
		want(t, call(t, "GET", "/patients/"+pid+"/normals", tok, nil), 200)
	})
	t.Run("works with records", func(t *testing.T) {
		want(t, call(t, "POST", "/patients/"+pid+"/records", tok, map[string]any{"metric": "height", "value": 84.2}), 201)
		want(t, call(t, "POST", "/patients/"+pid+"/records", tok, map[string]any{"metric": "weight", "value": 11.4}), 201)
		r := call(t, "GET", "/patients/"+pid+"/normals", tok, nil)
		want(t, r, 200)
		if r.JSON == nil {
			t.Fatalf("expected a JSON object: %s", r.Raw)
		}
	})
}

// The critical one: user B must get 404 (not 403, not data) for everything user A owns.
func TestOwnership(t *testing.T) {
	a := sharedToken(t)
	_, b := newUser(t)
	pid := newPatient(t, a)

	rec := call(t, "POST", "/patients/"+pid+"/records", a, map[string]any{"metric": "height", "value": 84.2})
	want(t, rec, 201)
	rid := str(rec.JSON, "id")
	want(t, call(t, "POST", "/patients/"+pid+"/logs", a, map[string]string{"note": "private"}), 201)

	cases := []struct {
		name, method, path string
		body               any
	}{
		{"view record", "GET", "/records/" + rid, nil},
		{"edit record", "PATCH", "/records/" + rid, map[string]any{"value": 99}},
		{"list records", "GET", "/patients/" + pid + "/records", nil},
		{"add record", "POST", "/patients/" + pid + "/records", map[string]any{"metric": "height", "value": 90}},
		{"view normals", "GET", "/patients/" + pid + "/normals", nil},
		{"list logs", "GET", "/patients/" + pid + "/logs", nil},
		{"add log", "POST", "/patients/" + pid + "/logs", map[string]string{"note": "hax"}},
		{"list vaccines", "GET", "/patients/" + pid + "/vaccines", nil},
		{"complete vaccine", "POST", "/patients/" + pid + "/vaccines/1/complete", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := call(t, c.method, c.path, b, c.body)
			if r.Status != 404 {
				t.Fatalf("want 404, got %d: %s", r.Status, r.Raw)
			}
			if strings.Contains(r.Raw, "private") {
				t.Fatalf("response leaks owner's data: %s", r.Raw)
			}
		})
	}

	t.Run("owner's data is untouched", func(t *testing.T) {
		r := call(t, "GET", "/records/"+rid, a, nil)
		want(t, r, 200)
		if r.JSON["value"].(float64) != 84.2 {
			t.Fatalf("record was modified by another user: %s", r.Raw)
		}
		if l := call(t, "GET", "/patients/"+pid+"/logs", a, nil); len(l.List) != 1 {
			t.Fatalf("owner's logs changed: %s", l.Raw)
		}
	})
}

