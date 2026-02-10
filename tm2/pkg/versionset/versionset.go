package versionset

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

// VersionInfo is used to negotiate between clients.
type VersionInfo struct {
	Name     string // abci, p2p, app, block, etc.
	Version  string // semver.
	Optional bool   // default required.
}

type VersionSet []VersionInfo

func (pvs VersionSet) Sort() {
	sort.Slice(pvs, func(i, j int) bool {
		if pvs[i].Name < pvs[j].Name {
			return true
		} else if pvs[i].Name == pvs[j].Name {
			panic("should not happen")
		} else {
			return false
		}
	})
}

func (pvs *VersionSet) Set(pv VersionInfo) {
	for i, pv2 := range *pvs {
		if pv2.Name == pv.Name {
			(*pvs)[i] = pv
			return
		}
	}
	*pvs = append(*pvs, pv)
	pvs.Sort()
}

func (pvs VersionSet) Get(name string) (pv VersionInfo, ok bool) {
	for _, pv := range pvs {
		if pv.Name == name {
			return pv, true
		}
	}
	ok = false
	return
}

// Returns an error if not compatible.
// Otherwise, returns the set of compatible interfaces.
// Only the Major and Minor versions are returned; Patch, Prerelease, and Build
// portions of Semver2.0 are discarded in the resulting intersection
// VersionSet.
// TODO: test
func (pvs VersionSet) CompatibleWith(other VersionSet) (res VersionSet, err error) {
	var errs []string
	type pvpair [2]*VersionInfo
	name2Pair := map[string]*pvpair{}
	for _, pv := range pvs {
		pv := pv
		name2Pair[pv.Name] = &pvpair{&pv, nil}
	}
	for _, pv := range other {
		pv := pv
		item, ok := name2Pair[pv.Name]
		if ok {
			item[1] = &pv
		} else {
			name2Pair[pv.Name] = &pvpair{nil, &pv}
		}
	}
	for _, pair := range name2Pair {
		pv1, pv2 := pair[0], pair[1]
		if pv1 == nil {
			if pv2.Optional {
				// fine
			} else {
				errs = append(errs, fmt.Sprintf("Other VersionSet requires %v", pv2))
			}
		} else if pv2 == nil {
			if pv1.Optional {
				// fine
			} else {
				errs = append(errs, fmt.Sprintf("Our VersionSet requires %v", pv1))
			}
		} else {
			pv1mm := semver.MajorMinor(pv1.Version)
			pv2mm := semver.MajorMinor(pv2.Version)
			if semver.Major(pv1mm) == semver.Major(pv2mm) {
				if semver.Compare(semver.Major(pv1mm), semver.Major(pv2mm)) > 0 {
					res = append(res, VersionInfo{Name: pv1.Name, Version: pv2mm, Optional: pv1.Optional && pv2.Optional})
				} else {
					res = append(res, VersionInfo{Name: pv1.Name, Version: pv1mm, Optional: pv1.Optional && pv2.Optional})
				}
			} else {
				errs = append(errs, fmt.Sprintf("VersionInfos not compatible: %v vs %v", pv1, pv2))
			}
		}
	}
	if errs != nil {
		return res, fmt.Errorf("VersionSet not compatible...\n%s", strings.Join(errs, "\n"))
	}
	return res, nil
}
