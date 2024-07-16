package forms

import (
	"strings"

	"gno.land/p/demo/json"
)

func ValidateBooleanField(node *json.Node, field Field) bool {
	if node.IsBool() == false {
		return false
	}

	answer, err := node.GetBool()
	if err != nil {
		return false
	}

	if field.Required == true && answer == false {
		return false
	}

	return true
}

func ValidateStringField(node *json.Node, field Field) bool {
	if node.IsString() == false {
		return false
	}

	answer, err := node.GetString()
	if err != nil {
		return false
	}

	if field.Required == true && answer == "" {
		return false
	}

	return true
}

func ValidateNumberField(node *json.Node, field Field) bool {
	if node.IsNumber() == false {
		return false
	}

	answer, err := node.GetNumeric()
	if err != nil {
		return false
	}

	if field.Required == true && answer == 0 {
		return false
	}

	return true
}

func ValidateMultiChoiceField(node *json.Node, field Field) bool {
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

func ValideChoiceField(node *json.Node, field Field) bool {
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

func ValidateAnswers(answers string, fields []Field) bool {
	unmarshalled, err := json.Unmarshal([]byte(answers))
	if err != nil {
		return false
	}

	if len(fields) != unmarshalled.Size() {
		return false
	}

	for i, field := range fields {
		answer, err := unmarshalled.GetIndex(i)
		if err != nil {
			return false
		}
		if field.FieldType == "boolean" {
			if ValidateBooleanField(answer, field) == false {
				return false
			}
		}
		if field.FieldType == "string" {
			if ValidateStringField(answer, field) == false {
				return false
			}
		}
		if field.FieldType == "number" {
			if ValidateNumberField(answer, field) == false {
				return false
			}
		}
		if field.FieldType[0] == '{' && field.FieldType[len(field.FieldType)-1] == '}' {
			if ValidateMultiChoiceField(answer, field) == false {
				return false
			}
		}
		if field.FieldType[0] == '[' && field.FieldType[len(field.FieldType)-1] == ']' {
			if ValideChoiceField(answer, field) == false {
				return false
			}
		}
	}

	return true
}
