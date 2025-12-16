package templating

import (
	"fmt"
	"regexp"
	"strings"
)

var tokenRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_\-]+)\s*\}\}`)

func RenderString(in string, memory map[string]string) (string, error) {
	if in == "" {
		return "", nil
	}

	var missing []string
	out := tokenRe.ReplaceAllStringFunc(in, func(match string) string {
		parts := tokenRe.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		key := parts[1]
		val, ok := memory[key]
		if !ok {
			missing = append(missing, key)
			return ""
		}
		return val
	})

	if len(missing) > 0 {
		return "", fmt.Errorf("missing variables: %s", strings.Join(unique(missing), ", "))
	}
	return out, nil
}

func unique(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
