package image

import "fmt"

// NormalizeExtraNames parses additional tag arguments into NamedAs() form.
// Names equal to primary (or duplicates) are dropped.
func NormalizeExtraNames(primary string, extras []string) ([]string, error) {
	seen := map[string]bool{primary: true}
	var out []string
	for _, raw := range extras {
		ref, err := Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid --tag %q: %w", raw, err)
		}
		name := ref.NamedAs()
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out, nil
}
