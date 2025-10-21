package types_test

import (
	"bytes"
	"testing"

	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/urequire"
)

func TestAppAcknowledgementUnmarshalJSON(t *testing.T) {
	var a types.AppAcknowledgement

	err := a.UnmarshalJSON([]byte(`{"response":{"result":"aGVsbG8="}}`))

	urequire.NoError(t, err)
	urequire.True(t, bytes.Equal([]byte("hello"), a.Response.Result))
	urequire.Equal(t, "", a.Response.Error)
	urequire.True(t, a.Success())

	err = a.UnmarshalJSON([]byte(`{"response":{"error":"oups"}}`))

	urequire.NoError(t, err)
	urequire.Equal(t, 0, len(a.Response.Result))
	urequire.Equal(t, "oups", a.Response.Error)
	urequire.False(t, a.Success(), "expect fail")
}
