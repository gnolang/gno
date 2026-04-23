// Package crosspkg holds types used by tests in the parent `tests` package
// to exercise cross-package code generation paths in genproto2. The types
// live in a child package so that, from the tests package's perspective,
// their reflect.Type has a different PkgPath() than the target.
package crosspkg

// SmallCount is an AminoMarshaler with uint8 repr. When used as
// []*crosspkg.SmallCount in a struct field, it exercises the packed-list
// + pointer + cross-package branch in gen_marshal.go's writeUnpackedListMarshal.
// A regression to `new(SmallCount)` (bare name) fails to compile from tests/.
type SmallCount uint8

func (c SmallCount) MarshalAmino() (uint8, error) {
	return uint8(c), nil
}

func (c *SmallCount) UnmarshalAmino(repr uint8) error {
	*c = SmallCount(repr)
	return nil
}

// BoxedInt is an AminoMarshaler whose repr is a same-package struct. When
// a tests-package type uses BoxedInt as a field, the repr decode in
// gen_unmarshal.go writeReprUnmarshal references `crosspkg.Inner` which
// must be qualified. A regression to `var repr Inner` fails to compile.
type BoxedInt struct {
	V int64
}

type Inner struct {
	N int64
}

func (b BoxedInt) MarshalAmino() (Inner, error) {
	return Inner{N: b.V}, nil
}

func (b *BoxedInt) UnmarshalAmino(r Inner) error {
	b.V = r.N
	return nil
}
