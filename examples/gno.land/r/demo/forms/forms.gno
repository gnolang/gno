package forms

import (
	"std"

	"gno.land/p/demo/forms"
	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

var db *forms.FormDB

func init() {
	db = forms.NewDB()
}

func CreateForm(title string, description string, openAt string, closeAt string, data string) string {
	id, err := db.CreateForm(title, description, openAt, closeAt, data)
	if err != nil {
		panic(err)
	}
	return id
}

func GetForms() string {
	builder := json.Builder()

	builder.WriteArray("forms", func(builder *json.ArrayBuilder) {
		for _, form := range db.Forms {
			builder.WriteObject(func(builder *json.NodeBuilder) {
				builder.WriteForm("form", form)
			})
		}
	})

	encoded, err := json.Marshal(builder.Node())
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func GetFormByID(id string) string {
	form, err := db.GetForm(id)
	if err != nil {
		panic(err)
	}

	builder := json.Builder()

	builder.WriteForm("form", form).
		WriteObject("submissions", func(builder *json.NodeBuilder) {
			formSubmissions := db.GetSubmissionsByFormID(form.ID)
			for _, submission := range formSubmissions {
				builder.WriteFormSubmission(submission.Author.String(), submission)
			}
		})

	openAt, err := form.OpenAt()
	if err == nil {
		builder.WriteString("openAt", openAt.Format("2006-01-02 15:04:05"))
	}
	closeAt, err := form.CloseAt()
	if err == nil {
		builder.WriteString("closeAt", closeAt.Format("2006-01-02 15:04:05"))
	}

	encoded, err := json.Marshal(builder.Node())
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func GetAnswer(formID string, authorID string) string {
	_, err := db.GetForm(formID)
	if err != nil {
		panic(err)
	}

	answer, err := db.GetAnswer(formID, std.Address(authorID))
	if answer != nil {
		panic(err)
	}

	return answer.Answers
}

func SubmitForm(formID string, answers string) {
	_, err := db.GetForm(formID)
	if err != nil {
		panic(err)
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
