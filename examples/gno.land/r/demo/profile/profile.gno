package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
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

var profiles *avl.Tree
var usernameIndex *avl.Tree // Index to quickly find a profile by username

func init() {
	profiles = avl.NewTree()
	usernameIndex = avl.NewTree()
}

func CreateProfile(username, avatarurl, age, gender, website, country, city string) {
	caller := std.GetOrigCaller()

	// Check if a profile already exists for this address
	if _, exists := profiles.Get(caller.String()); exists {
		panic("A profile already exists for this address")
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

	// Add the username to the index
	if username != "" {
		usernameIndex.Set(username, caller.String())
	}
}

func UpdateProfile(username, avatarurl, age, gender, website, country, city string) {
	caller := std.GetOrigCaller()
	value, exists := profiles.Get(caller.String())
	if !exists {
		panic("Profile not found")
	}

	// Retrieve the existing profile
	existingProfile := value.(UserProfile)

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
	if username != "" {
		usernameIndex.Set(username, caller.String())
	}
}

// Display complete Profile by Address
func DisplayProfileByAddress(address std.Address) (UserProfile, bool) {
	if value, exists := profiles.Get(address.String()); exists {
		return value.(UserProfile), true
	}
	return UserProfile{}, false
}

// Get complete profile by username
func DisplayProfileByUsername(username string) (UserProfile, bool) {
	if addr, exists := usernameIndex.Get(username); exists {
		return DisplayProfileByAddress(std.Address(addr.(string)))
	}
	return UserProfile{}, false
}

// Display username
func DisplayUsername(address std.Address) string {
	if value, exists := profiles.Get(address.String()); exists {
		return value.(UserProfile).Username
	}
	return "Address not found"
}

// Display address from username
func DisplayAddress(username string) string {
	if addr, exists := usernameIndex.Get(username); exists {
		return addr.(string)
	}
	return "Username not found"
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
	return "Address or Username not found"
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
	return "Address or Username not found"
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
	return "Address or Username not found"
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
	return "Address or Username not found"
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
	return "Address or Username not found"
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
	return "Address or Username not found"
}

// Display modification history specific to a field from address or username
func DisplayModificationHistory(address std.Address, username, field string) string {
	if address != "" {
		if value, exists := profiles.Get(address.String()); exists {
			if value.(UserProfile).ModificationHistory != nil {
				if changes, exists := value.(UserProfile).ModificationHistory[field]; exists {
					var b bytes.Buffer
					for _, change := range changes {
						b.WriteString(ufmt.Sprintf("%s\n", change))
					}
					return b.String()
				}
			}
		}
	} else if username != "" {
		if addr, exists := usernameIndex.Get(username); exists {
			if value, exists := profiles.Get(addr.(string)); exists {
				if value.(UserProfile).ModificationHistory != nil {
					if changes, exists := value.(UserProfile).ModificationHistory[field]; exists {
						var b bytes.Buffer
						for _, change := range changes {
							b.WriteString(ufmt.Sprintf("%s\n", change))
						}
						return b.String()
					}
				}
			}
		}
	}
	return "Address or Username not found or no modifications recorded for the field"
}

func Render(path string) string {
	if path == "" {
		return renderHome()
	}

	return "unknown path"
}

func renderHome() string {
	var b bytes.Buffer

	profiles.Iterate("", "", func(key string, value interface{}) bool {
		profile := value.(UserProfile)
		b.WriteString(ufmt.Sprintf("Username: %s\n\n", profile.Username))
		b.WriteString(ufmt.Sprintf("Address: %s\n\n", profile.Address))
		b.WriteString(ufmt.Sprintf("AvatarURL: %s\n\n", profile.AvatarURL))
		b.WriteString(ufmt.Sprintf("Age: %s\n\n", profile.Age))
		b.WriteString(ufmt.Sprintf("Gender: %s\n\n", profile.Gender))
		b.WriteString(ufmt.Sprintf("Website: %s\n\n", profile.Website))
		b.WriteString(ufmt.Sprintf("Country: %s\n\n", profile.Country))
		b.WriteString(ufmt.Sprintf("City: %s\n\n", profile.City))

		// Properly iterate over the modification history map
		if profile.ModificationHistory != nil {
			b.WriteString("ModificationHistory:\n\n")
			for key, changes := range profile.ModificationHistory {
				b.WriteString(ufmt.Sprintf("%s:\n", key))
				for _, change := range changes {
					b.WriteString(ufmt.Sprintf("  %s | \n", change))
				}
			}
		} else {
			b.WriteString("No modifications recorded.\n")
		}

		b.WriteString("\n")

		return false
	})

	return b.String()
}
