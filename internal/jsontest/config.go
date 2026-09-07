package jsontest

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalYAML(n *yaml.Node) error {
	v, err := time.ParseDuration(n.Value)
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

type Request struct {
	Curl string `yaml:"curl"`
}

func (r *Request) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		r.Curl = n.Value
		return nil
	}
	type plain Request
	return n.Decode((*plain)(r))
}

type Manifest struct {
	Version       int      `yaml:"version"`
	Report        string   `yaml:"report"`
	Timeout       Duration `yaml:"timeout"`
	CompareStatus *bool    `yaml:"compareStatus"`
	Tests         []Case   `yaml:"tests"`
}
type Case struct {
	Name          string   `yaml:"name"`
	Baseline      Request  `yaml:"baseline"`
	Candidate     Request  `yaml:"candidate"`
	Include       []string `yaml:"include"`
	Exclude       []string `yaml:"exclude"`
	Timeout       Duration `yaml:"timeout"`
	CompareStatus *bool    `yaml:"compareStatus"`
}

func Load(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer f.Close()
	var m Manifest
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if m.Version != 0 && m.Version != 1 {
		return m, fmt.Errorf("unsupported version %d", m.Version)
	}
	if m.Report == "" {
		m.Report = "text"
	}
	if m.Report != "text" && m.Report != "json" {
		return m, fmt.Errorf("report must be text or json")
	}
	if len(m.Tests) == 0 {
		return m, fmt.Errorf("tests must not be empty")
	}
	seen := map[string]bool{}
	for i := range m.Tests {
		c := &m.Tests[i]
		if strings.TrimSpace(c.Name) == "" {
			return m, fmt.Errorf("tests[%d].name is required", i)
		}
		if seen[c.Name] {
			return m, fmt.Errorf("duplicate test name %q", c.Name)
		}
		seen[c.Name] = true
		if c.Baseline.Curl == "" || c.Candidate.Curl == "" {
			return m, fmt.Errorf("test %q requires baseline and candidate curl", c.Name)
		}
		if c.Timeout.Duration < 0 {
			return m, fmt.Errorf("test %q timeout must be positive", c.Name)
		}
		for _, p := range append(append([]string{}, c.Include...), c.Exclude...) {
			if _, err := parsePath(p); err != nil {
				return m, fmt.Errorf("test %q path %q: %w", c.Name, p, err)
			}
		}
		if _, err := Tokenize(c.Baseline.Curl); err != nil {
			return m, fmt.Errorf("test %q baseline: %w", c.Name, err)
		}
		if _, err := Tokenize(c.Candidate.Curl); err != nil {
			return m, fmt.Errorf("test %q candidate: %w", c.Name, err)
		}
	}
	return m, nil
}
