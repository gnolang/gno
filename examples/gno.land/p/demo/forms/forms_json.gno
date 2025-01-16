package forms

import (
	"std"
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

func (b *json.NodeBuilder) WriteFormFields(key string, fields []Field) *json.NodeBuilder {
	b.WriteArray(key, func(builder *json.ArrayBuilder) {
		for _, field := range fields {
			builder.WriteObject(func(builder *json.NodeBuilder) {
				builder.WriteString("label", field.Label).
					WriteString("fieldType", field.FieldType).
					WriteBool("required", field.Required)
			})
		}
	})
	return b
}

func (b *json.NodeBuilder) WriteFormSubmission(key string, submission *Submission) *json.NodeBuilder {
	b.WriteObject(key, func(builder *json.NodeBuilder) {
		builder.WriteString("submittedAt", submission.SubmittedAt.Format("2006-01-02 15:04:05")).
			WriteString("answers", strings.ReplaceAll(submission.Answers, "\"", "'"))
	})
	return b
}

func (b *json.NodeBuilder) WriteForm(key string, value *Form) *json.NodeBuilder {
	b.WriteString("id", value.ID).
		WriteString("owner", value.Owner.String()).
		WriteString("title", value.Title).
		WriteString("description", value.Description).
		WriteString("createdAt", value.CreatedAt.Format("2006-01-02 15:04:05")).
		WriteFormFields("fields", value.Fields)
	return b
}

func (b *json.ArrayBuilder) WriteForm(key string, value *Form) *json.ArrayBuilder {
	b.WriteObject(func(builder *json.NodeBuilder) {
		builder.WriteString("id", value.ID).
			WriteString("owner", value.Owner.String()).
			WriteString("title", value.Title).
			WriteString("description", value.Description).
			WriteString("createdAt", value.CreatedAt.Format("2006-01-02 15:04:05")).
			WriteFormFields("fields", value.Fields)
	})
	return b
}
