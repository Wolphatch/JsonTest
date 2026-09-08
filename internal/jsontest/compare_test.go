package jsontest

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompareKindsAndPaths(t *testing.T) {
	a := []byte(`{"user":{"id":1,"secret":"a"},"items":[{"price":2},{"price":3}],"gone":true}`)
	b := []byte(`{"user":{"id":"1","secret":"b"},"items":[{"price":9}],"new":true}`)
	d, err := compareJSON(a, b, []string{"user.id", "items[*].price", "gone", "new"}, []string{"user.secret"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"missing_field", "value_mismatch", "array_length_mismatch", "type_mismatch", "missing_field"}
	if len(d) != len(want) {
		t.Fatalf("got %#v", d)
	}
	for i := range want {
		if d[i].Kind != want[i] {
			t.Errorf("difference %d kind %s", i, d[i].Kind)
		}
	}
}

func TestExcludeWins(t *testing.T) {
	d, e := compareJSON([]byte(`{"a":{"b":1}}`), []byte(`{"a":{"b":2}}`), []string{"a.b"}, []string{"a"})
	if e != nil || len(d) != 0 {
		t.Fatalf("%v %#v", e, d)
	}
}

func TestCompareListOrder(t *testing.T) {
	a := []byte(`{"items":[{"id":1},{"id":2},{"id":2}]}`)
	b := []byte(`{"items":[{"id":2},{"id":1},{"id":2}]}`)

	ordered, err := compareJSON(a, b, nil, nil)
	if err != nil || len(ordered) == 0 {
		t.Fatalf("ordered comparison: err=%v differences=%#v", err, ordered)
	}
	unordered, err := compareJSONWithListOrder(a, b, nil, nil, false)
	if err != nil || len(unordered) != 0 {
		t.Fatalf("unordered comparison: err=%v differences=%#v", err, unordered)
	}
}

func TestCompareListOrderStillReportsChanges(t *testing.T) {
	a := []byte(`{"items":[{"id":1,"name":"one"},{"id":2,"name":"two"}]}`)
	b := []byte(`{"items":[{"id":2,"name":"changed"},{"id":1,"name":"one"}]}`)

	d, err := compareJSONWithListOrder(a, b, nil, nil, false)
	if err != nil || len(d) != 1 || d[0].Path != "items[1].name" || d[0].Kind != "value_mismatch" {
		t.Fatalf("err=%v differences=%#v", err, d)
	}
}

func TestWriteTextTable(t *testing.T) {
	r := Report{Cases: []CaseResult{{
		Name: "different response",
		Differences: []Difference{
			{Path: "user.name", Kind: "value_mismatch", Baseline: "old", Candidate: "new"},
			{Path: "items", Kind: "array_length_mismatch", Baseline: 1, Candidate: 2},
		},
	}}}
	var out bytes.Buffer
	WriteText(&out, r)
	for _, want := range []string{"PATH", "KIND", "BASELINE", "CANDIDATE", "user.name", `"old"`, "array_length_mismatch"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, out.String())
		}
	}
	if strings.Count(out.String(), "-------------------------------------------------------------------------------") != 2 {
		t.Errorf("difference table is not separated:\n%s", out.String())
	}
}
