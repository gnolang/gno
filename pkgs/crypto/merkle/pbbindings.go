package merkle

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	merklepb "github.com/gnolang/gno/pkgs/crypto/merkle/pb"
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
func (goo SimpleProof) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *merklepb.SimpleProof
	{
		if IsSimpleProofReprEmpty(goo) {
			var pbov *merklepb.SimpleProof
			msg = pbov
			return
		}
		pbo = new(merklepb.SimpleProof)
		{
			pbo.Total = int64(goo.Total)
		}
		{
			pbo.Index = int64(goo.Index)
		}
		{
			goorl := len(goo.LeafHash)
			if goorl == 0 {
				pbo.LeafHash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.LeafHash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.LeafHash = pbos
			}
		}
		{
			goorl := len(goo.Aunts)
			if goorl == 0 {
				pbo.Aunts = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Aunts[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint8, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = byte(goore)
										}
									}
								}
								pbos[i] = pbos1
							}
						}
					}
				}
				pbo.Aunts = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo SimpleProof) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(merklepb.SimpleProof)
	msg = pbo
	return
}
func (goo *SimpleProof) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *merklepb.SimpleProof = msg.(*merklepb.SimpleProof)
	{
		if pbo != nil {
			{
				(*goo).Total = int(int(pbo.Total))
			}
			{
				(*goo).Index = int(int(pbo.Index))
			}
			{
				var pbol int = 0
				if pbo.LeafHash != nil {
					pbol = len(pbo.LeafHash)
				}
				if pbol == 0 {
					(*goo).LeafHash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.LeafHash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).LeafHash = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Aunts != nil {
					pbol = len(pbo.Aunts)
				}
				if pbol == 0 {
					(*goo).Aunts = nil
				} else {
					var goors = make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Aunts[i]
							{
								pboev := pboe
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									goors[i] = nil
								} else {
									var goors1 = make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboev[i]
											{
												pboev := pboe
												goors1[i] = uint8(uint8(pboev))
											}
										}
									}
									goors[i] = goors1
								}
							}
						}
					}
					(*goo).Aunts = goors
				}
			}
		}
	}
	return
}
func (_ SimpleProof) GetTypeURL() (typeURL string) {
	return "/tm.SimpleProof"
}
func IsSimpleProofReprEmpty(goor SimpleProof) (empty bool) {
	{
		empty = true
		{
			if goor.Total != 0 {
				return false
			}
		}
		{
			if goor.Index != 0 {
				return false
			}
		}
		{
			if len(goor.LeafHash) != 0 {
				return false
			}
		}
		{
			if len(goor.Aunts) != 0 {
				return false
			}
		}
	}
	return
}
func (goo SimpleProofNode) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *merklepb.SimpleProofNode
	{
		if IsSimpleProofNodeReprEmpty(goo) {
			var pbov *merklepb.SimpleProofNode
			msg = pbov
			return
		}
		pbo = new(merklepb.SimpleProofNode)
		{
			goorl := len(goo.Hash)
			if goorl == 0 {
				pbo.Hash = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Hash[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Hash = pbos
			}
		}
		{
			if goo.Parent != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Parent.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Parent = pbom.(*merklepb.SimpleProofNode)
				if pbo.Parent == nil {
					pbo.Parent = new(merklepb.SimpleProofNode)
				}
			}
		}
		{
			if goo.Left != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Left.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Left = pbom.(*merklepb.SimpleProofNode)
				if pbo.Left == nil {
					pbo.Left = new(merklepb.SimpleProofNode)
				}
			}
		}
		{
			if goo.Right != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Right.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Right = pbom.(*merklepb.SimpleProofNode)
				if pbo.Right == nil {
					pbo.Right = new(merklepb.SimpleProofNode)
				}
			}
		}
	}
	msg = pbo
	return
}
func (goo SimpleProofNode) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(merklepb.SimpleProofNode)
	msg = pbo
	return
}
func (goo *SimpleProofNode) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *merklepb.SimpleProofNode = msg.(*merklepb.SimpleProofNode)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Hash != nil {
					pbol = len(pbo.Hash)
				}
				if pbol == 0 {
					(*goo).Hash = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Hash[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Hash = goors
				}
			}
			{
				if pbo.Parent != nil {
					(*goo).Parent = new(SimpleProofNode)
					err = (*goo).Parent.FromPBMessage(cdc, pbo.Parent)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Left != nil {
					(*goo).Left = new(SimpleProofNode)
					err = (*goo).Left.FromPBMessage(cdc, pbo.Left)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.Right != nil {
					(*goo).Right = new(SimpleProofNode)
					err = (*goo).Right.FromPBMessage(cdc, pbo.Right)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ SimpleProofNode) GetTypeURL() (typeURL string) {
	return "/tm.SimpleProofNode"
}
func IsSimpleProofNodeReprEmpty(goor SimpleProofNode) (empty bool) {
	{
		empty = true
		{
			if len(goor.Hash) != 0 {
				return false
			}
		}
		{
			if goor.Parent != nil {
				return false
			}
		}
		{
			if goor.Left != nil {
				return false
			}
		}
		{
			if goor.Right != nil {
				return false
			}
		}
	}
	return
}
