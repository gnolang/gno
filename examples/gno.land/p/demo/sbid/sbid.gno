package sbid

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

type Bid struct {
	Infos        *avl.Tree // caller <- *bidInfo because the *Bid point to the current session, so list of key is list of callers
	EndHashTime  int64
	EndPriceTime int64
	StartTime    int64
}

type bidInfo struct {
	BidObject        string
	Bidder           string
	HashString       string
	Price            int64
	IsCommittedHash  bool
	IsCommittedPrice bool
	IsWinner         bool
}

type BidStatus struct {
	BidObject    string
	Status       string
	EndHashTime  int64
	EndPriceTime int64
	ActiveUsers  int64
}

// SetTime sets the start time, end time for each phase of commit
func (b *Bid) SetTime(startTime, endHashTime, endPriceTime int64) {
	b.StartTime = startTime
	b.EndHashTime = endHashTime
	b.EndPriceTime = endPriceTime
}

// *Bid is pointer to bid session in root tree. It is defined by object. -> only 1 Bid for each bidObject
// need 2 phases. these functions are for the commit hash string and commit the details
func (b *Bid) CommitHash(domainName string, hashString string, caller string) string {
	// TODO: Check for time of each phases?
	now := time.Now().UnixMilli()
	if now < b.StartTime {
		return "not the hash time"
	}
	if now > b.EndHashTime {
		return "can not commit hash anymore"
	}
	// check if caller is committed hash - can only commit once
	if b.Infos.Has(caller) {
		return "can not commit hash twice"
	}
	// new bid
	newBidInfo := bidInfo{
		BidObject:        domainName,
		Bidder:           caller,
		HashString:       hashString,
		IsCommittedHash:  true,
		IsCommittedPrice: false,
		IsWinner:         false,
	}
	// set to info storage
	b.Infos.Set(caller, &newBidInfo)
	// TODO: Charge fee ? -> dapp's scope
	return "committed hash"
}

// CommitPrice will check validity, duplicate, or time
func (b *Bid) CommitPrice(domainName string, price int64, secret string, caller string) string {
	if time.Now().UnixMilli() > b.EndPriceTime {
		return "can not commit price and secret anymore"
	}
	if time.Now().UnixMilli() <= b.EndHashTime {
		return "price phase is not started"
	}
	// check in bidInfos list
	data, ext := b.Infos.Get(caller)
	if !ext {
		return "you are not in the info list"
	}
	info := data.(*bidInfo)
	if !info.IsCommittedHash {
		return "not committed hash yet"
	}
	if info.IsCommittedPrice {
		return "can not commit price twice"
	}
	hashCalculated := calculateHS(price, secret)
	if hashCalculated != info.HashString {
		return "invalid secret key or price"
	}
	if info.IsCommittedHash && hashCalculated == info.HashString && !info.IsCommittedPrice {
		info.IsCommittedPrice = true
		info.Price = price
		info.IsWinner = false
		// setup new record

		b.calculateAndRecordWinner(caller, info)
		return "committed price"
	}

	return "undefined"
}

func (b *Bid) GetCurrentWinnerInfo() bidInfo {
	winner := bidInfo{}
	b.Infos.Iterate("", "", func(key string, value interface{}) bool {
		bid := value.(*bidInfo)
		if bid.IsWinner {
			winner = *bid
			return true
		}
		return false
	})
	return winner
}

// GetStatus return status list of all joined session by a caller -> implement on realm
// on this function, just check status of the session that caller joined
func (b *Bid) GetStatus(domainName string, caller string) BidStatus {
	// return status
	bStt := BidStatus{}

	infoData, ext := b.Infos.Get(caller)
	if !ext {
		bStt.Status = "free"
		return bStt
	}
	info := infoData.(*bidInfo)
	stt := b.getStatusFromInfo(domainName, *info)
	bStt.BidObject = domainName
	bStt.Status = stt
	bStt.EndHashTime = b.EndHashTime
	bStt.EndPriceTime = b.EndPriceTime
	bStt.ActiveUsers = int64(b.Infos.Size())
	return bStt
}

// get current status of a bid info in the bidding session that caller joined. Because the info is caller's info, then will return this caller win or not
func (b *Bid) getStatusFromInfo(object string, info bidInfo) string {
	now := time.Now().UnixMilli()

	// Check if the auction is closed then return winner ?
	// TODO: Re-design this logical
	if now > b.EndPriceTime {
		if info.IsCommittedPrice {
			winner := b.GetCurrentWinnerInfo()
			if winner.Bidder != "" {
				return winner.Bidder + " is claiming " + object
			}
		}
		return "close"
	}

	// during bid session
	switch {
	// Waiting for hash commitment
	case !info.IsCommittedHash && !info.IsCommittedPrice && b.StartTime <= now && b.EndHashTime > now:
		return "waiting hash"

	// Committed hash
	case info.IsCommittedHash && !info.IsCommittedPrice && b.StartTime <= now && b.EndHashTime > now:
		return "committed hash"

	// Waiting for price commitment
	case info.IsCommittedHash && !info.IsCommittedPrice && b.EndHashTime <= now && b.EndPriceTime > now:
		return "waiting price"

	// Committed price
	case info.IsCommittedHash && info.IsCommittedPrice && b.EndHashTime <= now && b.EndPriceTime > now:
		return "committed price"

	// Winner claiming
	case info.IsCommittedHash && info.IsCommittedPrice && !info.IsWinner && b.EndPriceTime <= now:
		winner := b.GetCurrentWinnerInfo()
		if winner.Bidder == "" {
			return "no winner yet"
		}
		return winner.Bidder + " is claiming " + object

	// Missed hash commitment deadline
	case !info.IsCommittedHash && b.EndHashTime <= now:
		return "waiting for hash"

	default:
		return "undefined"
	}
}

// logic for handle new price commit - check winner
func (b *Bid) calculateAndRecordWinner(currentCaller string, newRecord *bidInfo) string {
	if b.Infos.Size() == 1 {
		// update winner now
		newRecord.IsWinner = true
		b.Infos.Set(currentCaller, newRecord)
		return "update first element - winner"
	}

	// simple solutions: take the list of bidInfo, find the winner, set it. Next time check for IsWinner
	infoList := []*bidInfo{}
	b.Infos.Iterate("", "", func(key string, value interface{}) bool {
		bidInf := value.(*bidInfo)
		infoList = append(infoList, bidInf)
		return false
	})
	// find max of bidInfo.Price
	maxIndex := 0
	maxPrice := infoList[0].Price
	for index, bif := range infoList {
		if bif.Price > maxPrice {
			maxPrice = bif.Price
			maxIndex = index
		}
	}
	// if newRecord wins
	if maxPrice < newRecord.Price {
		newRecord.IsWinner = true
		b.Infos.Set(newRecord.Bidder, newRecord)
		loserBid := infoList[maxIndex]
		loserBid.IsWinner = false
		b.Infos.Set(loserBid.Bidder, loserBid)
	} else if maxPrice >= newRecord.Price {
		// save newRecord and update old records
		newRecord.IsWinner = false
		winnerBid := infoList[maxIndex]
		winnerBid.IsWinner = true
		b.Infos.Set(currentCaller, newRecord)
		b.Infos.Set(winnerBid.Bidder, winnerBid)
	}
	return "update winner for " + currentCaller
}

// calculate hash string from secret+price
func calculateHS(price int64, secret string) string {
	input := secret + strconv.Itoa(int(price))
	data := []byte(input)
	hashed := sha256.Sum256(data)
	hashedBytes := hashed[:]
	return hex.EncodeToString(hashedBytes)
}
