package forms

import (
	"strings"

	"gno.land/p/onbloc/json"
)

func validateBooleanField(node *json.Node, field Field) bool {
	if node.IsBool() == false {
		return false
	}

	answer, err := node.GetBool()
	if err != nil {
		return false
	}

	// If the field is required, checkbox must be checked
	if field.Required == true && answer == false {
		return false
	}

	return true
}

func validateStringField(node *json.Node, field Field) bool {
	if node.IsString() == false {
		return false
	}

	answer, err := node.GetString()
	if err != nil {
		return false
	}

	// If the field is required, the answer must not be empty
	if field.Required == true && answer == "" {
		return false
	}

	return true
}

func validateNumberField(node *json.Node, field Field) bool {
	if node.IsNumber() == false {
		return false
	}

	_, err := node.GetNumeric()
	if err != nil {
		return false
	}

	return true
}

func validateMultiChoiceField(node *json.Node, field Field) bool {
	choices := strings.Split(field.FieldType[1:len(field.FieldType)-1], "|")

	if node.IsArray() == false {
		return false
	}

	if field.Required == true && node.Size() == 0 {
		return false
	}

	if node.Size() > len(choices) {
		return false
	}

	for i := 0; i < node.Size(); i++ {
		choiceNode, err := node.GetIndex(i)
		if err != nil {
			return false
		}

		choiceIdx := choiceNode.MustNumeric()
		if choiceIdx < 0 || int(choiceIdx) >= len(choices) {
			return false
		}
	}

	return true
}

func validateChoiceField(node *json.Node, field Field) bool {
	choices := strings.Split(field.FieldType[1:len(field.FieldType)-1], "|")

	if node.IsNumber() == false {
		return false
	}

	choiceIdx := node.MustNumeric()
	if choiceIdx < 0 || int(choiceIdx) >= len(choices) {
		return false
	}

	return true
}

func ValidateAnswer(answer *json.Node, field Field) bool {
	if field.FieldType == "boolean" {
		return validateBooleanField(answer, field)
	} else if field.FieldType == "string" {
		return validateStringField(answer, field)
	} else if field.FieldType == "number" {
		return validateNumberField(answer, field)
	} else if strings.HasPrefix(field.FieldType, "{") && strings.HasSuffix(field.FieldType, "}") {
		return validateMultiChoiceField(answer, field)
	} else if strings.HasPrefix(field.FieldType, "[") && strings.HasSuffix(field.FieldType, "]") {
		return validateChoiceField(answer, field)
	}

	return false
}

// ValidateAnswers checks if the given answers are valid for the given fields
func ValidateAnswers(answers string, fields []Field) bool {
	unmarshalled, err := json.Unmarshal([]byte(answers))
	if err != nil {
		return false
	}

	// If the number of answers is different from the number of fields, it's invalid
	if len(fields) != unmarshalled.Size() {
		return false
	}

	for i, field := range fields {
		answer, err := unmarshalled.GetIndex(i)
		if err != nil {
			return false
		}

		// If the answer is empty and the field is not required, it's valid
		if answer.IsNull() && !field.Required {
			return true
		}

		if !ValidateAnswer(answer, field) {
			return false
		}
	}

	return true
}
