package conn_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

func TestCodecParity_Conn(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(conn.Package)
	cdc.Seal()

	cases := []struct {
		name string
		v    any
	}{
		{"PacketPing", &conn.PacketPing{}},
		{"PacketPong", &conn.PacketPong{}},
		{"PacketMsg/empty", &conn.PacketMsg{}},
		{"PacketMsg/payload", &conn.PacketMsg{
			ChannelID: 7,
			EOF:       1,
			Bytes:     []byte("hello world"),
		}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
