package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/mux"
	"std"
	"time"
)

type UserProfile struct {
	Username            string
	Address             std.Address
	AvatarURL           string
	Age                 string
	Gender              string
	Website             string
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

func CreateProfile(username, avatarurl, age, gender, website, country, city string) {
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

	profile := UserProfile{
		Username:  username,
		Address:   caller,
		AvatarURL: avatarurl,
		Age:       age,
		Gender:    gender,
		Website:   website,
		Country:   country,
		City:      city,
	}

	profiles.Set(caller.String(), profile)

	
	usernameIndex.Set(username, caller.String())
	std.Emit("UserProfileCreated", "username", username, "address", caller.String())
}

func UpdateProfile(username, avatarurl, age, gender, website, country, city string) {
	caller := std.GetOrigCaller()
	value, exists := profiles.Get(caller.String())
	if !exists {
		panic("Profile not found")
	}

	// Retrieve the existing profile
	existingProfile := value.(UserProfile)

	// update the username and manage index if changed
	if username != "" && username != existingProfile.Username {
		_, exists := usernameIndex.Get(username)
		if exists {
			panic("Username is already taken")
		}

		// remove the old username from the index
		if existingProfile.Username != "" {
			usernameIndex.Remove(existingProfile.Username)
		}

		// set the new username in the index
		usernameIndex.Set(username, caller.String())
	}

	// Initialize the ModificationHistory map if it's nil
	if existingProfile.ModificationHistory == nil {
		existingProfile.ModificationHistory = make(map[string][]string)
	}

	// Function to update the profile field if changed
	updateField := func(field *string, newValue, fieldName string, profile *UserProfile) {
		if newValue != "" && *field != newValue {
			changeDescription := time.Now().Format(time.RFC3339) + ": Changed from " + *field + " to " + newValue
			if profile.ModificationHistory == nil {
				profile.ModificationHistory = make(map[string][]string)
			}
			profile.ModificationHistory[fieldName] = append(profile.ModificationHistory[fieldName], changeDescription)
			*field = newValue
		}
	}

	// Update fields and log changes
	updateField(&existingProfile.Username, username, "Username", &existingProfile)
	updateField(&existingProfile.AvatarURL, avatarurl, "AvatarURL", &existingProfile)
	updateField(&existingProfile.Age, age, "Age", &existingProfile)
	updateField(&existingProfile.Gender, gender, "Gender", &existingProfile)
	updateField(&existingProfile.Website, website, "Website", &existingProfile)
	updateField(&existingProfile.Country, country, "Country", &existingProfile)
	updateField(&existingProfile.City, city, "City", &existingProfile)

	// Update the profile in the tree
	profiles.Set(caller.String(), existingProfile)

	// Update the username index if the username is changed
	/*if username != existingProfile.Username {
		usernameIndex.Set(username, caller.String())
	}*/
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
func DisplayWebsite(address std.Address, username string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			return value.(UserProfile).Website
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				return value.(UserProfile).Website
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