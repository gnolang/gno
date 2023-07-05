package groups

import (
	"encoding/binary"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/binutils"
)

type MemberID uint64

type Member struct {
	id        MemberID
	address   std.Address
	weight    int
	metadata  string
	createdAt time.Time
}

func (mid MemberID) String() string {
	return strconv.Itoa(int(mid))
}

func (member *Member) getMemberStr() string {
	memberDataStr := ""
	memberDataStr += "\t\t\t[" + memberIDKey(member.id) + ", " + member.address.String() + ", " + strconv.Itoa(member.weight) + ", " + member.metadata + ", " + member.createdAt.String() + "],\n\n"
	return memberDataStr
}

func (member *Member) Bytes() []byte {
	b := []byte{}
	b = binary.BigEndian.AppendUint64(b, uint64(member.id))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(member.address.String())...)
	b = binary.BigEndian.AppendUint32(b, uint32(member.weight))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(member.metadata)...)
	b = binary.BigEndian.AppendUint64(b, uint64(member.createdAt.UnixNano()))
	return b
}
