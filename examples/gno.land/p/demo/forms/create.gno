package forms

import (
	"std"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

func CreateField(label string, fieldType string, required bool) Field {
	return Field{
		Label:     label,
		FieldType: fieldType,
		Required:  required,
	}
}

func parseDates(openAt string, closeAt string) (*time.Time, *time.Time) {
	var openAtTime, closeAtTime *time.Time

	dateFormat := "2006-01-02T15:04:05Z"

	// Parse openAt if it's not empty
	if openAt != "" {
		res, err := time.Parse(dateFormat, openAt)
		if err != nil {
			panic(ufmt.Errorf("invalid date: %v", openAt))
		}
		openAtTime = &res
	}

	// Parse closeAt if it's not empty
	if closeAt != "" {
		res, err := time.Parse(dateFormat, closeAt)
		if err != nil {
			panic(ufmt.Errorf("invalid date: %v", closeAt))
		}
		closeAtTime = &res
	}

	return openAtTime, closeAtTime
}

func (db *FormDB) CreateForm(title string, description string, openAt string, closeAt string, data string) string {
	// Parsing dates
	openAtTime, closeAtTime := parseDates(openAt, closeAt)

	// Parsing the json submission
	node, err := json.Unmarshal([]byte(data))
	if err != nil {
		ufmt.Errorf("invalid json: %v", err)
	}

	fieldsCount := node.Size()
	fields := make([]Field, fieldsCount)

	// Parsing the json submission to create the gno data structures
	for i := 0; i < fieldsCount; i++ {
		field, err := node.GetIndex(i)
		if err != nil {
			ufmt.Errorf("error: %v", err)
		}

		labelNode, _ := field.GetKey("label")
		fieldTypeNode, _ := field.GetKey("fieldType")
		requiredNode, _ := field.GetKey("required")

		label, _ := labelNode.GetString()
		fieldType, _ := fieldTypeNode.GetString()
		required, _ := requiredNode.GetBool()

		fields[i] = CreateField(label, fieldType, required)
	}

	// Generating the form ID
	id := db.IDCounter.Next().String()

	// Creating the form
	form := Form{
		ID:          id,
		Owner:       std.PrevRealm().Addr(),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		openAt:      openAtTime,
		closeAt:     closeAtTime,
		Fields:      fields,
	}

	// Adding the form to the database
	db.Forms = append(db.Forms, &form)

	return id
}
