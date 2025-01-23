package ruler

import "std"

var PrevDAOs []string
var DeletedDAOs []string
var CurrentDAO string

func init() {
	CurrentDAO = "gno.land/r/gov/dao/v3"
}

// Render outputs a human readable state of the ruler
func Render(string) string {
	out := ""
	out += "Current GovDAO in charge: " + CurrentDAO + "\n\n"
	out += "Previous GovDAO implementations in charge: \n"
	for _, d := range PrevDAOs {
		out += "- " + d +"\n"
	}

	out += "\nPrevious deleted GovDAOs due to security issues or similar: \n"
	for _, d := range DeletedDAOs {
		out += "- " + d +"\n"
	} 

	return out
}

// Rollback allows to rollback to a previous validated package in case the actual one has a breaking bug.
// Any current or actual package is allowed to call Rollback function. If some of previous GovDAOs are insecure, 
// we recommend to call Remove method to avoid letting it rollback to itself.
func Rollback(pkg string) {
	if !checkCurrent() && !checkPrev() {
		panic("package " + pkg + " cannot be rollback by " + caller())
	}
	
	idx := has(pkg)
	if idx < 0 {
		panic("package " + pkg + " does not exists, so it cannot be rollback")
	}

	CurrentDAO = pkg
	remove(idx)
}

// New changes who is the current DAO in charge. it can be called only by the actual DAO in charge only.
func New(pkg string) {
	if !checkCurrent() {
		panic("package " + pkg + " cannot be added by " + caller())
	}

	PrevDAOs = append(PrevDAOs, CurrentDAO)
	CurrentDAO = pkg
}

// Remove deletes previous DAOs that can contain security issues. That will avoid calls to rollback from them.
func Remove(pkg string) {
	if !checkCurrent() {
		panic("package " + pkg + " cannot be removed by " + caller())
	}

	idx := has(pkg)
	if idx < 0 {
		panic("package " + pkg + " does not exists, so it cannot be rollback")
	}

	DeletedDAOs = append(DeletedDAOs, pkg)

	remove(idx)
}

func caller() string {
	return std.PrevRealm().PkgPath()
}

func checkCurrent() bool {
	return caller() == CurrentDAO
}

func checkPrev() bool {
	return has(caller()) >= 0 
}

func has(pkg string) int {
	for i, d := range PrevDAOs {
		if pkg == d {
			return i
		}	
	}

	return -1
}

func remove(i int) {
	if i < 0 {
		return 
	}

    PrevDAOs = append(PrevDAOs[:i], PrevDAOs[i+1:]...)
}