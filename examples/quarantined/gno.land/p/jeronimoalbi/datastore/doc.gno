// Package datastore provides support to store multiple collections of records.
//
// It supports the definition of multiple storages, where each one is a collection
// of records. Records can have any number of user defined fields which are added
// dynamically when values are set on a record. These fields can also be renamed
// or removed.
//
// Storages have support for simple schemas that allows users to pre-define fields
// which can optionally have a default value also defined. Default values are
// assigned to new records on creation.
//
// User defined schemas can optionally be strict, which means that records from a
// storage using the schema can only assign values to the pre-defined set of fields.
// In which case, assigning a value to an unknown field would result on an error.
//
// Package also support the definition of custom record indexes. Indexes are used
// by storages to search and iterate records.
// The default index is the ID index but custom single and multi value indexes can
// be defined.
//
// WARNING: Using this package to store your realm data must be carefully considered.
// The fact that record fields are not strictly typed and can be renamed or removed
// could lead to issues if not careful when coding your realm(s). So it's recommended
// that you consider other alternatives first, like alternative patterns or solutions
// provided by the blockchain to deal with data, types and data migration for example.
//
// Example usage:
//
//	var db datastore.Datastore
//
//	// Define a unique case insensitive index for user emails
//	emailIdx := datastore.NewIndex("email", func(r datastore.Record) string {
//	  return r.MustGet("email").(string)
//	}).Unique().CaseInsensitive()
//
//	// Create a new storage for user records
//	storage := db.CreateStorage("users", datastore.WithIndex(emailIdx))
//
//	// Add a user with a single "email" field
//	user := storage.NewRecord()
//	user.Set("email", "foo@bar.org")
//
//	// Save to assing user ID and update indexes
//	user.Save()
//
//	// Find user by email using the custom index
//	user, _ = storage.Get(emailIdx.Name(), "foo@bar.org")
//
//	// Find user by ID
//	user, _ = storage.GetByID(user.ID())
//
//	// Search user's profile by email in another existing storage
//	storage = db.GetStorage("profiles")
//	email := user.MustGet("email").(string)
//	profile, found := storage.Get("user", email)
//	if !found {
//	  panic("profile not found")
//	}
//
//	// Delete the profile from the storage and update indexes
//	storage.Delete(profile.ID())
//
// Example query usage:
//
//	var db datastore.Datastore
//
//	// Create a query with a custom offset and size
//	storage := db.GetStorage("users")
//	recordset, err := storage.Query(datastore.WithOffset(100), datastore.WithSize(50))
//	if err != nil {
//	  panic(err)
//	}
//
//	// Get all query results
//	var records []Record
//	recordset.Iterate(func(r datastore.Record) bool {
//	  records = append(records, r)
//	  return false
//	})
//
// Example query using a custom index usage:
//
//	var db datastore.Datastore
//
//	// Create a query to get records using a custom pre-defined index
//	storage := db.GetStorage("posts")
//	recordset, err := storage.Query(datastore.UseIndex("tags", "tagname"))
//	if err != nil {
//	  panic(err)
//	}
//
//	// Get all query results
//	var records []Record
//	recordset.Iterate(func(r datastore.Record) bool {
//	  records = append(records, r)
//	  return false
//	})
package datastore
