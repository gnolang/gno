package btree_dao

import (
	"errors"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/btree"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

// RegistrationDetails holds the details of a user's registration
type RegistrationDetails struct {
	Address   std.Address
	RegTime   time.Time
	UserBTree *btree.BTree
	NFTID     int
}

// Less implements the Record interface for RegistrationDetails
// It compares based on the time of registration
func (rd *RegistrationDetails) Less(than btree.Record) bool {
	other := than.(*RegistrationDetails)
	return rd.RegTime.Before(other.RegTime)
}

var (
	dao     = grc721.NewBasicNFT("BTree DAO", "BTDAO")
	tokenID = 0
	members = btree.New()
)

// PlantTree allows a user to plant a B-Tree in the DAO forest
func PlantTree(userBTree *btree.BTree) error {
	return plantImpl(userBTree, "")
}

// PlantSeed allows a user to register as a seed with a message
func PlantSeed(message string) error {
	return plantImpl(nil, message)
}

// plantImpl is the internal implementation that handles both trees and seeds
func plantImpl(userBTree *btree.BTree, seedMessage string) error {
	// Get the caller's address
	var userAddress std.Address
	if std.PrevRealm().IsUser() {
		userAddress = std.PrevRealm().Addr()
	} else {
		userAddress = std.GetOrigCaller()
	}

	var nftID string
	var regDetails *RegistrationDetails

	if userBTree != nil {
		// Handle tree planting
		var treeExists bool
		members.Ascend(func(record btree.Record) bool {
			regDetails := record.(*RegistrationDetails)
			if regDetails.UserBTree == userBTree {
				treeExists = true
				return false
			}
			return true
		})
		if treeExists {
			return errors.New("tree is already planted in the forest")
		}

		if userBTree.Len() == 0 {
			return errors.New("cannot plant an empty tree")
		}

		nftID = ufmt.Sprintf("%d", tokenID)
		regDetails = &RegistrationDetails{
			Address:   userAddress,
			RegTime:   time.Now(),
			UserBTree: userBTree,
			NFTID:     tokenID,
		}
	} else {
		// Handle seed planting
		if seedMessage == "" {
			return errors.New("seed message cannot be empty")
		}
		nftID = "seed_" + ufmt.Sprintf("%d", tokenID)
		regDetails = &RegistrationDetails{
			Address:   userAddress,
			RegTime:   time.Now(),
			UserBTree: nil,
			NFTID:     tokenID,
		}
	}

	// Mint an NFT to the user
	err := dao.Mint(userAddress, grc721.TokenID(nftID))
	if err != nil {
		return err
	}

	members.Insert(regDetails)
	tokenID++
	return nil
}

// Render generates a Markdown representation of the DAO members,
// displaying the latest 10 members and the first 3 members as "OGs".
func Render(path string) string {
	var latestMembers []string
	var ogMembers []string

	// Get total size and first member
	totalSize := members.Len()
	biggestTree := 0
	if maxMember := members.Max(); maxMember != nil {
		if userBTree := maxMember.(*RegistrationDetails).UserBTree; userBTree != nil {
			biggestTree = userBTree.Len()
		}
	}

	// Collect the latest 10 members
	members.Descend(func(record btree.Record) bool {
		if len(latestMembers) < 10 {
			regDetails := record.(*RegistrationDetails)
			addr := regDetails.Address
			nftList := ""
			balance, err := dao.BalanceOf(addr)
			if err == nil && balance > 0 {
				nftList = " (NFTs: "
				for i := uint64(0); i < balance; i++ {
					if i > 0 {
						nftList += ", "
					}
					if regDetails.UserBTree == nil {
						nftList += "🌱#" + ufmt.Sprintf("seed_%d", regDetails.NFTID) // Seed emoji for seeds
					} else {
						nftList += "🌳#" + ufmt.Sprintf("%d", regDetails.NFTID) // Tree emoji for tree planters
					}
				}
				nftList += ")"
			}
			latestMembers = append(latestMembers, string(addr)+nftList)
			return true
		}
		return false
	})

	// Collect the first 3 members (OGs)
	members.Ascend(func(record btree.Record) bool {
		if len(ogMembers) < 3 {
			regDetails := record.(*RegistrationDetails)
			addr := regDetails.Address
			nftList := ""
			balance, err := dao.BalanceOf(addr)
			if err == nil && balance > 0 {
				nftList = " (NFTs: "
				for i := uint64(0); i < balance; i++ {
					if i > 0 {
						nftList += ", "
					}
					if regDetails.UserBTree == nil {
						nftList += "🌱#" + ufmt.Sprintf("seed_%d", regDetails.NFTID) // Seed emoji for seeds
					} else {
						nftList += "🌳#" + ufmt.Sprintf("%d", regDetails.NFTID) // Tree emoji for tree planters
					}
				}
				nftList += ")"
			}
			ogMembers = append(ogMembers, string(addr)+nftList)
			return true
		}
		return false
	})

	// Use the md package to format the output
	var sb strings.Builder

	sb.WriteString(md.H1("B-Tree DAO Members"))
	sb.WriteString(md.H2("Total NFTs Minted"))
	sb.WriteString(ufmt.Sprintf("Total NFTs minted: %d\n\n", dao.TokenCount()))
	sb.WriteString(md.H2("Member Stats"))
	sb.WriteString(ufmt.Sprintf("Total members: %d\n", totalSize))
	if biggestTree > 0 {
		sb.WriteString(ufmt.Sprintf("Biggest tree size: %d\n", biggestTree))
	}
	sb.WriteString(md.H2("OG Members"))
	sb.WriteString(md.BulletList(ogMembers))
	sb.WriteString(md.H2("Latest Members"))
	sb.WriteString(md.BulletList(latestMembers))

	return sb.String()
}
