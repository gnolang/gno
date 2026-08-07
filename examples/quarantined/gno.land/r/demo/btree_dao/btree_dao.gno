package btree_dao

import (
	"chain/runtime/unsafe"
	"errors"
	"strings"
	"time"

	"gno.land/p/demo/btree"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt/v0"
)

type RegistrationDetails struct {
	Address   address
	RegTime   time.Time
	UserBTree *btree.BTree
	NFTID     string
}

// Less orders records by registration time (btree.Record interface).
func (rd *RegistrationDetails) Less(than btree.Record) bool {
	other := than.(*RegistrationDetails)
	return rd.RegTime.Before(other.RegTime)
}

var (
	dao       *grc721.Token
	daoLedger *grc721.PrivateLedger
	tokenID   = 0
	members   = btree.New()
)

func init(cur realm) {
	dao, daoLedger = grc721.NewToken("BTree DAO", "BTDAO", 0, cur)
}

func PlantTree(_ realm, userBTree *btree.BTree) error {
	return plantImpl(userBTree, "")
}

func PlantSeed(_ realm, message string) error {
	return plantImpl(nil, message)
}

// plantImpl handles both tree planting (userBTree != nil) and seed registration.
func plantImpl(userBTree *btree.BTree, seedMessage string) error {
	userAddress := unsafe.OriginCaller()

	var nftID string
	var regDetails *RegistrationDetails

	if userBTree != nil {
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

	err := daoLedger.Mint(userAddress, grc721.TokenID(nftID))
	if err != nil {
		return err
	}

	members.Insert(regDetails)
	tokenID++
	return nil
}

func Render(path string) string {
	var latestMembers []string
	var ogMembers []string

	totalSize := members.Len()
	biggestTree := 0
	if maxMember := members.Max(); maxMember != nil {
		if userBTree := maxMember.(*RegistrationDetails).UserBTree; userBTree != nil {
			biggestTree = userBTree.Len()
		}
	}

	members.Descend(func(record btree.Record) bool {
		if len(latestMembers) < 10 {
			regDetails := record.(*RegistrationDetails)
			addr := regDetails.Address
			nftList := ""
			balance := dao.BalanceOf(addr)
			if balance > 0 {
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

	members.Ascend(func(record btree.Record) bool {
		if len(ogMembers) < 3 {
			regDetails := record.(*RegistrationDetails)
			addr := regDetails.Address
			nftList := ""
			balance := dao.BalanceOf(addr)
			if balance > 0 {
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
	sb.WriteString(ufmt.Sprintf("Total NFTs minted: %d\n\n", dao.TotalSupply()))
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
