package jsontest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Difference struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Baseline  any    `json:"baseline,omitempty"`
	Candidate any    `json:"candidate,omitempty"`
}

var pathPart = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func parsePath(s string) ([]string, error) {
	if s == "" {
		return nil, fmt.Errorf("empty path")
	}
	var parts []string
	for _, dot := range strings.Split(s, ".") {
		if dot == "" {
			return nil, fmt.Errorf("empty segment")
		}
		for len(dot) > 0 {
			i := strings.IndexByte(dot, '[')
			if i < 0 {
				if !pathPart.MatchString(dot) {
					return nil, fmt.Errorf("invalid segment")
				}
				parts = append(parts, dot)
				break
			}
			if i > 0 {
				p := dot[:i]
				if !pathPart.MatchString(p) {
					return nil, fmt.Errorf("invalid segment")
				}
				parts = append(parts, p)
			}
			j := strings.IndexByte(dot[i:], ']')
			if j < 0 {
				return nil, fmt.Errorf("unclosed bracket")
			}
			token := dot[i+1 : i+j]
			if token != "*" {
				if _, e := strconv.Atoi(token); e != nil {
					return nil, fmt.Errorf("array selector must be * or an index")
				}
			}
			parts = append(parts, token)
			dot = dot[i+j+1:]
		}
	}
	return parts, nil
}
func matchPrefix(pattern, path []string) bool {
	if len(pattern) > len(path) {
		return false
	}
	for i := range pattern {
		if pattern[i] != "*" && pattern[i] != path[i] {
			return false
		}
	}
	return true
}
func pathPrefix(a, b []string) bool {
	if len(a) > len(b) {
		return false
	}
	for i := range a {
		if a[i] != "*" && b[i] != "*" && a[i] != b[i] {
			return false
		}
	}
	return true
}
func relevant(patterns [][]string, path []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if pathPrefix(p, path) || pathPrefix(path, p) {
			return true
		}
	}
	return false
}
func excluded(patterns [][]string, path []string) bool {
	for _, p := range patterns {
		if matchPrefix(p, path) {
			return true
		}
	}
	return false
}
func display(p []string) string {
	var b strings.Builder
	for _, x := range p {
		if _, e := strconv.Atoi(x); e == nil {
			fmt.Fprintf(&b, "[%s]", x)
		} else {
			if b.Len() > 0 {
				b.WriteByte('.')
			}
			b.WriteString(x)
		}
	}
	if b.Len() == 0 {
		return "$"
	}
	return b.String()
}

func compareJSON(a, b []byte, includes, excludes []string) ([]Difference, error) {
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		return nil, fmt.Errorf("baseline response is not JSON: %w", err)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		return nil, fmt.Errorf("candidate response is not JSON: %w", err)
	}
	var in, ex [][]string
	for _, s := range includes {
		p, _ := parsePath(s)
		in = append(in, p)
	}
	for _, s := range excludes {
		p, _ := parsePath(s)
		ex = append(ex, p)
	}
	var ds []Difference
	var walk func(any, any, []string)
	walk = func(a, b any, p []string) {
		if excluded(ex, p) || !relevant(in, p) {
			return
		}
		ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
		if ta != tb {
			ds = append(ds, Difference{display(p), "type_mismatch", typeName(a), typeName(b)})
			return
		}
		switch av := a.(type) {
		case map[string]any:
			bv := b.(map[string]any)
			keys := make([]string, 0, len(av))
			for k := range av {
				keys = append(keys, k)
			}
			sortStrings(keys)
			for _, k := range keys {
				np := appendPath(p, k)
				v, ok := bv[k]
				if !ok {
					if relevant(in, np) && !excluded(ex, np) {
						ds = append(ds, Difference{display(np), "missing_field", av[k], nil})
					}
					continue
				}
				walk(av[k], v, np)
			}
			keys = keys[:0]
			for k := range bv {
				if _, ok := av[k]; !ok {
					keys = append(keys, k)
				}
			}
			sortStrings(keys)
			for _, k := range keys {
				np := appendPath(p, k)
				if relevant(in, np) && !excluded(ex, np) {
					ds = append(ds, Difference{display(np), "missing_field", nil, bv[k]})
				}
			}
		case []any:
			bv := b.([]any)
			n := len(av)
			if len(bv) < n {
				n = len(bv)
			}
			for i := 0; i < n; i++ {
				walk(av[i], bv[i], appendPath(p, strconv.Itoa(i)))
			}
			if len(av) != len(bv) {
				ds = append(ds, Difference{display(p), "array_length_mismatch", len(av), len(bv)})
			}
		default:
			if !reflect.DeepEqual(a, b) {
				ds = append(ds, Difference{display(p), "value_mismatch", a, b})
			}
		}
	}
	walk(x, y, nil)
	return ds, nil
}
func appendPath(p []string, s string) []string {
	n := make([]string, len(p)+1)
	copy(n, p)
	n[len(p)] = s
	return n
}
func typeName(v any) any {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case float64:
		return "number"
	}
	return reflect.TypeOf(v).Name()
}
func sortStrings(v []string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}
