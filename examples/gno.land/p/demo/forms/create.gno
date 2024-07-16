package forms

import ( 
	"time" 
	"std"
	
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/json" 
)

func (form *Form) AddField(label string, fieldType string, required bool) {
	field := Field{
		Label: label,
		FieldType: fieldType,
		Required: required,
	}

	form.Fields = append(form.Fields, field)
}

func (db *FormDatabase) GetForm(id string) *Form {
	for _, form := range db.Forms {
		if form.ID == id {
			return form
		}
	}
	return nil
}

func (db *FormDatabase) CreateForm(title string, description string, openAt string, closeAt string, data string) string {
	id := db.IDCounter.Next().String()

	form := &Form{
		ID: id,
		Owner: std.PrevRealm().Addr(),
		Title: title,
		Description: description,
		CreatedAt: time.Now().String(),
		OpenAt: openAt,
		CloseAt: closeAt,
	}

	db.Forms = append(db.Forms, form)

    node, err := json.Unmarshal([]byte(data))
    if err != nil {
        ufmt.Errorf("error: %v", err)
    } 

	for i := 0; i < node.Size(); i++ {
		field, err := node.GetIndex(i)

		if err != nil {
			ufmt.Errorf("error: %v", err)
		}

		labelNode, _ := field.GetKey("label")
		fieldTypeNode, _ := field.GetKey("fieldType")
		requiredNode, _ := field.GetKey("required")

		label, _ := labelNode.GetString()
		fieldType, _ := fieldTypeNode.GetString()
		required, _ := requiredNode.GetBool()

		form.AddField(label, fieldType, required)
	}

	return id	
}