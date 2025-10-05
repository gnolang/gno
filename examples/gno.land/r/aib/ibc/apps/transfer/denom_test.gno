package transfer_test

import (
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
	"gno.land/r/aib/ibc/apps/transfer"
)

func TestDenomsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		denom    transfer.Denom
		expError string
	}{
		{
			"valid trace with client id",
			transfer.NewDenom("uatone",
				transfer.NewHop("transfer", "07-tendermint-0"),
			),
			"",
		},
		{
			"valid multiple trace info",
			transfer.NewDenom("uatom",
				transfer.NewHop("transfer", "channel-1"),
				transfer.NewHop("transfer", "channel-2"),
			),
			"",
		},
		{
			"empty base denom with trace",
			transfer.NewDenom("", transfer.NewHop("transfer", "channel-1")),
			"base denomination cannot be blank",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := tc.denom.ValidateBasic()

			if tc.expError == "" {
				urequire.NoError(t, err)
				return
			}
			urequire.ErrorContains(t, err, tc.expError)
		})
	}
}

func TestExtractDenomFromPath(t *testing.T) {
	testCases := []struct {
		name     string
		fullPath string
		expDenom transfer.Denom
	}{
		{"empty denom", "", transfer.Denom{}},
		{"base denom no slashes", "atom", transfer.NewDenom("atom")},
		{"base denom with trailing slash", "atom/", transfer.NewDenom("atom/")},
		{"base denom multiple trailing slash", "foo///bar//baz/atom/", transfer.NewDenom("foo///bar//baz/atom/")},
		{"ibc denom one hop", "transfer/channel-0/atom", transfer.NewDenom("atom", transfer.NewHop("transfer", "channel-0"))},
		{"ibc denom one hop with client id", "transfer/07-tendermint-0/atom", transfer.NewDenom("atom", transfer.NewHop("transfer", "07-tendermint-0"))},
		{"ibc denom one hop trailing slash", "transfer/channel-0/atom/", transfer.NewDenom("atom/", transfer.NewHop("transfer", "channel-0"))},
		{"ibc denom one hop multiple slashes", "transfer/channel-0//at/om/", transfer.NewDenom("/at/om/", transfer.NewHop("transfer", "channel-0"))},
		{"ibc denom two hops", "transfer/channel-0/transfer/channel-60/atom", transfer.NewDenom("atom", transfer.NewHop("transfer", "channel-0"), transfer.NewHop("transfer", "channel-60"))},
		{"ibc denom two hops trailing slash", "transfer/channel-0/transfer/channel-60/atom/", transfer.NewDenom("atom/", transfer.NewHop("transfer", "channel-0"), transfer.NewHop("transfer", "channel-60"))},
		{"empty prefix", "/uatom", transfer.NewDenom("/uatom")},
		{"empty identifiers", "//uatom", transfer.NewDenom("//uatom")},
		{"base denom with single '/'", "erc20/0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA", transfer.NewDenom("erc20/0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA")},
		{"trace info and base denom with single '/'", "transfer/channel-1/erc20/0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA", transfer.NewDenom("erc20/0x85bcBCd7e79Ec36f4fBBDc54F90C643d921151AA", transfer.NewHop("transfer", "channel-1"))},
		{"single trace identifier", "transfer/", transfer.NewDenom("transfer/")},
		{"trace info with custom port", "customtransfer/channel-1/uatom", transfer.NewDenom("uatom", transfer.NewHop("customtransfer", "channel-1"))},
		{"invalid path (1)", "channel-1/transfer/uatom", transfer.NewDenom("channel-1/transfer/uatom")},
		{"invalid path (2)", "transfer/channel-1", transfer.NewDenom("transfer/channel-1")},
		{"invalid path (3)", "transfer/channel-1/transfer/channel-2", transfer.NewDenom("", transfer.NewHop("transfer", "channel-1"), transfer.NewHop("transfer", "channel-2"))},
		{"invalid path (4)", "transfer/channelToA/uatom", transfer.NewDenom("transfer/channelToA/uatom")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			denom := transfer.ExtractDenomFromPath(tc.fullPath)

			uassert.Equal(t, tc.expDenom.Base, denom.Base)
			if uassert.Equal(t, len(tc.expDenom.Trace), len(denom.Trace), "trace: unexpected number of hops") {
				for i := range tc.expDenom.Trace {
					uassert.Equal(t, tc.expDenom.Trace[i].PortId, denom.Trace[i].PortId,
						ufmt.Sprintf("unexpected hop #%d portId", i))
					uassert.Equal(t, tc.expDenom.Trace[i].ClientId, denom.Trace[i].ClientId,
						ufmt.Sprintf("unexpected hop #%d channelId", i))
				}
			}
		})
	}
}
