package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"std"
	"time"
)

type WebsiteInfo struct {
	Name string
	URL  string
}

type UserProfile struct {
	Username            string
	Address             std.Address
	AvatarURL           string
	Age                 string
	Gender              string
	Websites            *avl.Tree // Name -> URL
	Country             string
	City                string
	ModificationHistory map[string][]string
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
	router.HandleFunc("history/{username}", historyHandler)
}

func CreateProfile(username, avatarurl, age, gender, websiteName, websiteURL, country, city string) {
	caller := std.GetOrigCaller()

	// Check if a profile already exists for this address
	if _, exists := profiles.Get(caller.String()); exists {
		panic("a profile already exists for this address")
	}

	// Check if the username is already taken
	if username != "" {
		_, exists := usernameIndex.Get(username)
		if exists {
			panic("Username is already taken")
		}
	}

	websites := avl.NewTree()
	websites.Set(websiteName, websiteURL)

	profile := UserProfile{
		Username:  username,
		Address:   caller,
		AvatarURL: avatarurl,
		Age:       age,
		Gender:    gender,
		Websites:  websites,
		Country:   country,
		City:      city,
	}

	profiles.Set(caller.String(), profile)

	usernameIndex.Set(username, caller.String())
	std.Emit("UserProfileCreated", "username", username, "address", caller.String())
}

func UpdateProfile(username, avatarurl, age, gender, websiteName, websiteURL, country, city string) {
    caller := std.GetOrigCaller()
    value, exists := profiles.Get(caller.String())
    if !exists {
        panic("Profile not found")
    }

    // Retrieve the existing profile
    existingProfile := value.(UserProfile)
    changes := make(map[string]string)

    // update the username and manage index if changed
    if username != "" && username != existingProfile.Username {
        _, exists := usernameIndex.Get(username)
        if exists {
            panic("Username is already taken")
        }

        if existingProfile.Username != "" {
            usernameIndex.Remove(existingProfile.Username)
        }
        usernameIndex.Set(username, caller.String())

        changes["Username"] = "Changed from " + existingProfile.Username + " to " + username
        existingProfile.Username = username
    }

    // Initialize the ModificationHistory map if it's nil
    if existingProfile.ModificationHistory == nil {
        existingProfile.ModificationHistory = make(map[string][]string)
    }

    // Function to update the profile field if changed
    updateField := func(field *string, newValue, fieldName string, profile *UserProfile) {
        if newValue != "" && *field != newValue {
            changeDescription := "Changed from " + *field + " to " + newValue
            profile.ModificationHistory[fieldName] = append(profile.ModificationHistory[fieldName], time.Now().Format(time.RFC3339)+": "+changeDescription)
            changes[fieldName] = changeDescription
            *field = newValue
        }
    }

    // Update fields and log changes
    updateField(&existingProfile.AvatarURL, avatarurl, "AvatarURL", &existingProfile)
    updateField(&existingProfile.Age, age, "Age", &existingProfile)
    updateField(&existingProfile.Gender, gender, "Gender", &existingProfile)
    updateField(&existingProfile.Country, country, "Country", &existingProfile)
    updateField(&existingProfile.City, city, "City", &existingProfile)

    // Update the website tree if a new website is provided
    if websiteName != "" && websiteURL != "" {
        existingProfile.Websites.Set(websiteName, websiteURL)
        changeDescription := "Updated website " + websiteName + " to " + websiteURL
        existingProfile.ModificationHistory["Websites"] = append(existingProfile.ModificationHistory["Websites"], time.Now().Format(time.RFC3339)+": "+changeDescription)
        changes["Website"] = changeDescription
    }

    // Update the profile in the tree
    profiles.Set(caller.String(), existingProfile)

    // Emit the changes
    if len(changes) > 0 {
        emitArgs := []string{"address", caller.String()}
        for key, value := range changes {
            emitArgs = append(emitArgs, key, value)
        }
        std.Emit("UserProfileUpdated", emitArgs...)
    }
}

// Display complete Profile by Address
func DisplayProfileByAddress(address std.Address) UserProfile {
	if value, exists := profiles.Get(address.String()); exists {
		return value.(UserProfile)
	}

	panic("profile not found")
}

// Get complete profile by username
func DisplayProfileByUsername(username string) UserProfile {
	if addr, exists := usernameIndex.Get(username); exists {
		return DisplayProfileByAddress(std.Address(addr.(string)))
	}

	panic("username not found")
}

// Display username
func DisplayUsernameByAddress(address std.Address) string {
	if value, exists := profiles.Get(address.String()); exists {
		return value.(UserProfile).Username
	}

	panic("address not found")
}

// Display address from username
func DisplayAddressByUsername(username string) string {
	if addr, exists := usernameIndex.Get(username); exists {
		return addr.(string)
	}

	panic("username not found")
}

// Display avatarurl from address or username
func DisplayAvatarURL(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).AvatarURL
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).AvatarURL
			}
		}
	}

	panic("address or username not found")
}

// Display age from address or username
func DisplayAge(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).Age
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).Age
			}
		}
	}

	panic("address or username not found")
}

// Display gender from address or username
func DisplayGender(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).Gender
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).Gender
			}
		}
	}

	panic("address or username not found")
}

// Display website from address or username
func DisplayWebsite(address std.Address, username, websiteName string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			profile := value.(UserProfile)
			if url, exists := profile.Websites.Get(websiteName); exists {
				return url.(string)
			}
			panic("website not found")
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				profile := value.(UserProfile)
				if url, exists := profile.Websites.Get(websiteName); exists {
					return url.(string)
				}
				panic("website not found")
			}
		}
	}

	panic("address or username not found")
}

func DisplayAllWebsites(address std.Address, username string) map[string]string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			profile := value.(UserProfile)
			websites := make(map[string]string)
			profile.Websites.Iterate("", "", func(key string, value interface{}) bool {
				websites[key] = value.(string)
				return true
			})
			return websites
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				profile := value.(UserProfile)
				websites := make(map[string]string)
				profile.Websites.Iterate("", "", func(key string, value interface{}) bool {
					websites[key] = value.(string)
					return true
				})
				return websites
			}
		}
	}
	panic("address or username not found")
}

// Display country from address or username
func DisplayCountry(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).Country
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).Country
			}
		}
	}

	panic("address or username not found")
}

// Display city from address or username
func DisplayCity(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).City
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).City
			}
		}
	}

	panic("address or username not found")
}

// Display modification history specific to a field from address or username
func DisplayModificationHistory(address std.Address, username, field string) string {
	var value interface{}
	var exists bool

	if address != "" {
		value, exists = profiles.Get(address.String())
	} else if username != "" {
		if addr, ok := usernameIndex.Get(username); ok {
			value, exists = profiles.Get(addr.(string))
		}
	}

	if !exists {
		panic("address or username not found")
	}

	profile, ok := value.(UserProfile)
	if !ok || profile.ModificationHistory == nil {
		panic("no modification history available")
	}

	changes, found := profile.ModificationHistory[field]
	if !found {
		panic("no modifications recorded for the field")
	}

	var b bytes.Buffer
	for _, change := range changes {
		b.WriteString(ufmt.Sprintf("%s\n", change))
	}
	return b.String()
}

func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<html><body>")
	res.Write("<h1>Profiles List</h1><br>")
	usernameIndex.Iterate("", "", func(key string, value interface{}) bool {
		username := key
		res.Write("<p>Username: " + username + " [<a href='/r/demo/profile:info/" + username + "'>" + "View Profile</a>]</p>")
		return true
	})
	b.WriteString("</body></html>")
	res.Write(b.String())
}

func profileHandler(res *mux.ResponseWriter, req *mux.Request) {
	username := req.GetVar("username")
	profile := DisplayProfileByUsername(username)

	var b bytes.Buffer

	b.WriteString("<html><body>")
	// Reteturn home link
	b.WriteString("<p>[<a href='/r/demo/profile'>Home</a>]</p>")
	b.WriteString("<h1>Profile of " + profile.Username + "</h1>")
	b.WriteString("<p>Username: " + profile.Username + "</p>")
	b.WriteString("<p>Address: " + profile.Address.String() + "</p>")
	b.WriteString("<p>Avatar URL: <img src='" + profile.AvatarURL + "' /></p>")
	b.WriteString("<p>Age: " + profile.Age + "</p>")
	b.WriteString("<p>Gender: " + profile.Gender + "</p>")
	b.WriteString("<p>Country: " + profile.Country + "</p>")
	b.WriteString("<p>City: " + profile.City + "</p>")

	b.WriteString("<h2>Websites:</h2>")
	profile.Websites.Iterate("", "", func(key string, value interface{}) bool {
		b.WriteString("<p>" + key + ": <a href='" + value.(string) + "'>" + value.(string) + "</a></p>")
		return true
	})

	b.WriteString("<p>[<a href='/r/demo/profile:history/" + username + "'>View Modification History</a>]</p>")
	b.WriteString("</body></html>")

	res.Write(b.String())
}

func historyHandler(res *mux.ResponseWriter, req *mux.Request) {
	username := req.GetVar("username")
	profile := DisplayProfileByUsername(username)

	var b bytes.Buffer

	b.WriteString("<html><body>")

	b.WriteString("<p>[<a href='/r/demo/profile'>Home</a>] | [<a href='/r/demo/profile:info/" + username + "'>Profile " + username + "</a>]</p>")
	b.WriteString("<h1>Modification History for " + profile.Username + "</h1>")

	if profile.ModificationHistory != nil && len(profile.ModificationHistory) > 0 {
		for field, changes := range profile.ModificationHistory {
			b.WriteString("<h2>" + field + "</h2>")
			for _, change := range changes {
				b.WriteString("<p>" + change + "</p>")
			}
		}
	} else {
		b.WriteString("<p>No modifications recorded.</p>")
	}

	b.WriteString("</body></html>")
	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
