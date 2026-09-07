package jsontest

import (
	"encoding/json"
	"fmt"
	"io"
)

type CaseResult struct {
	Name            string       `json:"name"`
	Matched         bool         `json:"matched"`
	BaselineStatus  int          `json:"baselineStatus,omitempty"`
	CandidateStatus int          `json:"candidateStatus,omitempty"`
	Differences     []Difference `json:"differences"`
	Error           string       `json:"error,omitempty"`
}
type Report struct {
	Matched bool         `json:"matched"`
	Cases   []CaseResult `json:"cases"`
}

func Run(m Manifest) Report {
	r := Report{Matched: true, Cases: make([]CaseResult, 0, len(m.Tests))}
	for _, c := range m.Tests {
		cr := CaseResult{Name: c.Name, Matched: true, Differences: []Difference{}}
		t := c.Timeout.Duration
		if t == 0 {
			t = m.Timeout.Duration
		}
		a, e := execute(c.Baseline.Curl, t)
		if e != nil {
			cr.Error = "baseline request: " + e.Error()
		} else {
			cr.BaselineStatus = a.status
			b, e := execute(c.Candidate.Curl, t)
			if e != nil {
				cr.Error = "candidate request: " + e.Error()
			} else {
				cr.CandidateStatus = b.status
				status := true
				if m.CompareStatus != nil {
					status = *m.CompareStatus
				}
				if c.CompareStatus != nil {
					status = *c.CompareStatus
				}
				if status && a.status != b.status {
					cr.Differences = append(cr.Differences, Difference{"$status", "status_mismatch", a.status, b.status})
				}
				ds, e := compareJSON(a.body, b.body, c.Include, c.Exclude)
				if e != nil {
					cr.Error = e.Error()
				} else {
					cr.Differences = append(cr.Differences, ds...)
				}
			}
		}
		cr.Matched = cr.Error == "" && len(cr.Differences) == 0
		if !cr.Matched {
			r.Matched = false
		}
		r.Cases = append(r.Cases, cr)
	}
	return r
}
func WriteJSON(w io.Writer, r Report) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")
	return e.Encode(r)
}
func WriteText(w io.Writer, r Report) {
	for _, c := range r.Cases {
		state := "PASS"
		if !c.Matched {
			state = "FAIL"
		}
		fmt.Fprintf(w, "%s %s\n", state, c.Name)
		if c.Error != "" {
			fmt.Fprintf(w, "  error: %s\n", c.Error)
		}
		for _, d := range c.Differences {
			fmt.Fprintf(w, "  %s: %s (baseline=%v candidate=%v)\n", d.Path, d.Kind, d.Baseline, d.Candidate)
		}
	}
	fmt.Fprintf(w, "\n%d case(s), matched=%t\n", len(r.Cases), r.Matched)
}
