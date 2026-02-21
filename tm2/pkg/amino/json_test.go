package amino_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/pkg"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

type Dummy struct{}

var gopkg = reflect.TypeOf(Dummy{}).PkgPath()

var transportPackage = pkg.NewPackage(gopkg, "amino_test", "").
	WithTypes(&Transport{}, Car(""), insurancePlan(0), Boat(""), Plane{})

func registerTransports(cdc *amino.Codec) {
	cdc.RegisterPackage(transportPackage)
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)
	cases := []struct {
		in      any
		want    string
		wantErr string
	}{
		{&noFields{}, "{}", ""},                        // #0
		{&noExportedFields{a: 10, b: "foo"}, "{}", ""}, // #1
		{nil, "null", ""},                              // #2
		{&oneExportedField{}, `{"A":""}`, ""},          // #3
		{Car(""), `""`, ""},                            // #4
		{Car("Tesla"), `"Tesla"`, ""},                  // #5
		{&oneExportedField{A: "Z"}, `{"A":"Z"}`, ""},   // #6
		{[]string{"a", "bc"}, `["a","bc"]`, ""},        // #7
		{
			[]any{"a", "bc", 10, 10.93, 1e3},
			``, "unregistered",
		}, // #8
		{
			aPointerField{Foo: new(int), Name: "name"},
			`{"Foo":"0","nm":"name"}`, "",
		}, // #9
		{
			aPointerFieldAndEmbeddedField{intPtr(11), "ap", nil, &oneExportedField{A: "foo"}},
			`{"Foo":"11","nm":"ap","bz":{"A":"foo"}}`, "",
		}, // #10
		{
			doublyEmbedded{
				Inner: &aPointerFieldAndEmbeddedField{
					intPtr(11), "ap", nil, &oneExportedField{A: "foo"},
				},
			},
			`{"Inner":{"Foo":"11","nm":"ap","bz":{"A":"foo"}},"year":0}`, "",
		}, // #11
		{
			struct{}{}, `{}`, "",
		}, // #12
		{
			struct{ A int }{A: 10}, `{"A":"10"}`, "",
		}, // #13
		{
			Transport{},
			`{"Vehicle":null,"Capacity":"0"}`, "",
		}, // #14
		{
			Transport{Vehicle: Car("Bugatti")},
			`{"Vehicle":{"@type":"/amino_test.Car","value":"Bugatti"},"Capacity":"0"}`, "",
		}, // #15
		{
			BalanceSheet{Assets: []Asset{Car("Corolla"), insurancePlan(1e7)}},
			`{"assets":[{"@type":"/amino_test.Car","value":"Corolla"},{"@type":"/amino_test.insurancePlan","value":"10000000"}]}`, "",
		}, // #16
		{
			Transport{Vehicle: Boat("Poseidon"), Capacity: 1789},
			`{"Vehicle":{"@type":"/amino_test.Boat","value":"Poseidon"},"Capacity":"1789"}`, "",
		}, // #17
		{
			withCustomMarshaler{A: &aPointerField{Foo: intPtr(12)}, F: customJSONMarshaler(10)},
			`{"fx":"10","A":{"Foo":"12"}}`, "",
		}, // #18 (NOTE: MarshalJSON of customJSONMarshaler has no effect)
		{
			func() json.Marshaler { v := customJSONMarshaler(10); return &v }(),
			`"10"`, "",
		}, // #19 (NOTE: MarshalJSON of customJSONMarshaler has no effect)
		{
			interfacePtr("a"), `{"@type":"/google.protobuf.StringValue","value":"a"}`, "",
		}, // #20
		{&fp{"Foo", 10}, `"Foo@10"`, ""}, // #21
		{(*fp)(nil), "null", ""},         // #22
		{
			struct {
				FP      *fp
				Package string
			}{FP: &fp{"Foo", 10}, Package: "bytes"},
			`{"FP":"Foo@10","Package":"bytes"}`, "",
		}, // #23
	}

	for i, tt := range cases {
		t.Logf("Trying case #%v", i)
		blob, err := cdc.JSONMarshal(tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%v\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v\nblob: %v", i, err, tt.in)
			continue
		}
		if g, w := string(blob), tt.want; g != w {
			t.Errorf("#%d:\ngot:\n\t%s\nwant:\n\t%s", i, g, w)
		}
	}
}

func TestMarshalJSONTime(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().Round(0).UTC(), // strip monotonic.
	}

	b, err := cdc.JSONMarshal(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.JSONUnmarshal(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)
}

func TestMarshalJSONPBTime(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)

	type SimpleStruct struct {
		Timestamp *timestamppb.Timestamp
		Duration  *durationpb.Duration
	}

	s := SimpleStruct{
		Timestamp: &timestamppb.Timestamp{Seconds: 1296012345, Nanos: 940483},
		Duration:  &durationpb.Duration{Seconds: 100},
	}

	b, err := cdc.JSONMarshal(s)
	assert.Nil(t, err)

	var s2 SimpleStruct
	err = cdc.JSONUnmarshal(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)
}

type fp struct {
	Name    string
	Version int
}

func (f fp) MarshalAmino() (string, error) {
	return fmt.Sprintf("%v@%v", f.Name, f.Version), nil
}

func (f *fp) UnmarshalAmino(repr string) (err error) {
	parts := strings.Split(repr, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format %v", repr)
	}
	f.Name = parts[0]
	f.Version, err = strconv.Atoi(parts[1])
	return
}

type innerFP struct {
	PC uint64
	FP *fp
}

// We don't support maps.
func TestUnmarshalMap(t *testing.T) {
	t.Parallel()

	jsonBytes := []byte("dontcare")
	obj := new(map[string]int)
	cdc := amino.NewCodec()
	assert.Panics(t, func() {
		err := cdc.JSONUnmarshal(jsonBytes, &obj)
		assert.Fail(t, "should have panicked but got err: %v", err)
	})
	assert.Panics(t, func() {
		err := cdc.JSONUnmarshal(jsonBytes, obj)
		assert.Fail(t, "should have panicked but got err: %v", err)
	})
	assert.Panics(t, func() {
		bz, err := cdc.JSONMarshal(obj)
		assert.Fail(t, "should have panicked but got bz: %X err: %v", bz, err)
	})
}

func TestUnmarshalFunc(t *testing.T) {
	t.Parallel()

	jsonBytes := []byte(`"dontcare"`)
	obj := func() {}
	cdc := amino.NewCodec()
	assert.Panics(t, func() {
		err := cdc.JSONUnmarshal(jsonBytes, &obj)
		assert.Fail(t, "should have panicked but got err: %v", err)
	})

	err := cdc.JSONUnmarshal(jsonBytes, obj)
	// JSONUnmarshal expects a pointer
	assert.Error(t, err)

	// ... nor encoding it.
	assert.Panics(t, func() {
		bz, err := cdc.JSONMarshal(obj)
		assert.Fail(t, "should have panicked but got bz: %X err: %v", bz, err)
	})
}

func TestJSONUnmarshal(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)
	cases := []struct {
		blob    string
		in      any
		want    any
		wantErr string
	}{
		{ // #0
			`null`, 2, nil, "expected a pointer",
		},
		{ // #1
			`null`, new(int), new(int), "",
		},
		{ // #2
			`"2"`, new(int), intPtr(2), "",
		},
		{ // #3
			`{"null"}`, new(int), nil, "invalid character",
		},
		{ // #4
			`{"Vehicle":null,"Capacity":"0"}`, new(Transport), new(Transport), "",
		},
		{ // #5
			`{"Vehicle":{"@type":"/amino_test.Car","value":"Bugatti"},"Capacity":"10"}`,
			new(Transport),
			&Transport{
				Vehicle:  Car("Bugatti"),
				Capacity: 10,
			}, "",
		},
		{ // #6
			`"Bugatti"`, new(Car), func() *Car { c := Car("Bugatti"); return &c }(), "",
		},
		{ // #7
			`["1", "2", "3"]`, new([]int), func() any {
				v := []int{1, 2, 3}
				return &v
			}(), "",
		},
		{ // #8
			`["1", "2", "3"]`, new([]string), func() any {
				v := []string{"1", "2", "3"}
				return &v
			}(), "",
		},
		{ // #9
			`[{"@type":"/google.protobuf.Int32Value","value":1},{"@type":"/google.protobuf.StringValue","value":"2"}]`,
			new([]any), &([]any{int32(1), string("2")}), "",
		},
		{ // #10
			`2.34`, floatPtr(2.34), nil, "float* support requires",
		},
		{ // #11
			`"FooBar@1"`, new(fp), &fp{"FooBar", 1}, "",
		},
		{ // #12
			`"10@0"`, new(fp), &fp{Name: "10"}, "",
		},
		{ // #13
			`{"PC":"125","FP":"10@0"}`, new(innerFP), &innerFP{PC: 125, FP: &fp{Name: `10`}}, "",
		},
		{ // #14
			`{"PC":"125","FP":"<FP-FOO>@0"}`, new(innerFP), &innerFP{PC: 125, FP: &fp{Name: `<FP-FOO>`}}, "",
		},
	}

	for i, tt := range cases {
		err := cdc.JSONUnmarshal([]byte(tt.blob), tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error: %v\nblob: %s\nin: %+v\n", i, err, tt.blob, tt.in)
			continue
		}
		if g, w := tt.in, tt.want; !reflect.DeepEqual(g, w) {
			gb, err := json.MarshalIndent(g, "", "  ")
			require.NoError(t, err)
			wb, err := json.MarshalIndent(w, "", "  ")
			require.NoError(t, err)
			t.Errorf("#%d:\ngot:\n\t%#v\n(%s)\n\nwant:\n\t%#v\n(%s)", i, g, gb, w, wb)
		}
	}
}

func TestJSONCodecRoundTrip(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)
	type allInclusive struct {
		Tr      Transport `json:"trx"`
		Vehicle Vehicle   `json:"v,omitempty"`
		Comment string
		Data    []byte
	}

	cases := []struct {
		in      any
		want    any
		out     any
		wantErr string
	}{
		0: {
			in: &allInclusive{
				Tr: Transport{
					Vehicle: Boat("Oracle"),
				},
				Comment: "To the Cosmos! баллинг в космос",
				Data:    []byte("祝你好运"),
			},
			out: new(allInclusive),
			want: &allInclusive{
				Tr: Transport{
					Vehicle: Boat("Oracle"),
				},
				Comment: "To the Cosmos! баллинг в космос",
				Data:    []byte("祝你好运"),
			},
		},

		1: {
			in:   Transport{Vehicle: Plane{Name: "G6", MaxAltitude: 51e3}, Capacity: 18},
			out:  new(Transport),
			want: &Transport{Vehicle: Plane{Name: "G6", MaxAltitude: 51e3}, Capacity: 18},
		},
	}

	for i, tt := range cases {
		mBlob, err := cdc.JSONMarshal(tt.in)
		if tt.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("#%d:\ngot:\n\t%q\nwant non-nil error containing\n\t%q", i,
					err, tt.wantErr)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: unexpected error after JSONMarshal: %v", i, err)
			continue
		}

		if err = cdc.JSONUnmarshal(mBlob, tt.out); err != nil {
			t.Errorf("#%d: unexpected error after JSONUnmarshal: %v\nmBlob: %s", i, err, mBlob)
			continue
		}

		// Now check that the input is exactly equal to the output
		uBlob, err := cdc.JSONMarshal(tt.out)
		assert.NoError(t, err)
		if err := cdc.JSONUnmarshal(mBlob, tt.out); err != nil {
			t.Errorf("#%d: unexpected error after second JSONMarshal: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(tt.want, tt.out) {
			t.Errorf("#%d: After roundtrip JSONUnmarshal\ngot: \t%v\nwant:\t%v", i, tt.out, tt.want)
		}
		if !bytes.Equal(mBlob, uBlob) {
			t.Errorf("#%d: After roundtrip JSONMarshal\ngot: \t%s\nwant:\t%s", i, uBlob, mBlob)
		}
	}
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

type (
	noFields         struct{}
	noExportedFields struct {
		a int
		b string
	}
)

type oneExportedField struct {
	A string
}

type aPointerField struct {
	Foo  *int
	Name string `json:"nm,omitempty"`
}

type doublyEmbedded struct {
	Inner *aPointerFieldAndEmbeddedField
	Year  int32 `json:"year"`
}

type aPointerFieldAndEmbeddedField struct {
	Foo  *int
	Name string `json:"nm,omitempty"`
	*oneExportedField
	B *oneExportedField `json:"bz,omitempty"`
}

type customJSONMarshaler int

var _ json.Marshaler = (*customJSONMarshaler)(nil)

func (cm customJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`"WRONG"`), nil
}

type withCustomMarshaler struct {
	F customJSONMarshaler `json:"fx"`
	A *aPointerField
}

type Transport struct {
	Vehicle
	Capacity int
}

type Vehicle interface {
	Move() error
}

type Asset interface {
	Value() float64
}

func (c Car) Value() float64 {
	return 60000.0
}

type BalanceSheet struct {
	Assets []Asset `json:"assets"`
}

type (
	Car   string
	Boat  string
	Plane struct {
		Name        string
		MaxAltitude int64
	}
)
type insurancePlan int

func (ip insurancePlan) Value() float64 { return float64(ip) }

func (c Car) Move() error   { return nil }
func (b Boat) Move() error  { return nil }
func (p Plane) Move() error { return nil }

func interfacePtr(v any) *any {
	return &v
}

// Test to ensure that Amino codec's time encoding/decoding roundtrip
// produces the same result as the standard library json's.
func TestAminoJSONTimeEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	din := time.Date(2008, 9, 15, 14, 13, 12, 11109876, loc).Round(time.Millisecond).UTC()

	cdc := amino.NewCodec()
	blobAmino, err := cdc.JSONMarshal(din)
	require.Nil(t, err, "amino.Codec.JSONMarshal should succeed")
	var tAminoOut time.Time
	require.Nil(t, cdc.JSONUnmarshal(blobAmino, &tAminoOut), "amino.Codec.JSONUnmarshal should succeed")
	require.NotEqual(t, tAminoOut, time.Time{}, "amino.marshaled definitely isn't equal to zero time")
	require.Equal(t, tAminoOut, din, "expecting marshaled in to be equal to marshaled out")

	blobStdlib, err := json.Marshal(din)
	require.Nil(t, err, "json.Marshal should succeed")
	var tStdlibOut time.Time
	require.Nil(t, json.Unmarshal(blobStdlib, &tStdlibOut), "json.Unmarshal should succeed")
	require.NotEqual(t, tStdlibOut, time.Time{}, "stdlib.marshaled definitely isn't equal to zero time")
	require.Equal(t, tStdlibOut, din, "expecting stdlib.marshaled to be equal to time in")

	require.Equal(t, tAminoOut, tStdlibOut, "expecting amino.unmarshalled to be equal to json.unmarshalled")
}

func TestMarshalJSONIndent(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	registerTransports(cdc)
	obj := Transport{Vehicle: Car("Tesla")}
	expected := `{
  "Vehicle": {
    "@type": "/amino_test.Car",
    "value": "Tesla"
  },
  "Capacity": "0"
}`

	blob, err := cdc.MarshalJSONIndent(obj, "", "  ")
	assert.Nil(t, err)
	assert.Equal(t, expected, string(blob))
}

func TestJSONInterfaceTypeAssignability(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)

	type AnyWrapper struct {
		Value any
	}
	type Interface1Wrapper struct {
		Value tests.Interface1
	}

	src := AnyWrapper{Value: tests.PrimitivesStruct{Int: 42}}
	bz, err := cdc.JSONMarshal(src)
	require.NoError(t, err)

	var dst Interface1Wrapper
	err = cdc.JSONUnmarshal(bz, &dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not assignable")
}
