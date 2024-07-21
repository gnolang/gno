package forms

import (
	"std"
	"time"

	"gno.land/p/demo/json"
)

func CreateField(label string, fieldType string, required bool) Field {
	return Field{
		Label:     label,
		FieldType: fieldType,
		Required:  required,
	}
}

type times struct {
	openAt  *time.Time
	closeAt *time.Time
}

func parseDates(openAt string, closeAt string) (times, error) {
	var openAtTime, closeAtTime *time.Time

	dateFormat := "2006-01-02T15:04:05Z"

	// Parse openAt if it's not empty
	if openAt != "" {
		res, err := time.Parse(dateFormat, openAt)
		if err != nil {
			return times{}, errInvalidDate
		}
		openAtTime = &res
	}

	// Parse closeAt if it's not empty
	if closeAt != "" {
		res, err := time.Parse(dateFormat, closeAt)
		if err != nil {
			return times{}, errInvalidDate
		}
		closeAtTime = &res
	}

	return times{openAtTime, closeAtTime}, nil
}

func (db *FormDB) CreateForm(title string, description string, openAt string, closeAt string, data string) (string, error) {
	// Parsing dates
	times, err := parseDates(openAt, closeAt)
	if err != nil {
		return "", err
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
		field, err := node.GetIndex(i)
		if err != nil {
			return "", errInvalidJson
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
		openAt:      times.openAt,
		closeAt:     times.closeAt,
		Fields:      fields,
	}

	// Adding the form to the database
	db.Forms = append(db.Forms, &form)

	return id, nil
}
