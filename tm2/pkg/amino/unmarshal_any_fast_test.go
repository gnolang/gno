package amino_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func init() {
	amino.RegisterInterfaceSetter[tests.Interface1]()
}

func fastAnyCodec() *amino.Codec {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()
	return cdc
}

// anyDecoders decode Any bytes into an interface pointer three ways: by
// reflection (amino.Unmarshal), by genproto2 assigning with reflect.Value.Set,
// and by genproto2 assigning through the registered setter.
func anyDecoders(cdc *amino.Codec) map[string]func(bz []byte) (tests.Interface1, error) {
	return map[string]func(bz []byte) (tests.Interface1, error){
		"reflect": func(bz []byte) (v tests.Interface1, err error) {
			err = cdc.Unmarshal(bz, &v)
			return
		},
		"genproto2": func(bz []byte) (v tests.Interface1, err error) {
			amino.WithoutInterfaceSetters(func() { err = cdc.UnmarshalAnyBinary2(bz, &v, 0) })
			return
		},
		"genproto2+setter": func(bz []byte) (v tests.Interface1, err error) {
			err = cdc.UnmarshalAnyBinary2(bz, &v, 0)
			return
		},
	}
}

func TestUnmarshalAnyBinary2_MatchesReflect(t *testing.T) {
	cdc := fastAnyCodec()
	recursive := tests.ConcreteRecursive{Inner: tests.ConcreteRecursive{Inner: tests.Concrete1{}}}
	cases := []struct {
		name string
		v    tests.Interface1
	}{
		{"empty struct", tests.Concrete1{}},
		{"other concrete, same interface", tests.Concrete2{}},
		{"type definition of an array", tests.ConcreteTypeDef{1, 2, 3, 4}},
		{"wrapped bytes", tests.ConcreteWrappedBytes{Value: []byte("abc")}},
		{"wrapped empty bytes", tests.ConcreteWrappedBytes{Value: []byte{}}},
		{"pointer concrete, nil interface fields", &tests.InterfaceFieldsStruct{}},
		{"pointer concrete, interface fields set", &tests.InterfaceFieldsStruct{
			F1: tests.Concrete1{}, F2: tests.ConcreteTypeDef{9, 9, 9, 9},
			F3: tests.Concrete2{}, F4: &tests.InterfaceFieldsStruct{F1: tests.Concrete2{}},
		}},
		{"nested Any, nil leaf", tests.ConcreteRecursive{}},
		{"nested Any, three levels", recursive},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.Marshal(&tc.v)
			require.NoError(t, err)
			decs := anyDecoders(cdc)
			want, err := decs["reflect"](bz)
			require.NoError(t, err)
			require.Equal(t, reflect.TypeOf(tc.v), reflect.TypeOf(want))
			for name, dec := range decs {
				got, err := dec(bz)
				require.NoError(t, err, name)
				// Compared to the reflect result, not tc.v: amino is lossy on
				// purpose for some values (an empty []byte decodes as nil).
				require.Equal(t, want, got, name)
				rbz, err := cdc.Marshal(&got)
				require.NoError(t, err, name)
				require.True(t, bytes.Equal(bz, rbz), "%s: re-encoding differs\n want %X\n got  %X", name, bz, rbz)
			}
		})
	}
}

// Malformed and edge inputs: every decoder must agree on success or failure,
// and on the value when they succeed.
func TestUnmarshalAnyBinary2_EdgeInputsAgree(t *testing.T) {
	cdc := fastAnyCodec()
	good, err := cdc.Marshal(&[]tests.Interface1{tests.ConcreteWrappedBytes{Value: []byte("x")}}[0])
	require.NoError(t, err)
	cases := []struct {
		name string
		bz   []byte
	}{
		{"empty input", []byte{}},
		{"valid", good},
		{"unknown type url", []byte{0x0A, 0x04, '/', 'x', '.', 'Y'}},
		{"unknown field after value", append(append([]byte{}, good...), 0x18, 0x01)},
		{"truncated", good[:len(good)-1]},
		{"type url only", good[:2+int(good[1])]},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results := map[string]string{}
			for name, dec := range anyDecoders(cdc) {
				got, err := dec(tc.bz)
				if err != nil {
					results[name] = "error"
				} else if got == nil {
					results[name] = "nil"
				} else {
					results[name] = reflect.TypeOf(got).String()
				}
			}
			require.Equal(t, results["genproto2"], results["genproto2+setter"], "setter changed the outcome: %v", results)
			if results["reflect"] != results["genproto2"] {
				t.Logf("reflect and genproto2 differ on master too (not changed here): %v", results)
			}
		})
	}
}

// Struct values whose fields are interfaces: the generated UnmarshalBinary2
// (which reaches UnmarshalAnyBinary2 per field) against reflection.
func TestUnmarshalBinary2_InterfaceFieldsMatchReflect(t *testing.T) {
	cdc := fastAnyCodec()
	cases := []struct {
		name string
		v    amino.PBMessager2
		zero func() amino.PBMessager2
	}{
		{"typed value, nil interfaces", &tests.GnoVMTypedValue{}, func() amino.PBMessager2 { return &tests.GnoVMTypedValue{} }},
		{"typed value", &tests.GnoVMTypedValue{T: tests.Concrete1{}, V: tests.ConcreteTypeDef{1, 2, 3, 4}, N: [8]byte{8}}, func() amino.PBMessager2 { return &tests.GnoVMTypedValue{} }},
		{"block, embedded struct and repeated values", &tests.GnoVMBlock{
			Source: tests.ConcreteRecursive{Inner: tests.Concrete2{}},
			Values: []tests.GnoVMTypedValue{{T: tests.Concrete1{}}, {}, {V: tests.ConcreteWrappedBytes{Value: []byte("v")}}},
		}, func() amino.PBMessager2 { return &tests.GnoVMBlock{} }},
		{"node, deep embedding and repeated interface", &tests.GnoVMNode{
			Op: 7, Left: tests.Concrete1{},
			Args: []tests.Interface1{tests.Concrete1{}, tests.Concrete2{}, tests.ConcreteTypeDef{4, 3, 2, 1}},
		}, func() amino.PBMessager2 { return &tests.GnoVMNode{} }},
		{"interface heavy", &tests.InterfaceHeavy{
			Field1: tests.Concrete1{}, Field2: tests.Concrete2{},
			Items: []tests.Interface1{tests.Concrete1{}, tests.Concrete2{}}, Name: "n",
		}, func() amino.PBMessager2 { return &tests.InterfaceHeavy{} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalReflect(tc.v)
			require.NoError(t, err)

			viaReflect := tc.zero()
			require.NoError(t, cdc.UnmarshalReflect(bz, viaReflect))

			viaGen := tc.zero()
			amino.WithoutInterfaceSetters(func() { err = viaGen.UnmarshalBinary2(cdc, bz, 0) })
			require.NoError(t, err)

			viaSetter := tc.zero()
			require.NoError(t, viaSetter.UnmarshalBinary2(cdc, bz, 0))

			require.Equal(t, viaReflect, viaGen)
			require.Equal(t, viaReflect, viaSetter)
			rbz, err := cdc.MarshalReflect(viaSetter)
			require.NoError(t, err)
			require.Equal(t, bz, rbz)
		})
	}
}

func TestRegisterInterfaceSetter_RejectsNonInterface(t *testing.T) {
	require.Panics(t, func() { amino.RegisterInterfaceSetter[tests.Concrete1]() })
}
