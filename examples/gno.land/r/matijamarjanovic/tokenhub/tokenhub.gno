package tokenhub

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

var (
	registeredNFTs = avl.NewTree() // symbol -> NFTGetter

	ErrNFTAlreadyRegistered = errors.New("NFT already registered")
	ErrNFTNotFound          = errors.New("NFT not found")
)

func RegisterNFT(nftGetter grc721.NFTGetter) error {
	nft := nftGetter()
	metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
	if !ok {
		return errors.New("NFT must implement IGRC721CollectionMetadata")
	}

	symbol := metadata.Symbol()

	if registeredNFTs.Has(symbol) {
		return ErrNFTAlreadyRegistered
	}

	registeredNFTs.Set(symbol, nftGetter)
	return nil
}

func getBalances(addr std.Address) string {
	balances := avl.NewTree()
	var sb strings.Builder

	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		balance, _ := nft.BalanceOf(addr)
		balances.Set(key, balance)
		return false
	})

	sb.WriteString("NFT Balances: ")
	balances.Iterate("", "", func(key string, value interface{}) bool {
		balance := value.(uint64)
		sb.WriteString(ufmt.Sprintf("%s: %d ", key, balance))
		return false
	})

	return sb.String()
}

func GetMyBalances() string {
	return getBalances(std.GetOrigCaller())
}

func Render(path string) string {
	var sb strings.Builder

	sb.WriteString(md.H1("NFT Hub"))
	sb.WriteString(md.Paragraph("A central registry for GRC721 NFTs on gno.land"))

	sb.WriteString(md.H2("Registered NFTs"))

	var nftItems []string
	registeredNFTs.Iterate("", "", func(key string, value interface{}) bool {
		nftGetter := value.(grc721.NFTGetter)
		nft := nftGetter()
		metadata, ok := nft.(grc721.IGRC721CollectionMetadata)
		if !ok {
			return false
		}
		nftItems = append(nftItems, ufmt.Sprintf("%s (%s)",
			metadata.Name(),
			metadata.Symbol()))
		return false
	})

	if len(nftItems) > 0 {
		sb.WriteString(md.BulletList(nftItems))
	} else {
		sb.WriteString(md.Italic("No NFTs registered yet"))
		sb.WriteString("\n")
	}

	return sb.String()
}
