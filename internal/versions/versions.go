package versions

import (
	"fmt"
	"strings"
)

// Versions holds parsed tool version mappings.
type Versions struct {
	m map[string]string
}

// Parse parses versions.env content into a Versions map.
func Parse(data string) *Versions {
	v := &Versions{m: make(map[string]string)}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, val, ok := strings.Cut(line, "="); ok {
			v.m[strings.TrimSpace(k)] = strings.TrimSpace(val)
		}
	}
	return v
}

// Get returns a version value or panics if missing.
func (v *Versions) Get(key string) string {
	val, ok := v.m[key]
	if !ok {
		panic(fmt.Sprintf("missing version key: %s", key))
	}
	return val
}
