package secp256k1

import (
	amino "github.com/gnolang/gno/pkgs/amino"
	secp256k1pb "github.com/gnolang/gno/pkgs/crypto/secp256k1/pb"
	proto "google.golang.org/protobuf/proto"
)

func (goo PubKeySecp256k1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *secp256k1pb.PubKeySecp256K1
	{
		if IsPubKeySecp256k1ReprEmpty(goo) {
			var pbov *secp256k1pb.PubKeySecp256K1
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &secp256k1pb.PubKeySecp256K1{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (goo PubKeySecp256k1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(secp256k1pb.PubKeySecp256K1)
	msg = pbo
	return
}

func (goo *PubKeySecp256k1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *secp256k1pb.PubKeySecp256K1 = msg.(*secp256k1pb.PubKeySecp256K1)
	{
		goors := [33]uint8{}
		for i := 0; i < 33; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = uint8(uint8(pboev))
				}
			}
		}
		*goo = goors
	}
	return
}

func (_ PubKeySecp256k1) GetTypeURL() (typeURL string) {
	return "/tm.PubKeySecp256k1"
}

func IsPubKeySecp256k1ReprEmpty(goor PubKeySecp256k1) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (goo PrivKeySecp256k1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *secp256k1pb.PrivKeySecp256K1
	{
		if IsPrivKeySecp256k1ReprEmpty(goo) {
			var pbov *secp256k1pb.PrivKeySecp256K1
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &secp256k1pb.PrivKeySecp256K1{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (goo PrivKeySecp256k1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(secp256k1pb.PrivKeySecp256K1)
	msg = pbo
	return
}

func (goo *PrivKeySecp256k1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *secp256k1pb.PrivKeySecp256K1 = msg.(*secp256k1pb.PrivKeySecp256K1)
	{
		goors := [32]uint8{}
		for i := 0; i < 32; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = uint8(uint8(pboev))
				}
			}
		}
		*goo = goors
	}
	return
}

func (_ PrivKeySecp256k1) GetTypeURL() (typeURL string) {
	return "/tm.PrivKeySecp256k1"
}

func IsPrivKeySecp256k1ReprEmpty(goor PrivKeySecp256k1) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}
