// Package registry is a service registry.
// Callers register, update, and remove named entries stored in an avl tree.
// Each entry is owned by the address that created it.
package registry

import (
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/markdown/sanitize/v0"
)

// Service represents a registered service
type Service struct {
	Name        string
	Description string
	Endpoint    string
	Owner       address
	Active      bool
}

var services avl.Tree // key (owner:name) -> *Service

const (
	maxNameLength        = 50
	maxDescriptionLength = 200
	maxEndpointLength    = 500
)

// Register adds a new service to the registry
func Register(cur realm, name, description, endpoint string) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		panic("name cannot be empty")
	}
	if len(name) > maxNameLength {
		panic("name too long")
	}
	if len(description) > maxDescriptionLength {
		panic("description too long")
	}
	if len(endpoint) > maxEndpointLength {
		panic("endpoint too long")
	}

	caller := cur.Previous().Address()
	key := caller.String() + ":" + name

	if services.Has(key) {
		panic("service with this name from this address already exists")
	}

	service := &Service{
		Name:        name,
		Description: description,
		Endpoint:    endpoint,
		Owner:       caller,
		Active:      true,
	}

	services.Set(key, service)
}

// Update updates an existing service (owner only)
func Update(cur realm, name, description, endpoint string) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if len(description) > maxDescriptionLength {
		panic("description too long")
	}
	if len(endpoint) > maxEndpointLength {
		panic("endpoint too long")
	}

	caller := cur.Previous().Address()
	key := caller.String() + ":" + name

	val, ok := services.Get(key)
	if !ok {
		panic("service not found")
	}

	service := val.(*Service)
	service.Description = description
	service.Endpoint = endpoint
	services.Set(key, service)
}

// Deactivate marks a service as inactive (owner only)
func Deactivate(cur realm, name string) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	caller := cur.Previous().Address()
	key := caller.String() + ":" + name

	val, ok := services.Get(key)
	if !ok {
		panic("service not found")
	}

	service := val.(*Service)
	service.Active = false
	services.Set(key, service)
}

// Delete removes a service permanently (owner only)
func Delete(cur realm, name string) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	caller := cur.Previous().Address()
	key := caller.String() + ":" + name

	if !services.Has(key) {
		panic("service not found")
	}

	services.Remove(key)
}

// Lookup returns a copy of a service by owner address and name. Returning a
// value (not the stored *Service) keeps callers from mutating registry state
// through the returned pointer. See /r/docs/mistakes.
func Lookup(owner address, name string) (Service, bool) {
	key := owner.String() + ":" + name
	val, ok := services.Get(key)
	if !ok {
		return Service{}, false
	}
	return *val.(*Service), true
}

// GetAllServices returns copies of all services. Copies, not the stored
// pointers, keep callers from mutating registry state. See /r/docs/mistakes.
func GetAllServices() []Service {
	var result []Service
	services.Iterate("", "", func(key string, value any) bool {
		result = append(result, *value.(*Service))
		return false
	})
	return result
}

func Render(path string) string {
	if path != "" {
		// Path format: "address:name". Split on the first colon only, since a
		// bech32 address never contains one but a name might.
		parts := strings.SplitN(path, ":", 2)
		if len(parts) != 2 {
			return "# Invalid Path\n\nPath should be in format `address:name`."
		}

		owner := address(parts[0])
		name := parts[1]
		service, ok := Lookup(owner, name)
		if !ok {
			return "# Service Not Found\n\nNo service with address `" + sanitize.InlineText(parts[0]) + "` and name `" + sanitize.InlineText(name) + "` is registered."
		}
		return renderService(service)
	}

	return renderHome()
}

func renderHome() string {
	out := `# Registry Pattern

Allows entities to register themselves in a shared
index. Other contracts can then look up registered entities.
Common use cases include mapping data, token discovery and realm showcasing.

## Code Pattern

` + "```go" + `
package registry

import (
    "gno.land/p/nt/avl/v0"
)

type Entry struct {
    Name  string
    Owner address
    Data  string // Your custom payload
}

var (
    entries        avl.Tree // key (owner:name) -> *Entry
)

func Register(cur realm, name, data string) {
    // Reject a non-current realm handle
    if !cur.IsCurrent() {
        panic("invalid realm")
    }
    caller := cur.Previous().Address()
    key := caller.String() + ":" + name
    
    if entries.Has(key) {
        panic("you already have an entry with this name")
    }
    
    entries.Set(key, &Entry{
        Name:  name,
        Owner: caller,
        Data:  data,
    })
}

// Return a copy, not the stored *Entry, so callers can't mutate state
// through the returned pointer (see /r/docs/mistakes).
func Lookup(owner address, name string) (Entry, bool) {
    key := owner.String() + ":" + name
    val, ok := entries.Get(key)
    if !ok {
        return Entry{}, false
    }
    return *val.(*Entry), true
}
` + "```" + `


## This Example

This service registry allows anyone to register a service with a unique name.
Only the owner can update or deactivate their service, but anyone can look up
services by name.

## Live Demo

[Register Service](/r/docs/registry$help&func=Register)

### Registered Services

`
	count := services.Size()

	if count == 0 {
		out += "_No services registered yet._\n"
	} else {
		services.Iterate("", "", func(key string, value any) bool {
			service := value.(*Service)
			status := "✅"
			if !service.Active {
				status = "❌"
			}
			out += "- " + status + " [" + sanitize.InlineText(service.Name) + "](/r/docs/registry:" + key + ")\n"
			return false
		})
	}

	out += `
---


## Examples

- [r/sys/users](/r/sys/users) - Official username registry. Users register
their username and address, and anyone can look them up.

- [r/demo/defi/grc20reg](/r/demo/defi/grc20reg) - GRC20 token registry.
Token realms register themselves so other contracts can discover available tokens.

- [r/leon/hor](/r/leon/hor) - Hall of Realms. Realms register themselves by
calling ` + "`hor.Register()`" + ` from their init function to be showcased in the exhibition.

[View source code](/r/docs/registry$source)
`
	return out
}

func renderService(service Service) string {
	status := "Active ✅"
	if !service.Active {
		status = "Inactive ❌"
	}

	return `# Service: ` + sanitize.InlineText(service.Name) + `

## Details

| Property | Value |
|----------|-------|
| Status | ` + status + ` |
| Description | ` + sanitize.TableCell(service.Description) + ` |
| Endpoint | ` + sanitize.TableCell(service.Endpoint) + ` |
| Owner | ` + service.Owner.String() + ` |

---

[Back to Registry](/r/docs/registry)
`
}
