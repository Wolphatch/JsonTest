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
