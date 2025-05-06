package gnoweb

import (
	"fmt"
	"os"
	"strings"
)

// IsHomePath checks if the given path is the home path.
func IsHomePath(path string) bool {
	return path == "/"
}

type AliasKind int

const (
	GnowebPath AliasKind = iota
	StaticMarkdown
)

type AliasTarget struct {
	value string
	kind  AliasKind
}

// aliases are gnoweb paths that are rewritten by the web handler.
var aliases = map[string]AliasTarget{
	"/":           {"/r/gnoland/home", GnowebPath},
	"/about":      {"/r/gnoland/pages:p/about", GnowebPath},
	"/gnolang":    {"/r/gnoland/pages:p/gnolang", GnowebPath},
	"/ecosystem":  {"/r/gnoland/pages:p/ecosystem", GnowebPath},
	"/start":      {"/r/gnoland/pages:p/start", GnowebPath},
	"/license":    {"/r/gnoland/pages:p/license", GnowebPath},
	"/contribute": {"/r/gnoland/pages:p/contribute", GnowebPath},
	"/events":     {"/r/gnoland/events", GnowebPath},
}

// SetAliases parses the given aliases string and sets the aliases map.
// Any value of form 'static:<file>' will be loaded as a static markdown file.
func SetAliases(aliasesStr string) error {
	// Split the aliases string by commas.
	aliasEntries := strings.Split(aliasesStr, ",")

	// Add each alias entry to the aliases map.
	for _, entry := range aliasEntries {
		parts := strings.Split(entry, "|")
		if len(parts) != 2 {
			return fmt.Errorf("invalid alias entry: %s", entry)
		}

		// Trim whitespace from both parts.
		parts[0] = strings.TrimSpace(parts[0])
		parts[1] = strings.TrimSpace(parts[1])

		// Check if the value is a path to a static file.
		if strings.HasPrefix(parts[1], "static:") {
			// If it is, load the static file content and set it as the alias.
			staticFilePath := strings.TrimPrefix(parts[1], "static:")

			content, err := os.ReadFile(staticFilePath)
			if err != nil {
				return fmt.Errorf("failed to read static file %s: %w", staticFilePath, err)
			}
			aliases[parts[0]] = AliasTarget{value: string(content), kind: StaticMarkdown}
		} else { // Otherwise, treat it as a normal alias.
			aliases[parts[0]] = AliasTarget{value: parts[1], kind: GnowebPath}
		}
	}

	return nil
}
