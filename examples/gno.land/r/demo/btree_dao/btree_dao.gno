package btree_dao

import (
	"chain/runtime"
	"errors"
	"strings"
	"time"

	"gno.land/p/demo/btree"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt/v0"
)

// RegistrationDetails holds the details of a user's registration in the BTree DAO.
// It stores the user's address, registration time, their B-Tree if they planted one,
// and their NFT ID.
type RegistrationDetails struct {
	Address   address
	RegTime   time.Time
	UserBTree *btree.BTree
	NFTID     string
}

// Less implements the btree.Record interface for RegistrationDetails.
// It compares two RegistrationDetails based on their registration time.
// Returns true if the current registration time is before the other registration time.
func (rd *RegistrationDetails) Less(than btree.Record) bool {
	other := than.(*RegistrationDetails)
	return rd.RegTime.Before(other.RegTime)
}

var (
	dao     = grc721.NewBasicNFT("BTree DAO", "BTDAO")
	tokenID = 0
	members = btree.New()
)

// PlantTree allows a user to plant their B-Tree in the DAO forest.
// It mints an NFT to the user and registers their tree in the DAO.
// Returns an error if the tree is already planted, empty, or if NFT minting fails.
func PlantTree(_ realm, userBTree *btree.BTree) error {
	return plantImpl(userBTree, "")
}

// PlantSeed allows a user to register as a seed in the DAO with a message.
// It mints an NFT to the user and registers them as a seed member.
// Returns an error if the message is empty or if NFT minting fails.
func PlantSeed(_ realm, message string) error {
	return plantImpl(nil, message)
}

// plantImpl is the internal implementation that handles both tree planting and seed registration.
// For tree planting (userBTree != nil), it verifies the tree isn't already planted and isn't empty.
// For seed planting (userBTree == nil), it verifies the seed message isn't empty.
// In both cases, it mints an NFT to the user and adds their registration details to the members tree.
// Returns an error if any validation fails or if NFT minting fails.
func plantImpl(userBTree *btree.BTree, seedMessage string) error {
	// Get the caller's address
	userAddress := runtime.OriginCaller()

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
			NFTID:     nftID,
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
			NFTID:     nftID,
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

// Render generates a Markdown representation of the DAO members.
// It displays:
// - Total number of NFTs minted
// - Total number of members
// - Size of the biggest planted tree
// - The first 3 members (OGs)
// - The latest 10 members
// Each member entry includes their address and owned NFTs (🌳 for trees, 🌱 for seeds).
// The path parameter is currently unused.
// Returns a formatted Markdown string.
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
				for i := int64(0); i < balance; i++ {
					if i > 0 {
						nftList += ", "
					}
					if regDetails.UserBTree == nil {
						nftList += "🌱#" + regDetails.NFTID
					} else {
						nftList += "🌳#" + regDetails.NFTID
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
				for i := int64(0); i < balance; i++ {
					if i > 0 {
						nftList += ", "
					}
					if regDetails.UserBTree == nil {
						nftList += "🌱#" + regDetails.NFTID
					} else {
						nftList += "🌳#" + regDetails.NFTID
					}
				}
				nftList += ")"
			}
			ogMembers = append(ogMembers, string(addr)+nftList)
			return true
		}
		return false
	})

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
