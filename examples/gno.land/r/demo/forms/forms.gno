package forms

import (
	"gno.land/p/demo/forms" 
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/json"
)

var db *forms.FormDatabase

func init() {
	db = forms.NewDatabase()
}

func CreateForm(title string, description string, openAt string, closeAt string, data string) string {
	return db.CreateForm(title, description, openAt, closeAt, data)
}

func GetForms() string {
	formsJson := json.ArrayNode("", []*json.Node{})
	for _, form := range db.Forms {
		fieldsJson := json.ArrayNode("", []*json.Node{})
		for _, field := range form.Fields {
			fieldJson := json.ObjectNode("", map[string]*json.Node{
				"label": json.StringNode("label", field.Label),
				"fieldType": json.StringNode("fieldType", field.FieldType),
				"required": json.BoolNode("required", field.Required),
			})
			fieldsJson.AppendArray(fieldJson)
		} 

		formJson := json.ObjectNode("", map[string]*json.Node{
			"id": json.StringNode("id", form.ID),
			"owner": json.StringNode("owner", form.Owner.String()),
			"title": json.StringNode("title", form.Title),
			"description": json.StringNode("description", form.Description),
			"createdAt": json.StringNode("createdAt", form.CreatedAt),
			"openAt": json.StringNode("openAt", form.OpenAt),
			"closeAt": json.StringNode("closeAt", form.CloseAt),
			"fields": fieldsJson,
		})
		formsJson.AppendArray(formJson)
	}

	encoded, err := json.Marshal(formsJson)
	if err != nil {
		panic(ufmt.Errorf("error: %v", err))
	}

	return string(encoded)
}

func GetFormByID(id string) string {
	form := db.GetForm(id)
	if form == nil {
		panic(ufmt.Errorf("Form not found: %s", id))
	}

	fieldsJson := json.ArrayNode("", []*json.Node{})
	for _, field := range form.Fields {
		fieldJson := json.ObjectNode("", map[string]*json.Node{
			"label": json.StringNode("label", field.Label),
			"fieldType": json.StringNode("fieldType", field.FieldType),
			"required": json.BoolNode("required", field.Required),
		})
		fieldsJson.AppendArray(fieldJson)
	}

	encoded, _ := json.Marshal(fieldsJson)
	ufmt.Println(string(encoded))
	ufmt.Println("DWADAWDWA")

	jsonRes := json.ObjectNode("", map[string]*json.Node{
		"id": json.StringNode("id", form.ID),
		"owner": json.StringNode("owner", form.Owner.String()),
		"title": json.StringNode("title", form.Title),
		"description": json.StringNode("description", form.Description),
		"createdAt": json.StringNode("createdAt", form.CreatedAt),
		"openAt": json.StringNode("openAt", form.OpenAt),
		"closeAt": json.StringNode("closeAt", form.CloseAt),
		"fields": fieldsJson,
	})	

	encoded, err := json.Marshal(jsonRes)
	if err != nil {
		panic(ufmt.Errorf("error: %v", err))
	}

	return string(encoded)
}

func GetAnswer(formID string, authorID string) string {
	form := db.GetForm(formID)
	if form == nil {
		panic(ufmt.Errorf("Form not found: %s", formID))
	}

	answer := db.GetAnswer(formID, authorID)
	if answer == nil {
		panic(ufmt.Errorf("Answer not found: %s", authorID))
	}

	return answer.Answers
}

func SubmitForm(formID string, answers string) {
	form := db.GetForm(formID)
	if form == nil {
		panic(ufmt.Errorf("Form not found: %s", formID))
	}
	
	db.SubmitForm(formID, answers)
}

func Render(path string) string {
	response := "Forms:\n\n"
	for _, form := range db.Forms {
		response += ufmt.Sprintf("- %s\n\n", GetFormByID(form.ID))
	}
	response += "Answers:\n\n"
	for _, answer := range db.Answers {
		response += ufmt.Sprintf("- Form ID: %s\nAuthor: %s\nSubmitted At: %s\n>Answers: %s\n\n", answer.FormID, answer.Author, answer.SubmittedAt, answer.Answers)	
	}

	return response
}