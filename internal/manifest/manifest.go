// Package manifest extracts image references from Dockerfile and
// docker-compose files.
package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Kind is the detected manifest type.
type Kind int

const (
	// Dockerfile is a Dockerfile (FROM lines).
	Dockerfile Kind = iota
	// Compose is a docker-compose.yml / .yaml file (image: keys).
	Compose
)

// Detect guesses the manifest kind from the file name.
func Detect(path string) (Kind, error) {
	base := strings.ToLower(filepath.Base(path))
	switch {
	case strings.HasPrefix(base, "dockerfile") || strings.HasSuffix(base, ".dockerfile"):
		return Dockerfile, nil
	case strings.HasPrefix(base, "compose") || strings.HasPrefix(base, "docker-compose"):
		return Compose, nil
	}
	return 0, fmt.Errorf("cannot detect manifest type from %q: expected Dockerfile* or docker-compose*.yml/.yaml", path)
}

// Images returns the image references found in the manifest at path.
// Duplicates are removed; order of first appearance is kept.
func Images(path string) ([]string, error) {
	kind, err := Detect(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var refs []string
	if kind == Dockerfile {
		refs = dockerfileImages(string(data))
	} else {
		refs = composeImages(string(data))
	}
	return dedupe(refs), nil
}

var (
	fromRe    = regexp.MustCompile(`(?i)^\s*FROM\s+(?:--\S+\s+)*(\S+)`)
	imageRe   = regexp.MustCompile(`^\s*-?\s*image\s*:\s*(.+?)\s*$`)
	commentRe = regexp.MustCompile(`\s+#.*$`)
)

// dockerfileImages collects FROM targets, skipping build stage names and
// build args (e.g. FROM base AS builder, FROM ${BASE_IMAGE}).
func dockerfileImages(content string) []string {
	var stages []string
	var refs []string
	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		m := fromRe.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		ref := m[1]
		if strings.Contains(ref, "$") {
			continue
		}
		if contains(stages, ref) {
			continue
		}
		refs = append(refs, ref)
		if i := strings.Index(strings.ToUpper(sc.Text()), " AS "); i >= 0 {
			stages = append(stages, strings.TrimSpace(sc.Text()[i+4:]))
		}
	}
	return refs
}

// composeImages collects image: values, skipping empty values and values
// that reference variables (e.g. image: ${IMAGE}).
func composeImages(content string) []string {
	var refs []string
	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		m := imageRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		val := strings.TrimSpace(commentRe.ReplaceAllString(m[1], ""))
		val = strings.Trim(val, `"'`)
		if val == "" || strings.Contains(val, "$") {
			continue
		}
		refs = append(refs, val)
	}
	return refs
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func dedupe(refs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range refs {
		if seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}
