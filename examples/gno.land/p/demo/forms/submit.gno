package forms

import ( 
	"time" 
	"std"
	
	"gno.land/p/demo/ufmt"
)

func (db *FormDatabase) SubmitForm(formID string, answers string) {
	form := db.GetForm(formID)
	if form == nil {
		panic(ufmt.Errorf("Form not found: %s", formID))
	}

	if ValidateAnswers(answers, form.Fields) == false {
		panic("Invalid answers")
	}

	answer := &Submission{
		FormID: formID,
		Answers: answers,
		Author: std.PrevRealm().Addr(),
		SubmittedAt: time.Now().String(),
	}

	db.Answers = append(db.Answers, answer)
}

func (db *FormDatabase) GetAnswer(formID string, authorID string) *Submission {
	for _, answer := range db.Answers {
		if answer.FormID == formID && answer.Author.String() == authorID {
			return answer
		}
	}
	return nil
}