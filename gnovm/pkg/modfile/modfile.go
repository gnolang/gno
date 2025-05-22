package modfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

// ErrModfileNotFound is returned by [ReadModfile] when, even after traversing
// up to the root directory (if requested), a gno.toml file could not be found.
var ErrModfileNotFound = errors.New("gno.toml file not found")

// ErrModfileExists is returned by [CreateModfile] if a gno.toml file already exists.
var ErrModfileExists = errors.New("gno.toml file already exists")

// ErrPkgPathEmpty is returned by [ParseModfile] if the PkgPath field is empty.
var ErrPkgPathEmpty = errors.New("pkgPath must be set in gno.toml")

// Modfile represents the structure of a gno.toml file.
type Modfile struct {
	PkgPath  string `toml:"pkgPath" json:"pkgPath"`
	Draft    bool   `toml:"draft,omitempty" json:"draft,omitempty"`
	Private  bool   `toml:"private,omitempty" json:"private,omitempty"`
	Uploader string `toml:"uploader,omitempty" json:"uploader,omitempty"`
	// Replace
	// ...
}

// ParseModfile parses the content of a gno.toml file.
func ParseModfile(data []byte) (*Modfile, error) {
	var mf Modfile
	if err := toml.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gno.toml: %w", err)
	}
	// Basic validation
	if mf.PkgPath == "" {
		return nil, ErrPkgPathEmpty
	}
	return &mf, nil
}

// CreateModfile creates a new gno.toml file in the specified directory
// using the provided Modfile struct.
func CreateModfile(dir string, mf *Modfile) error {
	if !filepath.IsAbs(dir) {
		return fmt.Errorf("dir %q is not absolute", dir)
	}
	if mf == nil {
		return errors.New("provided modfile data cannot be nil")
	}
	if mf.PkgPath == "" {
		return ErrPkgPathEmpty
	}

	modfilePath := filepath.Join(dir, "gno.toml")
	if _, err := os.Stat(modfilePath); err == nil {
		return fmt.Errorf("%w in %s", ErrModfileExists, dir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check for existing gno.toml: %w", err)
	}

	// Use the provided mf directly
	data, err := toml.Marshal(*mf) // Marshal the provided Modfile struct
	if err != nil {
		return fmt.Errorf("failed to marshal gno.toml content: %w", err)
	}

	if err := os.WriteFile(modfilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write gno.toml file: %w", err)
	}
	return nil
}

// ReadModfile reads and parses a gno.toml file.
// If findInParents is true, it searches in the given directory and its ancestors.
// Otherwise, it only checks the specified directory.
func ReadModfile(dir string, findInParents bool) (*Modfile, error) {
	var modfilePath string
	var err error

	if findInParents {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", dir, err)
		}

		currentDir := absDir
		volumeRoot := filepath.VolumeName(currentDir) + string(filepath.Separator)

		for {
			tryPath := filepath.Join(currentDir, "gno.toml")
			if _, statErr := os.Stat(tryPath); statErr == nil {
				modfilePath = tryPath
				break
			} else if !os.IsNotExist(statErr) {
				return nil, fmt.Errorf("error checking for gno.toml in %s: %w", currentDir, statErr)
			}

			if currentDir == volumeRoot || currentDir == filepath.Dir(currentDir) { // Reached filesystem root
				return nil, fmt.Errorf("%w: searched from %s upwards", ErrModfileNotFound, absDir)
			}
			currentDir = filepath.Dir(currentDir)
		}
	} else {
		modfilePath = filepath.Join(dir, "gno.toml")
		if _, statErr := os.Stat(modfilePath); statErr != nil {
			if os.IsNotExist(statErr) {
				return nil, fmt.Errorf("%w: checked in %s", ErrModfileNotFound, dir)
			}
			return nil, fmt.Errorf("error checking for gno.toml in %s: %w", dir, statErr)
		}
	}

	data, err := os.ReadFile(modfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gno.toml from %s: %w", modfilePath, err)
	}

	return ParseModfile(data)
}

// String implements the fmt.Stringer interface for Modfile.
// It returns the TOML representation of the Modfile.
// If marshalling fails, it returns a string indicating the error.
func (mf *Modfile) String() string {
	if mf == nil {
		return "<nil>"
	}
	data, err := toml.Marshal(*mf)
	if err != nil {
		// Return a best-effort string if marshalling fails
		return fmt.Sprintf("Modfile{PkgPath: %q, ErrorMarshalling: %v}", mf.PkgPath, err)
	}
	return string(data)
}
