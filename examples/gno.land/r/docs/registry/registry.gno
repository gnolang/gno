package registry

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/nt/avl"
)

// Service represents a registered service
type Service struct {
	Name        string
	Description string
	Endpoint    string
	Owner       address
	Active      bool
}

var (
	services     avl.Tree // key (owner:name) -> *Service
	serviceCount int
)

const (
	maxNameLength        = 50
	maxDescriptionLength = 200
	maxEndpointLength    = 500
)

// Register adds a new service to the registry
func Register(_ realm, name, description, endpoint string) {
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

	caller := runtime.PreviousRealm().Address()
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
	serviceCount++
}

// Update updates an existing service (owner only)
func Update(_ realm, name, description, endpoint string) {
	if len(description) > maxDescriptionLength {
		panic("description too long")
	}
	if len(endpoint) > maxEndpointLength {
		panic("endpoint too long")
	}

	caller := runtime.PreviousRealm().Address()
	key := caller.String() + ":" + name

	val, ok := services.Get(key)
	if !ok {
		panic("service not found")
	}

	service := val.(*Service)
	if service.Owner != caller {
		panic("only owner can update service")
	}

	service.Description = description
	service.Endpoint = endpoint
	services.Set(key, service)
}

// Deactivate marks a service as inactive (owner only)
func Deactivate(_ realm, name string) {
	caller := runtime.PreviousRealm().Address()
	key := caller.String() + ":" + name

	val, ok := services.Get(key)
	if !ok {
		panic("service not found")
	}

	service := val.(*Service)
	if service.Owner != caller {
		panic("only owner can deactivate service")
	}

	service.Active = false
	services.Set(key, service)
}

// Delete removes a service permanently (owner only)
func Delete(_ realm, name string) {
	caller := runtime.PreviousRealm().Address()
	key := caller.String() + ":" + name

	val, ok := services.Get(key)
	if !ok {
		panic("service not found")
	}

	service := val.(*Service)
	if service.Owner != caller {
		panic("only owner can delete service")
	}

	services.Remove(key)
	serviceCount--
}

// Lookup returns a service by owner address and name
func Lookup(owner address, name string) *Service {
	key := owner.String() + ":" + name
	val, ok := services.Get(key)
	if !ok {
		return nil
	}
	return val.(*Service)
}

// GetAllServices returns all services (for iteration)
func GetAllServices() []*Service {
	var result []*Service
	services.Iterate("", "", func(key string, value any) bool {
		result = append(result, value.(*Service))
		return false
	})
	return result
}

func Render(path string) string {
	if path != "" {
		// Path format: "address:name"
		parts := strings.Split(path, ":")
		if len(parts) != 2 {
			return "# Invalid Path\n\nPath should be in format `address:name`."
		}

		owner := address(parts[0])
		name := parts[1]
		service := Lookup(owner, name)
		if service == nil {
			return "# Service Not Found\n\nNo service with address `" + parts[0] + "` and name `" + name + "` is registered."
		}
		return renderService(service)
	}

	return renderHome()
}

func renderHome() string {
	out := `# Registry Pattern

The registry pattern allows entities to register themselves in a shared
index. Other contracts or users can then look up registered entities.

Common use cases include name resolution (mapping addresses to human-readable
names), token discovery, and realm showcasing.

## This Example

This service registry allows anyone to register a service with a unique name.
Only the owner can update or deactivate their service, but anyone can look up
services by name.

## Live Demo

[Register Service](/r/docs/registry$help&func=Register)

### Registered Services

**Total services:** ` + strconv.Itoa(serviceCount) + `

`
	if serviceCount == 0 {
		out += "_No services registered yet._\n"
	} else {
		services.Iterate("", "", func(key string, value any) bool {
			service := value.(*Service)
			status := "✅"
			if !service.Active {
				status = "❌"
			}
			out += "- " + status + " [" + service.Name + "](/r/docs/registry:" + key + ")\n"
			return false
		})
	}

	out += `
---

## Code Pattern

` + "```go" + `
package registry

import (
    "chain/runtime"
    
    "gno.land/p/nt/avl"
)

type Entry struct {
    Name  string
    Owner address
    Data  string // Your custom payload
}

var (
    entries        avl.Tree // key (owner:name) -> *Entry
)

func Register(_ realm, name, data string) {
    caller := runtime.PreviousRealm().Address()
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

func Lookup(owner address, name string) *Entry {
    key := owner.String() + ":" + name
    val, ok := entries.Get(key)
    if !ok {
        return nil
    }
    return val.(*Entry)
}
` + "```" + `

## Real-World Examples in Gno

- [r/sys/users](/r/sys/users) - Official username registry. Users register
their username and address, and anyone can look them up.

- [r/demo/defi/grc20reg](/r/demo/defi/grc20reg) - GRC20 token registry.
Token realms register themselves so other contracts can discover available tokens.

- [r/leon/hor](/r/leon/hor) - Hall of Realms. Realms register themselves by
calling ^hor.Register()^ from their init function to be showcased in the exhibition.

[View source code](/r/docs/registry$source)

## See also

- [Factory Pattern](/r/docs/factory) - Creating and managing object instances
- [Crossing](/r/docs/crossing) - Understanding realm boundaries and caller validation
`
	return out
}

func renderService(service *Service) string {
	status := "Active ✅"
	if !service.Active {
		status = "Inactive ❌"
	}

	return `# Service: ` + service.Name + `

## Details

| Property | Value |
|----------|-------|
| Status | ` + status + ` |
| Description | ` + service.Description + ` |
| Endpoint | ` + service.Endpoint + ` |
| Owner | ` + service.Owner.String() + ` |

---

[Back to Registry](/r/docs/registry)
`
}
