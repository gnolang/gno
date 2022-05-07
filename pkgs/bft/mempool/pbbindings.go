package mempool

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	mempoolpb "github.com/gnolang/gno/pkgs/bft/mempool/pb"
)

func (goo TxMessage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *mempoolpb.TxMessage
	{
		pbo = new(mempoolpb.TxMessage)
		{
			goorl := len(goo.Tx)
			if goorl == 0 {
				pbo.Tx = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Tx[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Tx = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo TxMessage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(mempoolpb.TxMessage)
	msg = pbo
	return
}
func (goo *TxMessage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *mempoolpb.TxMessage = msg.(*mempoolpb.TxMessage)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Tx != nil {
					pbol = len(pbo.Tx)
				}
				if pbol == 0 {
					(*goo).Tx = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Tx[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Tx = goors
				}
			}
		}
	}
	return
}
func (_ TxMessage) GetTypeURL() (typeURL string) {
	return "/tm.TxMessage"
}
