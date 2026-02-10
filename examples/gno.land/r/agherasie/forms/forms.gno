package forms

import (
	"gno.land/p/agherasie/forms"
	"gno.land/p/nt/ufmt"
	"gno.land/p/onbloc/json"
	"gno.land/r/leon/hor"
)

var db *forms.FormDB

func init() {
	hor.Register(cross, "Ahgerasle's forms", "")
	db = forms.NewDB()
}

func CreateForm(_ realm, title string, description string, openAt string, closeAt string, data string) string {
	id, err := db.CreateForm(title, description, openAt, closeAt, data)
	if err != nil {
		panic(err)
	}
	return id
}

func GetForms() string {
	builder := forms.FormNodeBuilder{json.Builder()}

	builder.WriteArray("forms", func(builder *forms.FormArrayBuilder) {
		for _, form := range db.Forms {
			builder.WriteObject(func(builder *forms.FormNodeBuilder) {
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

	builder := forms.FormNodeBuilder{json.Builder()}

	builder.WriteForm("form", form).
		WriteObject("submissions", func(builder *forms.FormNodeBuilder) {
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

	answer, err := db.GetAnswer(formID, address(authorID))
	if answer != nil {
		panic(err)
	}

	return answer.Answers
}

func SubmitForm(_ realm, formID string, answers string) {
	_, err := db.GetForm(formID)
	if err != nil {
		panic(err)
	}

	db.SubmitForm(formID, answers)
}

func Render(_ string) string {
	if len(db.Forms) == 0 {
		response := "No forms yet !"
		return response
	}

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
