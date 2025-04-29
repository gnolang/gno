package home

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	realmAllowPrefix = []string{}
	registry    = avl.NewTree()
	render []string
)

func SetOutput(s ...string) {
    AssertAdminAccess()
    render = []string{}
    appendOutput(s...)
}

func appendOutput(s ...string) {
    for _, v := range s {
        render = append(render, v)
    }
}

func init() {
	realmAllowPrefix = append(realmAllowPrefix, std.CurrentRealm().PkgPath()+"/patch")
	appendOutput("foo", "bar")
    registry.Set("foo", "testing!")
    registry.Set("bar", "\n123")
}

/* Auth API */
func HasAllowedPrefix() bool {
    prevPath := std.PreviousRealm().PkgPath()
    if prevPath != "" {
        return false
    }
    currentPath := std.CurrentRealm().PkgPath()
    for _, callerPath := range realmAllowPrefix {
        if strings.HasPrefix(currentPath, callerPath) {
            return true
        }
    }
    return false
}

func AssertAdminAccess() {
    if !HasAllowedPrefix() {
        panic("access denied: " + std.PreviousRealm().PkgPath() +
            " realm must match an allowed prefix:[" + strings.Join(realmAllowPrefix, ",") + "]" +
            " got: " + std.CurrentRealm().PkgPath())
    }
}

func Set(key string, val interface{}) {
    AssertAdminAccess()
    registry.Set(key, val)
}

func GetString(key string) string {
    val, _ := registry.Get(key)
    switch val.(type) {
    case string, int, float64, bool:
        return ufmt.Sprintf("%v", val)
    default:
        return ""
    }
}

func Render(_ string) string {
    out := strings.Builder{}
    for _, key := range render {
        val := GetString(key)
        if val != "" {
            out.WriteString(val)
        }
    }
    return out.String()
}

