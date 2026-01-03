package userprofile

import (
	"gno.land/r/demo/profile"
	"gno.land/r/sys/users"
)

// This example demonstrates using both r/sys/users and r/demo/profile
// to create a rich user experience.

func Render(path string) string {
	if path == "" {
		return renderHome()
	}

	return renderUser(path)
}

func renderHome() string {
	return `# User & Profile Integration

This example demonstrates how to use both [` + "`/r/sys/users`" + `](/r/sys/users) and [` + "`/r/demo/profile`" + `](/r/demo/profile)
to create rich user experiences in your realm.

## /r/sys/users

The [user registry](/r/sys/users) provides:
- Username registration
- Address ↔ Username resolution

` + "```go" + `
import "gno.land/r/sys/users"

// Resolve a username to user data
data, _ := users.ResolveName("alice")
addr := data.Address()

// Resolve an address to user data
data, _ = users.ResolveAddress(someAddress)
name := data.Name()

// Resolve either (smart detection)
data, _ = users.ResolveAny("alice")
data, _ = users.ResolveAny("g1...")
` + "```" + `

## /r/demo/profile

The [profile registry](/r/demo/profile) provides:
- Display name
- Avatar
- Bio
- Custom fields

[Set your profile →](/r/demo/profile$help&func=SetStringField)

` + "```go" + `
import "gno.land/r/demo/profile"

// Get profile fields
displayName := profile.GetStringField(addr, profile.DisplayName, "Anonymous")
avatar := profile.GetStringField(addr, profile.Avatar, "")
bio := profile.GetStringField(addr, profile.Bio, "")

// Set profile fields (for current user)
profile.SetStringField(profile.DisplayName, "Alice")
profile.SetStringField(profile.Bio, "Gno developer")
` + "```" + `

## Combined Usage Pattern

` + "```go" + `
func RenderUser(nameOrAddr string) string {
    // First, resolve the user
    userData, _ := users.ResolveAny(nameOrAddr)
    if userData == nil {
        return "User not found"
    }
    
    addr := userData.Address()
    name := userData.Name()
    
    // Then, get their profile
    displayName := profile.GetStringField(addr, profile.DisplayName, name)
    avatar := profile.GetStringField(addr, profile.Avatar, "")
    bio := profile.GetStringField(addr, profile.Bio, "No bio")
    
    // Render combined info
    out := "# " + displayName + "\n"
    if avatar != "" {
        out += "![avatar](" + avatar + ")\n"
    }
    out += bio + "\n"
    out += "\nUsername: @" + name + "\n"
    out += "Address: " + addr.String() + "\n"
    
    return out
}
` + "```" + `

## Live Demo

Try looking up at test user:
- [@test1](/r/docs/userprofile:test1)

---

See also:
- [r/sys/users](/r/sys/users) - User registry
- [r/demo/profile](/r/demo/profile) - Profile data
- [Resolving Users](/r/docs/resolveusers) - Basic user resolution
- [Registry Pattern](/r/docs/registry) - How registries work
- [Crossing](/r/docs/crossing) - Understanding realm boundaries
`
}

func renderUser(nameOrAddr string) string {
	out := "# User Lookup\n\n"

	// Try to resolve the user
	userData, _ := users.ResolveAny(nameOrAddr)
	if userData == nil {
		out += "User not found: `" + nameOrAddr + "`\n\n"
		out += "[Back to Home](/r/docs/userprofile)\n"
		return out
	}

	addr := userData.Addr()
	name := userData.Name()

	out += "## Found User\n\n"

	// User registry data
	out += "### Registry Data (r/sys/users)\n\n"
	out += "| Field | Value |\n"
	out += "|-------|-------|\n"
	out += "| Username | @" + name + " |\n"
	out += "| Address | `" + addr.String() + "` |\n"
	out += "| Profile Link | " + userData.RenderLink("") + " |\n"
	out += "\n"

	// Profile data
	out += "### Profile Data (r/demo/profile)\n\n"

	displayName := profile.GetStringField(addr, profile.DisplayName, "_not set_")
	avatar := profile.GetStringField(addr, profile.Avatar, "_not set_")
	bio := profile.GetStringField(addr, profile.Bio, "_not set_")

	out += "| Field | Value |\n"
	out += "|-------|-------|\n"
	out += "| Display Name | " + displayName + " |\n"
	out += "| Avatar | " + avatar + " |\n"
	out += "| Bio | " + bio + " |\n"
	out += "\n"

	out += "---\n\n"
	out += "[Back to Home](/r/docs/userprofile)\n"

	return out
}
