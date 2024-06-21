package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/mux"
	"std"
)

type contextKey string

type UserProfile struct {
	Username string
	Address  std.Address
	Fields   *avl.Tree
}

var profiles *avl.Tree      // address -> UserProfile
var usernameIndex *avl.Tree // username -> address
var router *mux.Router

func init() {
	profiles = avl.NewTree()
	usernameIndex = avl.NewTree()
	router = mux.NewRouter()

	router.HandleFunc("", homeHandler)
	router.HandleFunc("info/{username}", profileHandler)
}

// CreateProfile creates a new user profile with a username and address
func CreateProfile(username string) {
	caller := std.GetOrigCaller()

	// Check if a profile already exists for this address
	if _, exists := profiles.Get(caller.String()); exists {
		panic("a profile already exists for this address")
	}

	// Check if the username is already taken
	if username != "" {
		if _, exists := usernameIndex.Get(username); exists {
			panic("username is already taken")
		}
	}

	profile := UserProfile{
		Username: username,
		Address:  caller,
		Fields:   avl.NewTree(),
	}

	profiles.Set(caller.String(), profile)
	usernameIndex.Set(username, caller.String())
	std.Emit("UserProfileCreated", "username", username, "address", caller.String())
}

func UpdateProfile(username string, fields ...interface{}) {
	profile := GetProfileByUsername(username)

	// Check for username change
	var newUsername string
	for i := 0; i < len(fields); i += 2 {
		if fields[i] == "Username" {
			newUsername = fields[i+1].(string)
			break
		}
	}

	// Update username if it has changed
	if newUsername != "" && newUsername != profile.Username {
		if _, exists := usernameIndex.Get(newUsername); exists {
			panic("username is already taken")
		}
		usernameIndex.Remove(profile.Username)
		profile.Username = newUsername
		usernameIndex.Set(newUsername, profile.Address.String())
		std.Emit("UserProfileUpdated", "username", newUsername)
	}

	// Update existing fields only
	for i := 0; i < len(fields); i += 2 {
		key := fields[i].(string)
		value := fields[i+1]
		if key == "Username" {
			continue
		}
		if _, exists := profile.Fields.Get(key); exists {
			profile.Fields.Set(key, value)
			std.Emit("UserProfileUpdated", "username", profile.Username, "key", key, "value", ufmt.Sprintf("%v", value))
		} else {
			panic(ufmt.Sprintf("field '%s' does not exist in profile", key))
		}
	}

	profiles.Set(profile.Address.String(), profile)
}

// AddFields adds multiple fields to a UserProfile using key-value pairs
func AddFields(profile UserProfile, fields ...interface{}) UserProfile {
	if len(fields)%2 != 0 {
		panic("fields must be in key-value pairs")
	}

	for i := 0; i < len(fields); i += 2 {
		key := fields[i].(string)
		value := fields[i+1]

		// Get the current values
		var values []interface{}
		if existing, exists := profile.Fields.Get(key); exists {
			values = existing.([]interface{})
		}

		// Append the new value
		values = append(values, value)
		profile.Fields.Set(key, values)
		std.Emit("UserProfileUpdated", "username", profile.Username, "key", key, "value", ufmt.Sprintf("%v", value))
	}
	profiles.Set(profile.Address.String(), profile)
	return profile
}

// RemoveProfile deletes a user profile
func RemoveProfile(username string) {
	profile := GetProfileByUsername(username)
	profiles.Remove(profile.Address.String())
	usernameIndex.Remove(username)
	std.Emit("UserProfileDeleted", "username", username, "address", profile.Address.String())
}

// GetField retrieves a field from a UserProfile
func GetField(profile UserProfile, key string) interface{} {
	if value, exists := profile.Fields.Get(key); exists {
		return value
	}
	return nil
}

// GetProfileByUsername retrieves a profile by username
func GetProfileByUsername(username string) UserProfile {
	if addr, exists := usernameIndex.Get(username); exists {
		if profile, ok := profiles.Get(addr.(string)); ok {
			return profile.(UserProfile)
		}
	}
	panic("username not found")
}

// GetProfileByAddress retrieves a profile by address
func GetProfileByAddress(address std.Address) UserProfile {
	if profile, exists := profiles.Get(address.String()); exists {
		return profile.(UserProfile)
	}
	panic("profile not found")
}


// formatFieldValue formats the value for display
func formatFieldValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return ufmt.Sprintf("%d", v)
	case float64:
		return ufmt.Sprintf("%f", v)
	case []interface{}:
		var result string
		for _, val := range v {
			result += formatFieldValue(val) + ", "
		}
		if len(result) > 0 {
			return result[:len(result)-2] // Remove trailing comma and space
		}
		return result
	default:
		return ufmt.Sprintf("%v", v)
	}
}

// Handlers

// homeHandler lists all profiles
func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<html><body>")
	b.WriteString("<h1>Profiles List</h1><br>")
	usernameIndex.Iterate("", "", func(key string, value interface{}) bool {
		username := key
		b.WriteString("<p>Username: " + username + " [<a href='/r/demo/profile:info/" + username + "'>View Profile</a>]</p>")
		return false
	})
	b.WriteString("</body></html>")
	res.Write(b.String())
}

// profileHandler displays a user profile
func profileHandler(res *mux.ResponseWriter, req *mux.Request) {
	username := req.GetVar("username")
	profile := GetProfileByUsername(username)

	var b bytes.Buffer

	b.WriteString("<html><body>")
	b.WriteString("<p>[<a href='/r/demo/profile'>Home</a>]</p>")
	b.WriteString("<h1>Profile of " + profile.Username + "</h1>")
	b.WriteString("<p>Address: " + profile.Address.String() + "</p>")

	profile.Fields.Iterate("", "", func(key string, value interface{}) bool {
		b.WriteString("<p>" + key + ": " + formatFieldValue(value) + "</p>")
		return false
	})

	b.WriteString("</body></html>")
	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
