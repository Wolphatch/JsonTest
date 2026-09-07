package jsontest

import "testing"

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
