package openocean

import (
	"std"
	"testing"
)

func TestInitCollection(t *testing.T) {
	{
		const adr std.Address = "g125fhetg9kc3surkxh78avppegc3997l3jr4wdf"
		var nftID uint64 = 0
		description := "This is a planet collection"
		clt, err := CreateCollection(
			0,
			&nftID,
			"Neptune",
			"NEP",
			adr,
			description,
			"no logo",
			5,
		)

		if err != nil {
			t.Fatal(err)
		}

		expectedId := "0"
		expectedName := "Neptune"
		expectedSymbol := "NEP"
		expectedOwner := adr
		expectedDescription := description
		expectedLogo := "no logo"
		var expectedAvaiableTokens uint64 = 5

		if expectedId != clt.GetId() {
			t.Fatalf("expected %s, got %s", expectedId, clt.GetId())
		}

		if expectedName != clt.GetName() {
			t.Fatalf("expected %s, got %s", expectedName, clt.GetName())
		}

		if expectedSymbol != clt.GetSymbol() {
			t.Fatalf("expected %s, got %s", expectedSymbol, clt.GetSymbol())
		}

		if expectedOwner != clt.GetOwner() {
			t.Fatalf("expected %s, got %s", expectedOwner, clt.GetOwner())
		}

		if expectedDescription != clt.GetDescription() {
			t.Fatalf("expected %s, got %s", expectedDescription, clt.GetDescription())
		}

		if expectedLogo != clt.GetLogo() {
			t.Fatalf("expected %s, got %s", expectedLogo, clt.GetLogo())
		}

		if expectedAvaiableTokens != clt.GetAvaiableTokens(){
			t.Fatalf("expected %d avai, got %d", expectedAvaiableTokens, clt.GetAvaiableTokens())
		}
	}
	{
		const adr std.Address = "g125fhetg9kc3surkxh78avppegc3997l3jr4wdf"
		var nftID uint64 = 1
		description := "This is a planet collection"
		clt, err := CreateCollection(
			0,
			&nftID,
			"Neptune",
			"NEP",
			adr,
			description,
			"no logo",
			5,
		)

		if err != nil {
			t.Fatal(err)
		}


	}
	{
		countOwner := 1
		var countNfts uint64 = 0
		expectedNbOfOwners := 1
		var expectedNbOfNfts uint64 = 5
		const adr std.Address = "g125fhetg9kc3surkxh78avppegc3997l3jr4wdf"
		var nftID uint64 = 0
		description := "This is a planet collection"
		clt, err := CreateCollection(
			0,
			&nftID,
			"Neptune",
			"NEP",
			adr,
			description,
			"no logo",
			5,
		)
		nfts := clt.GetNFTS()

		if err != nil {
			t.Fatal(err)
		}


		nfts.Iterate("", "", func (id string, i interface{}) bool {
			nft := i.(*NFTtoken721)
			if nft.GetOwner() != adr {
				countOwner++
			}
			countNfts++
			return false
		})

		if expectedNbOfOwners != countOwner{
			t.Fatalf("expected %d owners of nfts, got %d", expectedNbOfOwners, countOwner)
		}

		if expectedNbOfNfts != countNfts{
			t.Fatalf("expected %d nfts, got %d", expectedNbOfNfts, countNfts)
		}
	}
}
