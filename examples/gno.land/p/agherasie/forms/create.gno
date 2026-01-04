package forms

import (
	"chain/runtime"
	"time"

	"gno.land/p/onbloc/json"
)

const dateFormat = "2006-01-02T15:04:05Z"

func CreateField(label string, fieldType string, required bool) Field {
	return Field{
		Label:     label,
		FieldType: fieldType,
		Required:  required,
	}
}

// CreateForm creates a new form with the given parameters
func (db *FormDB) CreateForm(title string, description string, openAt string, closeAt string, data string) (string, error) {
	// Parsing the dates
	var parsedOpenTime, parsedCloseTime time.Time

	if openAt != "" {
		var err error
		parsedOpenTime, err = time.Parse(dateFormat, openAt)
		if err != nil {
			return "", errInvalidDate
		}
	}

	if closeAt != "" {
		var err error
		parsedCloseTime, err = time.Parse(dateFormat, closeAt)
		if err != nil {
			return "", errInvalidDate
		}
	}

	// Parsing the json submission
	node, err := json.Unmarshal([]byte(data))
	if err != nil {
		return "", errInvalidJson
	}

	fieldsCount := node.Size()
	fields := make([]Field, fieldsCount)

	// Parsing the json submission to create the gno data structures
	for i := 0; i < fieldsCount; i++ {
		field := node.MustIndex(i)

		fields[i] = CreateField(
			field.MustKey("label").MustString(),
			field.MustKey("fieldType").MustString(),
			field.MustKey("required").MustBool(),
		)
	}

	// Generating the form ID
	id := db.IDCounter.Next().String()

	// Creating the form
	form := Form{
		ID:          id,
		Owner:       runtime.CurrentRealm().Address(),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		openAt:      parsedOpenTime,
		closeAt:     parsedCloseTime,
		Fields:      fields,
	}

	// Adding the form to the database
	db.Forms = append(db.Forms, &form)

	return id, nil
}
