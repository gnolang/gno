package gnolang

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectIDMarshalAmino(t *testing.T) {
	var pid PkgID
	for i := range pid.Hashlet {
		pid.Hashlet[i] = byte(i*37 + 1)
	}
	for _, nt := range []uint64{0, 1, 9, 10, 12345, math.MaxUint32, math.MaxInt64, math.MaxUint64} {
		oid := ObjectID{PkgID: pid, NewTime: nt}
		got, err := oid.MarshalAmino()
		require.NoError(t, err)
		// The format the store keys and every stored object depend on.
		assert.Equal(t, fmt.Sprintf("%s:%d", hex.EncodeToString(pid.Hashlet[:]), nt), got)
		assert.Equal(t, got, oid.String())
	}
}

func TestObjectIDUnmarshalAmino(t *testing.T) {
	hexid := strings.Repeat("0a", HashSize)
	for _, tc := range []struct {
		in      string
		want    uint64
		wantErr bool
	}{
		{in: hexid + ":0", want: 0},
		{in: hexid + ":42", want: 42},
		{in: hexid + ":" + strconv.Itoa(math.MaxInt64), want: math.MaxInt64},
		{in: hexid, wantErr: true},
		{in: hexid + ":1:2", wantErr: true},
		{in: hexid + "::1", wantErr: true},
		{in: hexid + ":", wantErr: true},
		{in: hexid + ":x", wantErr: true},
		{in: "zz:1", wantErr: true},
		{in: "", wantErr: true},
	} {
		t.Run(tc.in, func(t *testing.T) {
			var oid ObjectID
			err := oid.UnmarshalAmino(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, oid.NewTime)
			assert.Equal(t, tc.in, oid.String())
		})
	}
}
