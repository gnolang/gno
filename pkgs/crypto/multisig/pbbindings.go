package multisig

import (
	amino "github.com/gnolang/gno/pkgs/amino"
	crypto "github.com/gnolang/gno/pkgs/crypto"
	multisigpb "github.com/gnolang/gno/pkgs/crypto/multisig/pb"
	proto "google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func (goo PubKeyMultisigThreshold) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *multisigpb.PubKeyMultisig
	{
		if IsPubKeyMultisigReprEmpty(goo) {
			var pbov *multisigpb.PubKeyMultisig
			msg = pbov
			return
		}
		pbo = new(multisigpb.PubKeyMultisig)
		{
			pbo.K = uint64(goo.K)
		}
		{
			goorl := len(goo.PubKeys)
			if goorl == 0 {
				pbo.PubKeys = nil
			} else {
				pbos := make([]*anypb.Any, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.PubKeys[i]
						{
							if goore != nil {
								typeUrl := cdc.GetTypeURL(goore)
								bz := []byte(nil)
								bz, err = cdc.Marshal(goore)
								if err != nil {
									return
								}
								pbos[i] = &anypb.Any{TypeUrl: typeUrl, Value: bz}
							}
						}
					}
				}
				pbo.PubKeys = pbos
			}
		}
	}
	msg = pbo
	return
}

func (goo PubKeyMultisigThreshold) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(multisigpb.PubKeyMultisig)
	msg = pbo
	return
}

func (goo *PubKeyMultisigThreshold) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *multisigpb.PubKeyMultisig = msg.(*multisigpb.PubKeyMultisig)
	{
		if pbo != nil {
			{
				(*goo).K = uint(uint(pbo.K))
			}
			{
				var pbol int = 0
				if pbo.PubKeys != nil {
					pbol = len(pbo.PubKeys)
				}
				if pbol == 0 {
					(*goo).PubKeys = nil
				} else {
					goors := make([]crypto.PubKey, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.PubKeys[i]
							{
								pboev := pboe
								typeUrl := pboev.TypeUrl
								bz := pboev.Value
								goorp := &goors[i]
								err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
								if err != nil {
									return
								}
							}
						}
					}
					(*goo).PubKeys = goors
				}
			}
		}
	}
	return
}

func (_ PubKeyMultisigThreshold) GetTypeURL() (typeURL string) {
	return "/tm.PubKeyMultisig"
}

func IsPubKeyMultisigReprEmpty(goor PubKeyMultisigThreshold) (empty bool) {
	{
		empty = true
		{
			if goor.K != 0 {
				return false
			}
		}
		{
			if len(goor.PubKeys) != 0 {
				return false
			}
		}
	}
	return
}
