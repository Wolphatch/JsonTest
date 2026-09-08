package jsontest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "m.yaml")
	os.WriteFile(p, []byte("tests:\n  - name: x\n    baseline: curl http://a\n    candidate:\n      curl: curl http://b\n"), 0600)
	m, e := Load(p)
	if e != nil || len(m.Tests) != 1 {
		t.Fatalf("%v", e)
	}
}

func TestLoadCompareListOrder(t *testing.T) {
	p := filepath.Join(t.TempDir(), "m.yaml")
	err := os.WriteFile(p, []byte("compareListOrder: false\ntests:\n  - name: x\n    baseline: curl http://a\n    candidate: curl http://b\n    compareListOrder: true\n"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if m.CompareListOrder == nil || *m.CompareListOrder || m.Tests[0].CompareListOrder == nil || !*m.Tests[0].CompareListOrder {
		t.Fatalf("list-order settings were not decoded: %#v", m)
	}
}
