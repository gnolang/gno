package acl

import "regexp"

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
