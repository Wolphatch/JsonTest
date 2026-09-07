package jsontest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var envRE = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Tokenize implements the useful quoting and escaping subset used by curl without a shell.
func Tokenize(s string) ([]string, error) {
	if strings.Contains(s, "$(") || strings.ContainsRune(s, '`') {
		return nil, fmt.Errorf("forbidden command substitution")
	}
	var out []string
	var b strings.Builder
	quote := byte(0)
	escaped := false
	active := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			b.WriteByte(ch)
			escaped = false
			active = true
			continue
		}
		if ch == '\\' && quote != '\'' {
			escaped = true
			active = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			} else {
				b.WriteByte(ch)
			}
			active = true
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			active = true
			continue
		}
		if strings.ContainsRune("|;&<>", rune(ch)) {
			return nil, fmt.Errorf("forbidden shell syntax")
		}
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if active {
				out = append(out, b.String())
				b.Reset()
				active = false
			}
			continue
		}
		b.WriteByte(ch)
		active = true
	}
	if escaped || quote != 0 {
		return nil, fmt.Errorf("unterminated escape or quote")
	}
	if active {
		out = append(out, b.String())
	}
	if len(out) == 0 || (out[0] != "curl" && out[0] != "curl.exe") {
		return nil, fmt.Errorf("command must start with curl")
	}
	return out, nil
}

func expand(arg string) (string, error) {
	missing := ""
	v := envRE.ReplaceAllStringFunc(arg, func(x string) string {
		k := envRE.FindStringSubmatch(x)[1]
		val, ok := os.LookupEnv(k)
		if !ok {
			missing = k
		}
		return val
	})
	if missing != "" {
		return "", fmt.Errorf("environment variable %s is not set", missing)
	}
	return v, nil
}

type response struct {
	status int
	body   []byte
}

func execute(raw string, timeout time.Duration) (response, error) {
	args, err := Tokenize(raw)
	if err != nil {
		return response{}, err
	}
	for i := range args {
		args[i], err = expand(args[i])
		if err != nil {
			return response{}, err
		}
	}
	executable := args[0]
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	const marker = "\n__JSON_TEST_HTTP_STATUS__="
	args = append(args[1:], "--silent", "--show-error", "--write-out", marker+"%{http_code}")
	cmd := exec.CommandContext(ctx, executable, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return response{}, fmt.Errorf("request timed out")
		}
		// Curl diagnostics can repeat request URLs or header values. Deliberately do
		// not copy stderr into reports, where credentials could be disclosed.
		return response{}, fmt.Errorf("curl failed with exit code %d", cmd.ProcessState.ExitCode())
	}
	i := bytes.LastIndex(stdout.Bytes(), []byte(marker))
	if i < 0 {
		return response{}, fmt.Errorf("curl did not return an HTTP status")
	}
	status, err := strconv.Atoi(strings.TrimSpace(stdout.String()[i+len(marker):]))
	if err != nil {
		return response{}, fmt.Errorf("invalid HTTP status")
	}
	return response{status: status, body: append([]byte(nil), stdout.Bytes()[:i]...)}, nil
}
