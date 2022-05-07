package bitarray

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	bitarraypb "github.com/gnolang/gno/pkgs/bitarray/pb"
)

func (goo BitArray) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *bitarraypb.BitArray
	{
		pbo = new(bitarraypb.BitArray)
		{
			pbo.Bits = int64(goo.Bits)
		}
		{
			goorl := len(goo.Elems)
			if goorl == 0 {
				pbo.Elems = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Elems[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Elems = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo BitArray) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(bitarraypb.BitArray)
	msg = pbo
	return
}
func (goo *BitArray) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *bitarraypb.BitArray = msg.(*bitarraypb.BitArray)
	{
		if pbo != nil {
			{
				(*goo).Bits = int(int(pbo.Bits))
			}
			{
				var pbol int = 0
				if pbo.Elems != nil {
					pbol = len(pbo.Elems)
				}
				if pbol == 0 {
					(*goo).Elems = nil
				} else {
					var goors = make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Elems[i]
							{
								pboev := pboe
								goors[i] = uint64(pboev)
							}
						}
					}
					(*goo).Elems = goors
				}
			}
		}
	}
	return
}
func (_ BitArray) GetTypeURL() (typeURL string) {
	return "/tm.BitArray"
}
