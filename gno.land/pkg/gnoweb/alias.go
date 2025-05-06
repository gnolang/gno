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

// aliases are gnoweb paths that are rewritten by the web handler.
var aliases = map[string]string{
	"/":           "/r/gnoland/home",
	"/about":      "/r/gnoland/pages:p/about",
	"/gnolang":    "/r/gnoland/pages:p/gnolang",
	"/ecosystem":  "/r/gnoland/pages:p/ecosystem",
	"/partners":   "/r/gnoland/pages:p/partners",
	"/testnets":   "/r/gnoland/pages:p/testnets",
	"/start":      "/r/gnoland/pages:p/start",
	"/license":    "/r/gnoland/pages:p/license",
	"/contribute": "/r/gnoland/pages:p/contribute",
	"/events":     "/r/gnoland/events",
}

// GetAlias retrieves the target and static status of the given path if any.
func GetAlias(path string) (target string, exists, static bool) {
	target, exists = aliases[path]

	// If the alias is a static file, set the static flag and trim the prefix.
	if strings.HasPrefix(target, "static:") {
		static = true
		target = strings.TrimPrefix(target, "static:")
	}

	return
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
			aliases[parts[0]] = fmt.Sprintf("static:%s", string(content))
		} else { // Otherwise, treat it as a normal alias.
			aliases[parts[0]] = parts[1]
		}
	}

	return nil
}
