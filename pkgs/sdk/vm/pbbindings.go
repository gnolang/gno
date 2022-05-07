package vm

import (
	proto "google.golang.org/protobuf/proto"
	amino "github.com/gnolang/gno/pkgs/amino"
	vmpb "github.com/gnolang/gno/pkgs/sdk/vm/pb"
	stdpb "github.com/gnolang/gno/pkgs/std/pb"
	std "github.com/gnolang/gno/pkgs/std"
)

func (goo MsgCall) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *vmpb.MCall
	{
		pbo = new(vmpb.MCall)
		{
			goor, err1 := goo.Caller.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Caller = string(goor)
		}
		{
			goor, err1 := goo.Send.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Send = string(goor)
		}
		{
			pbo.PkgPath = string(goo.PkgPath)
		}
		{
			pbo.Func = string(goo.Func)
		}
		{
			goorl := len(goo.Args)
			if goorl == 0 {
				pbo.Args = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Args[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.Args = pbos
			}
		}
	}
	msg = pbo
	return
}
func (goo MsgCall) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(vmpb.MCall)
	msg = pbo
	return
}
func (goo *MsgCall) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *vmpb.MCall = msg.(*vmpb.MCall)
	{
		if pbo != nil {
			{
				var goor string
				goor = string(pbo.Caller)
				err = (*goo).Caller.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				var goor string
				goor = string(pbo.Send)
				err = (*goo).Send.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				(*goo).PkgPath = string(pbo.PkgPath)
			}
			{
				(*goo).Func = string(pbo.Func)
			}
			{
				var pbol int = 0
				if pbo.Args != nil {
					pbol = len(pbo.Args)
				}
				if pbol == 0 {
					(*goo).Args = nil
				} else {
					var goors = make([]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Args[i]
							{
								pboev := pboe
								goors[i] = string(pboev)
							}
						}
					}
					(*goo).Args = goors
				}
			}
		}
	}
	return
}
func (_ MsgCall) GetTypeURL() (typeURL string) {
	return "/vm.m_call"
}
func (goo MsgAddPackage) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *vmpb.MAddpkg
	{
		pbo = new(vmpb.MAddpkg)
		{
			goor, err1 := goo.Creator.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Creator = string(goor)
		}
		{
			if goo.Package != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.Package.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.Package = pbom.(*stdpb.MemPackage)
				if pbo.Package == nil {
					pbo.Package = new(stdpb.MemPackage)
				}
			}
		}
		{
			goor, err1 := goo.Deposit.MarshalAmino()
			if err1 != nil {
				return nil, err1
			}
			pbo.Deposit = string(goor)
		}
	}
	msg = pbo
	return
}
func (goo MsgAddPackage) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(vmpb.MAddpkg)
	msg = pbo
	return
}
func (goo *MsgAddPackage) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *vmpb.MAddpkg = msg.(*vmpb.MAddpkg)
	{
		if pbo != nil {
			{
				var goor string
				goor = string(pbo.Creator)
				err = (*goo).Creator.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
			{
				if pbo.Package != nil {
					(*goo).Package = new(std.MemPackage)
					err = (*goo).Package.FromPBMessage(cdc, pbo.Package)
					if err != nil {
						return
					}
				}
			}
			{
				var goor string
				goor = string(pbo.Deposit)
				err = (*goo).Deposit.UnmarshalAmino(goor)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ MsgAddPackage) GetTypeURL() (typeURL string) {
	return "/vm.m_addpkg"
}
func (goo InvalidPkgPathError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *vmpb.InvalidPkgPathError
	{
		pbo = new(vmpb.InvalidPkgPathError)
	}
	msg = pbo
	return
}
func (goo InvalidPkgPathError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(vmpb.InvalidPkgPathError)
	msg = pbo
	return
}
func (goo *InvalidPkgPathError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *vmpb.InvalidPkgPathError = msg.(*vmpb.InvalidPkgPathError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidPkgPathError) GetTypeURL() (typeURL string) {
	return "/vm.InvalidPkgPathError"
}
func (goo InvalidStmtError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *vmpb.InvalidStmtError
	{
		pbo = new(vmpb.InvalidStmtError)
	}
	msg = pbo
	return
}
func (goo InvalidStmtError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(vmpb.InvalidStmtError)
	msg = pbo
	return
}
func (goo *InvalidStmtError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *vmpb.InvalidStmtError = msg.(*vmpb.InvalidStmtError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidStmtError) GetTypeURL() (typeURL string) {
	return "/vm.InvalidStmtError"
}
func (goo InvalidExprError) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *vmpb.InvalidExprError
	{
		pbo = new(vmpb.InvalidExprError)
	}
	msg = pbo
	return
}
func (goo InvalidExprError) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(vmpb.InvalidExprError)
	msg = pbo
	return
}
func (goo *InvalidExprError) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *vmpb.InvalidExprError = msg.(*vmpb.InvalidExprError)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ InvalidExprError) GetTypeURL() (typeURL string) {
	return "/vm.InvalidExprError"
}
