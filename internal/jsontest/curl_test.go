package jsontest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenize(t *testing.T) {
	v, e := Tokenize(`curl -H "Authorization: Bearer ${TOKEN}" 'https://example.test/a b'`)
	if e != nil || len(v) != 4 {
		t.Fatalf("%v %#v", e, v)
	}
	for _, s := range []string{`curl x | cat`, `curl x > out`, `curl $(evil)`, `curl "$(evil)"`, "curl '`evil`'", `curl x && evil`, `wget x`} {
		if _, e := Tokenize(s); e == nil {
			t.Errorf("accepted %q", s)
		}
	}
}

func TestExecuteRealCurl(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer s.Close()
	r, e := execute("curl "+s.URL, 0)
	if e != nil {
		t.Fatal(e)
	}
	if r.status != 201 || string(r.body) != `{"ok":true}` {
		t.Fatalf("%+v %q", r, string(r.body))
	}
}
