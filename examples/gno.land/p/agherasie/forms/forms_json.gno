package forms

import (
	"strings"

	"gno.land/p/onbloc/json"
)

type FormNodeBuilder struct {
	*json.NodeBuilder
}

type FormArrayBuilder struct {
	*json.ArrayBuilder
}

func (b *FormNodeBuilder) WriteArray(key string, fn func(*FormArrayBuilder)) *FormNodeBuilder {
	b.NodeBuilder.WriteArray(key, func(ab *json.ArrayBuilder) {
		fn(&FormArrayBuilder{ab})
	})
	return b
}

func (b *FormNodeBuilder) WriteObject(key string, fn func(*FormNodeBuilder)) *FormNodeBuilder {
	b.NodeBuilder.WriteObject(key, func(nb *json.NodeBuilder) {
		fn(&FormNodeBuilder{nb})
	})
	return b
}

func (b *FormArrayBuilder) WriteObject(fn func(*FormNodeBuilder)) *FormArrayBuilder {
	b.ArrayBuilder.WriteObject(func(nb *json.NodeBuilder) {
		fn(&FormNodeBuilder{nb})
	})
	return b
}

func (b *FormNodeBuilder) WriteFormFields(key string, fields []Field) *FormNodeBuilder {
	b.WriteArray(key, func(builder *FormArrayBuilder) {
		for _, field := range fields {
			builder.WriteObject(func(builder *FormNodeBuilder) {
				builder.WriteString("label", field.Label).
					WriteString("fieldType", field.FieldType).
					WriteBool("required", field.Required)
			})
		}
	})
	return b
}

func (b *FormNodeBuilder) WriteFormSubmission(key string, submission *Submission) *FormNodeBuilder {
	b.WriteObject(key, func(builder *FormNodeBuilder) {
		builder.WriteString("submittedAt", submission.SubmittedAt.Format("2006-01-02 15:04:05")).
			WriteString("answers", strings.ReplaceAll(submission.Answers, "\"", "'"))
	})
	return b
}

func (b *FormNodeBuilder) WriteForm(key string, value *Form) *FormNodeBuilder {
	b.WriteString("id", value.ID).
		WriteString("owner", value.Owner.String()).
		WriteString("title", value.Title).
		WriteString("description", value.Description).
		WriteString("createdAt", value.CreatedAt.Format("2006-01-02 15:04:05"))
	b.WriteFormFields("fields", value.Fields)
	return b
}

func (b *FormArrayBuilder) WriteForm(key string, value *Form) *FormArrayBuilder {
	b.WriteObject(func(builder *FormNodeBuilder) {
		builder.WriteString("id", value.ID).
			WriteString("owner", value.Owner.String()).
			WriteString("title", value.Title).
			WriteString("description", value.Description).
			WriteString("createdAt", value.CreatedAt.Format("2006-01-02 15:04:05"))
		builder.WriteFormFields("fields", value.Fields)
	})
	return b
}
