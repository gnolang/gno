package amino_test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
	"github.com/stretchr/testify/assert"
)

func TestAnyWellKnownNative(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	s1 := tests.InterfaceFieldsStruct{
		F3: string("dontcare"),
		F4: int(0),
	}

	bz, err := cdc.Marshal(s1)
	assert.Nil(t, err)
	assert.Equal(t,
		//     0x1a --> field #3 Typ3ByteLength (F3)
		//           0x2a --> length prefix (42 bytes)
		//                 0x0a --> field #1 Typ3ByteLength (Any TypeURL)
		//                       0x1c --> length prefix (28 bytes)
		//                             0x2f, ... 0x65 --> "/google.protobuf.StringValue"
		[]byte{
			0x1a, 0x2a, 0x0a, 0x1c, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65,
			//   0x12 --> field #2 Typ3ByteLength (Any Value)
			//         0x0a --> length prefix (10 bytes)
			//               0x0a --> field #1, one and only, of implicit struct.
			//                     0x08 --> length prefix (8 bytes)
			//                           0x64, ... 0x65 --> "dontcare"
			/**/ 0x12, 0x0a, 0x0a, 0x08, 0x64, 0x6f, 0x6e, 0x74, 0x63, 0x61, 0x72, 0x65,
			//   0x22 --> field #4 Typ3ByteLength (F4)
			//         0x1d --> length prefix (29 bytes)
			//               0x0a --> field #1 Typ3ByteLength (Any TypeURL)
			//                     0x1b --> length prefix (27 bytes)
			//                           0x2f, ... 0x65 --> "/google.protobuf.Int64Value"
			/**/ 0x22, 0x1d, 0x0a, 0x1b, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65,
		},
		bz,
		"InterfaceFieldsStruct incorrectly serialized")

	var s2 tests.InterfaceFieldsStruct
	err = cdc.Unmarshal(bz, &s2)
	assert.NoError(t, err)

	s3 := tests.InterfaceFieldsStruct{
		F3: string("dontcare"),
		F4: int64(0), // ints get decoded as int64.
	}
	assert.Equal(t, s3, s2)
}
