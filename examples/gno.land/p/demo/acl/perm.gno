package acl

import (
	"regexp"
	"sort"
    "strings"

	"gno.land/p/demo/ufmt"
)

type perm struct {
	verbs     []string
	resources []string
}

func (perm perm) hasPerm(verb, resource string) bool {
	// check verb
	verbOK := false
	for _, pattern := range perm.verbs {
		if match(pattern, verb) {
			verbOK = true
			break
		}
	}
	if !verbOK {
		return false
	}

	// check resource
	for _, pattern := range perm.resources {
		if match(pattern, resource) {
			return true
		}
	}
	return false
}

func match(pattern, target string) bool {
	if pattern == ".*" {
		return true
	}

	if pattern == target {
		return true
	}

	// regexp handling
	match, _ := regexp.MatchString(pattern, target)
	return match
}

func (p perm) Equal(p2 perm) bool {
	if len(p.verbs) != len(p2.verbs) {
		return false
	}
	if len(p.resources) != len(p2.resources) {
		return false
	}

	for i, verb := range p.verbs {
		if verb != p2.verbs[i] {
			return false
		}
	}

	for i, resource := range p.resources {
		if resource != p2.resources[i] {
			return false
		}
	}

	return true
}

func (p perm) String() string {
	// Sort the slices to ensure consistent ordering
    sortedVerbs := make([]string, len(p.verbs))
    copy(sortedVerbs, p.verbs)
    sort.Strings(sortedVerbs)

    sortedResources := make([]string, len(p.resources))
    copy(sortedResources, p.resources)
    sort.Strings(sortedResources)

    // Join the slices into a string representation
    return ufmt.Sprintf("verbs:[%s] resources:[%s]", 
    strings.Join(sortedVerbs, ", "), 
    strings.Join(sortedResources, ", "))
}
