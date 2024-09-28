package versionset

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

// VersionInfo represents a specific version of a component.
type VersionInfo struct {
	Name     string
	Version  string
	Optional bool
}

// VersionSet is a collection of VersionInfo.
type VersionSet []VersionInfo

// Sort arranges the VersionSet in ascending order by Name.
func (vs VersionSet) Sort() {
	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Name < vs[j].Name
	})
}

// Set updates or adds a VersionInfo to the VersionSet.
func (vs *VersionSet) Set(v VersionInfo) {
	for i, existing := range *vs {
		if existing.Name == v.Name {
			(*vs)[i] = v
			return
		}
	}
	*vs = append(*vs, v)
	vs.Sort()
}

// Get retrieves a VersionInfo by name.
func (vs VersionSet) Get(name string) (VersionInfo, bool) {
	for _, v := range vs {
		if v.Name == name {
			return v, true
		}
	}
	return VersionInfo{}, false
}

// CompatibleWith checks compatibility between two VersionSets.
func (vs VersionSet) CompatibleWith(other VersionSet) (VersionSet, error) {
	result := make(VersionSet, 0)
	incompatibilities := make([]string, 0)

	vMap := make(map[string]VersionInfo)
	for _, v := range vs {
		vMap[v.Name] = v
	}

	for _, otherV := range other {
		v, exists := vMap[otherV.Name]
		if !exists {
			if !otherV.Optional {
				incompatibilities = append(incompatibilities, fmt.Sprintf("Missing required version: %s", otherV.Name))
			}
			continue
		}

		compatible, resultV := compareVersions(v, otherV)
		if compatible {
			result = append(result, resultV)
		} else {
			incompatibilities = append(incompatibilities, fmt.Sprintf("Incompatible versions for %s: %s vs %s", v.Name, v.Version, otherV.Version))
		}

		delete(vMap, otherV.Name)
	}

	for _, v := range vMap {
		if !v.Optional {
			incompatibilities = append(incompatibilities, fmt.Sprintf("Missing required version: %s", v.Name))
		}
	}

	if len(incompatibilities) > 0 {
		return nil, fmt.Errorf("incompatibilities found:\n%s", strings.Join(incompatibilities, "\n"))
	}

	return result, nil
}

// compareVersions checks compatibility between two VersionInfos.
func compareVersions(v1, v2 VersionInfo) (bool, VersionInfo) {
	v1MM := semver.MajorMinor(v1.Version)
	v2MM := semver.MajorMinor(v2.Version)

	if semver.Major(v1MM) != semver.Major(v2MM) {
		return false, VersionInfo{}
	}

	resultV := VersionInfo{
		Name:     v1.Name,
		Optional: v1.Optional && v2.Optional,
	}

	if semver.Compare(v1MM, v2MM) > 0 {
		resultV.Version = v2.Version
	} else {
		resultV.Version = v1.Version
	}

	return true, resultV
}
