// Interactive prompt primitives for CLI wizard flows.
//
// This file provides reusable prompt functions that work with the commands.IO
// abstraction, making them testable with mock readers/writers and usable across
// any CLI tool in the Gno codebase.
//
// All prompt functions write to io.Err() (stderr) for prompts and error messages,
// read from io.In() (stdin) for user input, and return [ErrGoBack] when the user
// types "<" to navigate back to a previous wizard step.
//
// Available prompts:
//   - [PromptString]: free-text input with optional default and validation
//   - [PromptChoice]: single-key choice menu (e.g. [r]ealm, [P]ackage)
//   - [PromptSelect]: numbered list menu with name matching
//   - [PromptConfirm]: yes/no confirmation
//
// Use [IsInteractive] to check whether stdin is a terminal before entering
// an interactive flow.

package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// ErrGoBack is a sentinel error returned by prompt functions when the user
// types "<" to go back to a previous step in a wizard flow.
var ErrGoBack = fmt.Errorf("go back")

// IsInteractive returns true when standard input is a terminal,
// indicating that the program is running interactively and can
// prompt the user for input.
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Choice represents a single option in a PromptChoice menu.
type Choice struct {
	Key         string   // Short key the user can type (e.g. "r", "p", "m")
	Aliases     []string // Additional accepted inputs (e.g. "realm", "package")
	Description string   // Human-readable label shown in the prompt
	IsDefault   bool     // If true, this choice is selected when the user presses Enter
}

// SelectItem represents a single option in a numbered PromptSelect menu.
type SelectItem struct {
	Name        string // Short name (e.g. "basic", "dao")
	Description string // One-line description shown next to the name
}

// PromptString prompts the user for a string value with an optional default.
// If validate is non-nil, the input is validated and the user is re-prompted
// on failure. Returns ErrGoBack if the user types "<".
func PromptString(io IO, prompt string, defaultVal string, validate func(string) error) (string, error) {
	for {
		if defaultVal != "" {
			fmt.Fprintf(io.Err(), "%s [%s]: ", prompt, defaultVal)
		} else {
			fmt.Fprintf(io.Err(), "%s: ", prompt)
		}

		ans, err := readLine(io)
		if err != nil {
			return "", err
		}
		if ans == "<" {
			return "", ErrGoBack
		}
		if ans == "" {
			ans = defaultVal
		}
		if validate != nil {
			if err := validate(ans); err != nil {
				fmt.Fprintf(io.Err(), "%s\n", err)
				continue
			}
		}
		return ans, nil
	}
}

// PromptChoice presents a single-key choice menu (e.g. "[r]ealm, [P]ackage, [m]ain")
// and returns the index of the selected choice. The prompt string is displayed
// before the choices. Returns ErrGoBack if the user types "<".
func PromptChoice(io IO, prompt string, choices []Choice) (int, error) {
	for {
		fmt.Fprint(io.Err(), prompt)
		ans, err := readLine(io)
		if err != nil {
			return 0, err
		}
		if ans == "<" {
			return 0, ErrGoBack
		}

		lower := strings.ToLower(ans)

		// Check for default on empty input
		if lower == "" {
			for i, c := range choices {
				if c.IsDefault {
					return i, nil
				}
			}
			// No default found, show error
			fmt.Fprintf(io.Err(), "please enter a valid choice\n")
			continue
		}

		// Match by key or alias
		for i, c := range choices {
			if strings.EqualFold(c.Key, lower) {
				return i, nil
			}
			for _, alias := range c.Aliases {
				if strings.EqualFold(alias, lower) {
					return i, nil
				}
			}
		}

		fmt.Fprintf(io.Err(), "invalid choice: %q\n", ans)
	}
}

// PromptSelect presents a numbered list menu and returns the index of the
// selected item. If there is only one item, it is auto-selected without
// prompting. Returns ErrGoBack if the user types "<".
func PromptSelect(io IO, prompt string, items []SelectItem) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("no items available")
	}
	if len(items) == 1 {
		return 0, nil
	}

	for {
		fmt.Fprintf(io.Err(), "%s\n", prompt)
		for i, item := range items {
			fmt.Fprintf(io.Err(), "  %d. %s — %s\n", i+1, item.Name, item.Description)
		}
		fmt.Fprintf(io.Err(), "Choose [1]: ")

		ans, err := readLine(io)
		if err != nil {
			return 0, err
		}
		if ans == "<" {
			return 0, ErrGoBack
		}
		if ans == "" {
			return 0, nil // default to first item
		}

		// Try number
		if n, err := strconv.Atoi(ans); err == nil {
			if n >= 1 && n <= len(items) {
				return n - 1, nil
			}
			fmt.Fprintf(io.Err(), "invalid choice: %d (must be 1-%d)\n", n, len(items))
			continue
		}

		// Try name match
		for i := range items {
			if strings.EqualFold(items[i].Name, ans) {
				return i, nil
			}
		}
		fmt.Fprintf(io.Err(), "unknown choice: %q\n", ans)
	}
}

// PromptConfirm asks the user a yes/no question. Returns true for yes.
// Empty input defaults to the specified default value.
func PromptConfirm(io IO, prompt string, defaultYes bool) (bool, error) {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	fmt.Fprintf(io.Err(), "%s %s: ", prompt, hint)
	ans, err := readLine(io)
	if err != nil {
		return false, err
	}

	if ans == "" {
		return defaultYes, nil
	}

	lower := strings.ToLower(ans)
	return lower == "y" || lower == "yes", nil
}

// readLine reads a single line from io.In() via GetString and trims whitespace.
func readLine(io IO) (string, error) {
	s, err := io.GetString("")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}
