package btree_dao

import (
	"std"
	"strings"
	"time"
	"errors"
	"gno.land/p/demo/btree"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/moul/md"
	"gno.land/p/demo/ufmt"

)
// RegistrationDetails holds the details of a user's registration
type RegistrationDetails struct {
	Address     std.Address
	RegTime     time.Time
	UserBTree   *btree.BTree
	NFTID       int
}

// Less implements the Record interface for RegistrationDetails
// It compares based on the time of registration
func (rd *RegistrationDetails) Less(than btree.Record) bool {
	other := than.(*RegistrationDetails)
	return rd.RegTime.Before(other.RegTime)
}

var (
	dao = grc721.NewBasicNFT("BTree DAO", "BTDAO")
	tokenID = 0
	members = btree.New()
)


// Register allows a user to register by sending their B-Tree instance
func Register(userBTree *btree.BTree) error {
	// Check if this tree is already registered
	var treeExists bool
	members.Ascend(func(record btree.Record) bool {
		regDetails := record.(*RegistrationDetails)
		if regDetails.UserBTree == userBTree {
			treeExists = true
			return false // Stop iteration
		}
		return true
	})
	if treeExists {
		return errors.New("tree is already registered")
	}

	// Get the caller's address - use GetOrigCaller() if called from init()
	var userAddress std.Address
	if std.PrevRealm().IsUser() {
		userAddress = std.PrevRealm().Addr() // technically this will never happen as you can't register a btree from manually created tx
	} else {
		userAddress = std.GetOrigCaller()
	}

	// Check if the user has used a B-Tree or just made an empty one
	// User can also cheat this system but I don't see the point
	if userBTree == nil || userBTree.Len() == 0 {
		return errors.New("Invalid B-Tree instance")
	}

	// Mint an NFT to the user
	err := dao.Mint(userAddress, grc721.TokenID(tokenID))
	if err != nil {
		return err
	}

	regDetails := &RegistrationDetails{
		Address:   userAddress,
		RegTime:   time.Now(),
		UserBTree: userBTree,
		NFTID:     tokenID,
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
					nftList += "#" + ufmt.Sprintf("%d", regDetails.NFTID)
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
					nftList += "#" + ufmt.Sprintf("%d", regDetails.NFTID)
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
	sb.WriteString(md.H2("OG Members"))
	sb.WriteString(md.BulletList(ogMembers))
	sb.WriteString(md.H2("Latest Members"))
	sb.WriteString(md.BulletList(latestMembers))

	return sb.String()
}
