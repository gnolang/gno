package merkle

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/tendermint/go-amino-x"
	merklepb "github.com/tendermint/classic/crypto/merkle/pb"
)

func (goo ProofOp) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *merklepb.ProofOp
	{
		if IsProofOpReprEmpty(goo) {
			var pbov *merklepb.ProofOp
			msg = pbov
			return
		}
		pbo = new(merklepb.ProofOp)
		{
			pbo.Type = string(goo.Type)
		}
		{
			goorl := len(goo.Key)
			if goorl == 0 {
				pbo.Key = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Key[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Key = pbos
			}
		}
		{
			goorl := len(goo.Data)
			if goorl == 0 {
				pbo.Data = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Data[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Data = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo ProofOp) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(merklepb.ProofOp)
	msg = pbo
	return
}
func (goo *ProofOp) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *merklepb.ProofOp = msg.(*merklepb.ProofOp)
	{
		if pbo != nil {
			{
				(*goo).Type = string(pbo.Type)
			}
			{
				var pbol int = 0
				if pbo.Key != nil {
					pbol = len(pbo.Key)
				}
				if pbol == 0 {
					(*goo).Key = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Key[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Key = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Data != nil {
					pbol = len(pbo.Data)
				}
				if pbol == 0 {
					(*goo).Data = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Data[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Data = goors
				}
			}
		}
	}
	return
}
func (_ ProofOp) GetTypeURL() (typeURL string) {
	return "/tm.ProofOp"
}
func IsProofOpReprEmpty(goor ProofOp) (empty bool) {
	{
		empty = true
		{
			if goor.Type != "" {
				return false
			}
		}
		{
			if len(goor.Key) != 0 {
				return false
			}
		}
		{
			if len(goor.Data) != 0 {
				return false
			}
		}
	}
	return
}
func (goo Proof) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *merklepb.Proof
	{
		if IsProofReprEmpty(goo) {
			var pbov *merklepb.Proof
			msg = pbov
			return
		}
		pbo = new(merklepb.Proof)
		{
			goorl := len(goo.Ops)
			if goorl == 0 {
				pbo.Ops = nil
			} else {
				var pbos = make([]*merklepb.ProofOp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Ops[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*merklepb.ProofOp)
						}
					}
				}
				pbo.Ops = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo Proof) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(merklepb.Proof)
	msg = pbo
	return
}
func (goo *Proof) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *merklepb.Proof = msg.(*merklepb.Proof)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Ops != nil {
					pbol = len(pbo.Ops)
				}
				if pbol == 0 {
					(*goo).Ops = nil
				} else {
					var goors = make([]ProofOp, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Ops[i]
							{
								pboev := pboe
								if pboev != nil {
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*goo).Ops = goors
				}
			}
		}
	}
	return
}
func (_ Proof) GetTypeURL() (typeURL string) {
	return "/tm.Proof"
}
func IsProofReprEmpty(goor Proof) (empty bool) {
	{
		empty = true
		{
			if len(goor.Ops) != 0 {
				return false
			}
		}
	}
	return
}
