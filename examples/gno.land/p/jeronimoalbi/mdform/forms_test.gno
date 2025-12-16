package mdform_test

import (
	"strings"
	"testing"

	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		attrs    []string
		markdown string
		err      string
	}{
		{
			name:     "ok",
			attrs:    []string{"exec", "FunctionName"},
			markdown: `<gno-form exec="FunctionName"></gno-form>`,
		},
		{
			name:     "no attributes",
			markdown: `<gno-form></gno-form>`,
		},
		{
			name:  "uneven attributes",
			attrs: []string{"exec"},
			err:   "expected an even number of attribute arguments",
		},
		{
			name:  "invalid attribute",
			attrs: []string{"foo", ""},
			err:   "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New(tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}

func TestFormInput(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		attrs     []string
		markdown  string
		err       string
	}{
		{
			name:      "ok",
			inputName: "test",
			attrs:     []string{"value", "foo"},
			markdown:  `<gno-form><gno-input name="test" value="foo" /></gno-form>`,
		},
		{
			name:      "no attributes",
			inputName: "test",
			markdown:  `<gno-form><gno-input name="test" /></gno-form>`,
		},
		{
			name:      "empty name",
			inputName: "  ",
			err:       "form input name is required",
		},
		{
			name:      "radio type",
			inputName: "test",
			attrs:     []string{"type", "radio"},
			err:       "use form.Radio() to create inputs of type radio",
		},
		{
			name:      "checkbox type",
			inputName: "test",
			attrs:     []string{"type", "checkbox"},
			err:       "use form.Checkbox() to create inputs of type checkbox",
		},
		{
			name:      "uneven attributes",
			inputName: "test",
			attrs:     []string{"value"},
			err:       "expected an even number of attribute arguments",
		},
		{
			name:      "invalid attribute",
			inputName: "test",
			attrs:     []string{"foo", ""},
			err:       "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New()
				f.Input(tc.inputName, tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form input to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form input to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}

func TestFormRadio(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		value     string
		attrs     []string
		markdown  string
		err       string
	}{
		{
			name:      "ok",
			inputName: "test",
			value:     "foo",
			attrs:     []string{"readonly", "true"},
			markdown:  `<gno-form><gno-input type="radio" name="test" value="foo" readonly="true" /></gno-form>`,
		},
		{
			name:      "no attributes",
			inputName: "test",
			value:     "foo",
			markdown:  `<gno-form><gno-input type="radio" name="test" value="foo" /></gno-form>`,
		},
		{
			name:      "empty name",
			inputName: "  ",
			err:       "form radio input name is required",
		},
		{
			name:      "ignore type attribute",
			inputName: "test",
			attrs:     []string{"type", "text"},
			markdown:  `<gno-form><gno-input type="radio" name="test" value="" /></gno-form>`,
		},
		{
			name:      "ignore value attribute",
			inputName: "test",
			attrs:     []string{"value", "foo"},
			markdown:  `<gno-form><gno-input type="radio" name="test" value="" /></gno-form>`,
		},
		{
			name:      "uneven attributes",
			inputName: "test",
			attrs:     []string{"readonly"},
			err:       "expected an even number of attribute arguments",
		},
		{
			name:      "invalid attribute",
			inputName: "test",
			attrs:     []string{"foo", ""},
			err:       "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New()
				f.Radio(tc.inputName, tc.value, tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form radio input to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form radio input to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}

func TestFormCheckbox(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		value     string
		attrs     []string
		markdown  string
		err       string
	}{
		{
			name:      "ok",
			inputName: "test",
			value:     "foo",
			attrs:     []string{"readonly", "true"},
			markdown:  `<gno-form><gno-input type="checkbox" name="test" value="foo" readonly="true" /></gno-form>`,
		},
		{
			name:      "no attributes",
			inputName: "test",
			value:     "foo",
			markdown:  `<gno-form><gno-input type="checkbox" name="test" value="foo" /></gno-form>`,
		},
		{
			name:      "empty name",
			inputName: "  ",
			err:       "form checkbox input name is required",
		},
		{
			name:      "ignore type attribute",
			inputName: "test",
			attrs:     []string{"type", "text"},
			markdown:  `<gno-form><gno-input type="checkbox" name="test" value="" /></gno-form>`,
		},
		{
			name:      "ignore value attribute",
			inputName: "test",
			attrs:     []string{"value", "foo"},
			markdown:  `<gno-form><gno-input type="checkbox" name="test" value="" /></gno-form>`,
		},
		{
			name:      "uneven attributes",
			inputName: "test",
			attrs:     []string{"readonly"},
			err:       "expected an even number of attribute arguments",
		},
		{
			name:      "invalid attribute",
			inputName: "test",
			attrs:     []string{"foo", ""},
			err:       "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New()
				f.Checkbox(tc.inputName, tc.value, tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form checkbox input to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form checkbox input to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}

func TestFormTextarea(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		attrs     []string
		markdown  string
		err       string
	}{
		{
			name:      "ok",
			inputName: "test",
			attrs:     []string{"value", "foo"},
			markdown:  `<gno-form><gno-textarea name="test" value="foo" /></gno-form>`,
		},
		{
			name:      "no attributes",
			inputName: "test",
			markdown:  `<gno-form><gno-textarea name="test" /></gno-form>`,
		},
		{
			name:      "empty name",
			inputName: "  ",
			err:       "form textarea name is required",
		},
		{
			name:      "uneven attributes",
			inputName: "test",
			attrs:     []string{"value"},
			err:       "expected an even number of attribute arguments",
		},
		{
			name:      "invalid attribute",
			inputName: "test",
			attrs:     []string{"foo", ""},
			err:       "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New()
				f.Textarea(tc.inputName, tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form textarea to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form textarea to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}

func TestFormSelect(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		value     string
		attrs     []string
		markdown  string
		err       string
	}{
		{
			name:      "ok",
			inputName: "test",
			value:     "foo",
			attrs:     []string{"readonly", "true"},
			markdown:  `<gno-form><gno-select name="test" value="foo" readonly="true" /></gno-form>`,
		},
		{
			name:      "no attributes",
			inputName: "test",
			value:     "foo",
			markdown:  `<gno-form><gno-select name="test" value="foo" /></gno-form>`,
		},
		{
			name:      "empty name",
			inputName: "  ",
			err:       "form select name is required",
		},
		{
			name:      "uneven attributes",
			inputName: "test",
			attrs:     []string{"value"},
			err:       "expected an even number of attribute arguments",
		},
		{
			name:      "invalid attribute",
			inputName: "test",
			attrs:     []string{"foo", ""},
			err:       "invalid attribute: foo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var markdown string
			fn := func() {
				f := mdform.New()
				f.Select(tc.inputName, tc.value, tc.attrs...)
				markdown = strings.ReplaceAll(f.String(), "\n", "")
			}

			if tc.err != "" {
				urequire.PanicsWithMessage(t, tc.err, fn, "expect form textarea to fail")
				return
			}

			urequire.NotPanics(t, fn, "expect form textarea to be created")
			uassert.Equal(t, tc.markdown, markdown, "expected markdown to match")
		})
	}
}
