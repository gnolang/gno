package ed25519

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	ed25519pb "github.com/gnolang/gno/pkgs/crypto/ed25519/pb"
)

func (goo PubKeyEd25519) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *ed25519pb.PubKeyEd25519
	{
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &ed25519pb.PubKeyEd25519{Value: pbos}
		}
	}
	msg = pbo
	return
}
func (goo PubKeyEd25519) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(ed25519pb.PubKeyEd25519)
	msg = pbo
	return
}
func (goo *PubKeyEd25519) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *ed25519pb.PubKeyEd25519 = msg.(*ed25519pb.PubKeyEd25519)
	{
		var goors = [32]uint8{}
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
func (_ PubKeyEd25519) GetTypeURL() (typeURL string) {
	return "/tm.PubKeyEd25519"
}
func (goo PrivKeyEd25519) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *ed25519pb.PrivKeyEd25519
	{
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &ed25519pb.PrivKeyEd25519{Value: pbos}
		}
	}
	msg = pbo
	return
}
func (goo PrivKeyEd25519) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(ed25519pb.PrivKeyEd25519)
	msg = pbo
	return
}
func (goo *PrivKeyEd25519) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *ed25519pb.PrivKeyEd25519 = msg.(*ed25519pb.PrivKeyEd25519)
	{
		var goors = [64]uint8{}
		for i := 0; i < 64; i += 1 {
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
func (_ PrivKeyEd25519) GetTypeURL() (typeURL string) {
	return "/tm.PrivKeyEd25519"
}
