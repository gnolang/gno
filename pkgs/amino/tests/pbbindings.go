//nolint:stylecheck,unconvert
package tests

import (
	time "time"

	amino "github.com/gnolang/gno/pkgs/amino"
	testspb "github.com/gnolang/gno/pkgs/amino/tests/pb"
	proto "google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (es EmptyStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmptyStruct
	{
		if IsEmptyStructReprEmpty(es) {
			var pbov *testspb.EmptyStruct
			msg = pbov

			return
		}
		pbo = new(testspb.EmptyStruct)
	}
	msg = pbo

	return
}

func (es EmptyStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmptyStruct)
	msg = pbo

	return
}

func (es *EmptyStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmptyStruct = msg.(*testspb.EmptyStruct)
	{
		if pbo != nil {
		}
	}
	return
}

func (es EmptyStruct) GetTypeURL() (typeURL string) {
	return "/tests.EmptyStruct"
}

func IsEmptyStructReprEmpty(goor EmptyStruct) (empty bool) {
	{
		empty = true
	}
	return
}

func (ps PrimitivesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStruct
	{
		if IsPrimitivesStructReprEmpty(ps) {
			var pbov *testspb.PrimitivesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStruct)
		{
			pbo.Int8 = int32(ps.Int8)
		}
		{
			pbo.Int16 = int32(ps.Int16)
		}
		{
			pbo.Int32 = int32(ps.Int32)
		}
		{
			pbo.Int32Fixed = int32(ps.Int32Fixed)
		}
		{
			pbo.Int64 = int64(ps.Int64)
		}
		{
			pbo.Int64Fixed = int64(ps.Int64Fixed)
		}
		{
			pbo.Int = int64(ps.Int)
		}
		{
			pbo.Byte = uint32(ps.Byte)
		}
		{
			pbo.Uint8 = uint32(ps.Uint8)
		}
		{
			pbo.Uint16 = uint32(ps.Uint16)
		}
		{
			pbo.Uint32 = uint32(ps.Uint32)
		}
		{
			pbo.Uint32Fixed = uint32(ps.Uint32Fixed)
		}
		{
			pbo.Uint64 = uint64(ps.Uint64)
		}
		{
			pbo.Uint64Fixed = uint64(ps.Uint64Fixed)
		}
		{
			pbo.Uint = uint64(ps.Uint)
		}
		{
			pbo.Str = string(ps.Str)
		}
		{
			goorl := len(ps.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ps.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(ps.Time) {
				pbo.Time = timestamppb.New(ps.Time)
			}
		}
		{
			if ps.Duration.Nanoseconds() != 0 {
				pbo.Duration = durationpb.New(ps.Duration)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = ps.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}

func (ps PrimitivesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStruct)
	msg = pbo
	return
}

func (ps *PrimitivesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStruct = msg.(*testspb.PrimitivesStruct)
	{
		if pbo != nil {
			{
				(*ps).Int8 = int8(int8(pbo.Int8))
			}
			{
				(*ps).Int16 = int16(int16(pbo.Int16))
			}
			{
				(*ps).Int32 = int32(pbo.Int32)
			}
			{
				(*ps).Int32Fixed = int32(pbo.Int32Fixed)
			}
			{
				(*ps).Int64 = int64(pbo.Int64)
			}
			{
				(*ps).Int64Fixed = int64(pbo.Int64Fixed)
			}
			{
				(*ps).Int = int(int(pbo.Int))
			}
			{
				(*ps).Byte = uint8(uint8(pbo.Byte))
			}
			{
				(*ps).Uint8 = uint8(uint8(pbo.Uint8))
			}
			{
				(*ps).Uint16 = uint16(uint16(pbo.Uint16))
			}
			{
				(*ps).Uint32 = uint32(pbo.Uint32)
			}
			{
				(*ps).Uint32Fixed = uint32(pbo.Uint32Fixed)
			}
			{
				(*ps).Uint64 = uint64(pbo.Uint64)
			}
			{
				(*ps).Uint64Fixed = uint64(pbo.Uint64Fixed)
			}
			{
				(*ps).Uint = uint(uint(pbo.Uint))
			}
			{
				(*ps).Str = string(pbo.Str)
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					(*ps).Bytes = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*ps).Bytes = goors
				}
			}
			{
				(*ps).Time = pbo.Time.AsTime()
			}
			{
				(*ps).Duration = pbo.Duration.AsDuration()
			}
			{
				if pbo.Empty != nil {
					err = (*ps).Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (ps PrimitivesStruct) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStruct"
}

func IsPrimitivesStructReprEmpty(goor PrimitivesStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != int8(0) {
				return false
			}
		}
		{
			if goor.Int16 != int16(0) {
				return false
			}
		}
		{
			if goor.Int32 != int32(0) {
				return false
			}
		}
		{
			if goor.Int32Fixed != int32(0) {
				return false
			}
		}
		{
			if goor.Int64 != int64(0) {
				return false
			}
		}
		{
			if goor.Int64Fixed != int64(0) {
				return false
			}
		}
		{
			if goor.Int != int(0) {
				return false
			}
		}
		{
			if goor.Byte != uint8(0) {
				return false
			}
		}
		{
			if goor.Uint8 != uint8(0) {
				return false
			}
		}
		{
			if goor.Uint16 != uint16(0) {
				return false
			}
		}
		{
			if goor.Uint32 != uint32(0) {
				return false
			}
		}
		{
			if goor.Uint32Fixed != uint32(0) {
				return false
			}
		}
		{
			if goor.Uint64 != uint64(0) {
				return false
			}
		}
		{
			if goor.Uint64Fixed != uint64(0) {
				return false
			}
		}
		{
			if goor.Uint != uint(0) {
				return false
			}
		}
		{
			if goor.Str != string("") {
				return false
			}
		}
		{
			if len(goor.Bytes) != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.Time) {
				return false
			}
		}
		{
			if goor.Duration != time.Duration(0) {
				return false
			}
		}
		{
			e := IsEmptyStructReprEmpty(goor.Empty)
			if e == false {
				return false
			}
		}
	}
	return
}

func (sas ShortArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ShortArraysStruct
	{
		if IsShortArraysStructReprEmpty(sas) {
			var pbov *testspb.ShortArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ShortArraysStruct)
		{
			goorl := len(sas.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				pbos := make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sas.TimeAr[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeAr = pbos
			}
		}
		{
			goorl := len(sas.DurationAr)
			if goorl == 0 {
				pbo.DurationAr = nil
			} else {
				pbos := make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sas.DurationAr[i]
						{
							if goore.Nanoseconds() != 0 {
								pbos[i] = durationpb.New(goore)
							}
						}
					}
				}
				pbo.DurationAr = pbos
			}
		}
	}
	msg = pbo
	return
}

func (sas ShortArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ShortArraysStruct)
	msg = pbo
	return
}

func (sas *ShortArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ShortArraysStruct = msg.(*testspb.ShortArraysStruct)
	{
		if pbo != nil {
			{
				goors := [0]time.Time{}
				for i := 0; i < 0; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsTime()
						}
					}
				}
				(*sas).TimeAr = goors
			}
			{
				goors := [0]time.Duration{}
				for i := 0; i < 0; i += 1 {
					{
						pboe := pbo.DurationAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsDuration()
						}
					}
				}
				(*sas).DurationAr = goors
			}
		}
	}
	return
}

func (sas ShortArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ShortArraysStruct"
}

func IsShortArraysStructReprEmpty(goor ShortArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.TimeAr) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationAr) != 0 {
				return false
			}
		}
	}
	return
}

func (as ArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysStruct
	{
		if IsArraysStructReprEmpty(as) {
			var pbov *testspb.ArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysStruct)
		{
			goorl := len(as.Int8Ar)
			if goorl == 0 {
				pbo.Int8Ar = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int8Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Ar = pbos
			}
		}
		{
			goorl := len(as.Int16Ar)
			if goorl == 0 {
				pbo.Int16Ar = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int16Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Ar = pbos
			}
		}
		{
			goorl := len(as.Int32Ar)
			if goorl == 0 {
				pbo.Int32Ar = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int32Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32Ar = pbos
			}
		}
		{
			goorl := len(as.Int32FixedAr)
			if goorl == 0 {
				pbo.Int32FixedAr = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int32FixedAr[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32FixedAr = pbos
			}
		}
		{
			goorl := len(as.Int64Ar)
			if goorl == 0 {
				pbo.Int64Ar = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int64Ar[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64Ar = pbos
			}
		}
		{
			goorl := len(as.Int64FixedAr)
			if goorl == 0 {
				pbo.Int64FixedAr = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Int64FixedAr[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64FixedAr = pbos
			}
		}
		{
			goorl := len(as.IntAr)
			if goorl == 0 {
				pbo.IntAr = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.IntAr[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntAr = pbos
			}
		}
		{
			goorl := len(as.ByteAr)
			if goorl == 0 {
				pbo.ByteAr = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.ByteAr[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteAr = pbos
			}
		}
		{
			goorl := len(as.Uint8Ar)
			if goorl == 0 {
				pbo.Uint8Ar = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint8Ar[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Ar = pbos
			}
		}
		{
			goorl := len(as.Uint16Ar)
			if goorl == 0 {
				pbo.Uint16Ar = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint16Ar[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Ar = pbos
			}
		}
		{
			goorl := len(as.Uint32Ar)
			if goorl == 0 {
				pbo.Uint32Ar = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint32Ar[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32Ar = pbos
			}
		}
		{
			goorl := len(as.Uint32FixedAr)
			if goorl == 0 {
				pbo.Uint32FixedAr = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint32FixedAr[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32FixedAr = pbos
			}
		}
		{
			goorl := len(as.Uint64Ar)
			if goorl == 0 {
				pbo.Uint64Ar = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint64Ar[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64Ar = pbos
			}
		}
		{
			goorl := len(as.Uint64FixedAr)
			if goorl == 0 {
				pbo.Uint64FixedAr = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.Uint64FixedAr[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64FixedAr = pbos
			}
		}
		{
			goorl := len(as.UintAr)
			if goorl == 0 {
				pbo.UintAr = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.UintAr[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintAr = pbos
			}
		}
		{
			goorl := len(as.StrAr)
			if goorl == 0 {
				pbo.StrAr = nil
			} else {
				pbos := make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.StrAr[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.StrAr = pbos
			}
		}
		{
			goorl := len(as.BytesAr)
			if goorl == 0 {
				pbo.BytesAr = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.BytesAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.BytesAr = pbos
			}
		}
		{
			goorl := len(as.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				pbos := make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.TimeAr[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeAr = pbos
			}
		}
		{
			goorl := len(as.DurationAr)
			if goorl == 0 {
				pbo.DurationAr = nil
			} else {
				pbos := make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.DurationAr[i]
						{
							if goore.Nanoseconds() != 0 {
								pbos[i] = durationpb.New(goore)
							}
						}
					}
				}
				pbo.DurationAr = pbos
			}
		}
		{
			goorl := len(as.EmptyAr)
			if goorl == 0 {
				pbo.EmptyAr = nil
			} else {
				pbos := make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := as.EmptyAr[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*testspb.EmptyStruct)
						}
					}
				}
				pbo.EmptyAr = pbos
			}
		}
	}
	msg = pbo
	return
}

func (as ArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ArraysStruct)
	msg = pbo
	return
}

func (as *ArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysStruct = msg.(*testspb.ArraysStruct)
	{
		if pbo != nil {
			{
				goors := [4]int8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int8Ar[i]
						{
							pboev := pboe
							goors[i] = int8(int8(pboev))
						}
					}
				}
				(*as).Int8Ar = goors
			}
			{
				goors := [4]int16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int16Ar[i]
						{
							pboev := pboe
							goors[i] = int16(int16(pboev))
						}
					}
				}
				(*as).Int16Ar = goors
			}
			{
				goors := [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32Ar[i]
						{
							pboev := pboe
							goors[i] = int32(pboev)
						}
					}
				}
				(*as).Int32Ar = goors
			}
			{
				goors := [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32FixedAr[i]
						{
							pboev := pboe
							goors[i] = int32(pboev)
						}
					}
				}
				(*as).Int32FixedAr = goors
			}
			{
				goors := [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64Ar[i]
						{
							pboev := pboe
							goors[i] = int64(pboev)
						}
					}
				}
				(*as).Int64Ar = goors
			}
			{
				goors := [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64FixedAr[i]
						{
							pboev := pboe
							goors[i] = int64(pboev)
						}
					}
				}
				(*as).Int64FixedAr = goors
			}
			{
				goors := [4]int{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.IntAr[i]
						{
							pboev := pboe
							goors[i] = int(int(pboev))
						}
					}
				}
				(*as).IntAr = goors
			}
			{
				goors := [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.ByteAr[i]
						{
							pboev := pboe
							goors[i] = uint8(uint8(pboev))
						}
					}
				}
				(*as).ByteAr = goors
			}
			{
				goors := [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint8Ar[i]
						{
							pboev := pboe
							goors[i] = uint8(uint8(pboev))
						}
					}
				}
				(*as).Uint8Ar = goors
			}
			{
				goors := [4]uint16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint16Ar[i]
						{
							pboev := pboe
							goors[i] = uint16(uint16(pboev))
						}
					}
				}
				(*as).Uint16Ar = goors
			}
			{
				goors := [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32Ar[i]
						{
							pboev := pboe
							goors[i] = uint32(pboev)
						}
					}
				}
				(*as).Uint32Ar = goors
			}
			{
				goors := [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32FixedAr[i]
						{
							pboev := pboe
							goors[i] = uint32(pboev)
						}
					}
				}
				(*as).Uint32FixedAr = goors
			}
			{
				goors := [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64Ar[i]
						{
							pboev := pboe
							goors[i] = uint64(pboev)
						}
					}
				}
				(*as).Uint64Ar = goors
			}
			{
				goors := [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64FixedAr[i]
						{
							pboev := pboe
							goors[i] = uint64(pboev)
						}
					}
				}
				(*as).Uint64FixedAr = goors
			}
			{
				goors := [4]uint{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.UintAr[i]
						{
							pboev := pboe
							goors[i] = uint(uint(pboev))
						}
					}
				}
				(*as).UintAr = goors
			}
			{
				goors := [4]string{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.StrAr[i]
						{
							pboev := pboe
							goors[i] = string(pboev)
						}
					}
				}
				(*as).StrAr = goors
			}
			{
				goors := [4][]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.BytesAr[i]
						{
							pboev := pboe
							var pbol int = 0
							if pboev != nil {
								pbol = len(pboev)
							}
							if pbol == 0 {
								goors[i] = nil
							} else {
								goors1 := make([]uint8, pbol)
								for i := 0; i < pbol; i += 1 {
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
				(*as).BytesAr = goors
			}
			{
				goors := [4]time.Time{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsTime()
						}
					}
				}
				(*as).TimeAr = goors
			}
			{
				goors := [4]time.Duration{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.DurationAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsDuration()
						}
					}
				}
				(*as).DurationAr = goors
			}
			{
				goors := [4]EmptyStruct{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.EmptyAr[i]
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
				(*as).EmptyAr = goors
			}
		}
	}
	return
}

func (as ArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ArraysStruct"
}

func IsArraysStructReprEmpty(goor ArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.IntAr) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64Ar) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedAr) != 0 {
				return false
			}
		}
		{
			if len(goor.UintAr) != 0 {
				return false
			}
		}
		{
			if len(goor.StrAr) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesAr) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeAr) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationAr) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyAr) != 0 {
				return false
			}
		}
	}
	return
}

func (aas ArraysArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysArraysStruct
	{
		if IsArraysArraysStructReprEmpty(aas) {
			var pbov *testspb.ArraysArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysArraysStruct)
		{
			goorl := len(aas.Int8ArAr)
			if goorl == 0 {
				pbo.Int8ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int8ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8ArAr = pbos
			}
		}
		{
			goorl := len(aas.Int16ArAr)
			if goorl == 0 {
				pbo.Int16ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16ArAr = pbos
			}
		}
		{
			goorl := len(aas.Int32ArAr)
			if goorl == 0 {
				pbo.Int32ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32ArAr = pbos
			}
		}
		{
			goorl := len(aas.Int32FixedArAr)
			if goorl == 0 {
				pbo.Int32FixedArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed32Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed32Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedArAr = pbos
			}
		}
		{
			goorl := len(aas.Int64ArAr)
			if goorl == 0 {
				pbo.Int64ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64ArAr = pbos
			}
		}
		{
			goorl := len(aas.Int64FixedArAr)
			if goorl == 0 {
				pbo.Int64FixedArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed64Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Int64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed64Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedArAr = pbos
			}
		}
		{
			goorl := len(aas.IntArAr)
			if goorl == 0 {
				pbo.IntArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.IntArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntArAr = pbos
			}
		}
		{
			goorl := len(aas.ByteArAr)
			if goorl == 0 {
				pbo.ByteArAr = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.ByteArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.ByteArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint8ArAr)
			if goorl == 0 {
				pbo.Uint8ArAr = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint8ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.Uint8ArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint16ArAr)
			if goorl == 0 {
				pbo.Uint16ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16ArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint32ArAr)
			if goorl == 0 {
				pbo.Uint32ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32ArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint32FixedArAr)
			if goorl == 0 {
				pbo.Uint32FixedArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed32UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed32UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint64ArAr)
			if goorl == 0 {
				pbo.Uint64ArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64ArAr = pbos
			}
		}
		{
			goorl := len(aas.Uint64FixedArAr)
			if goorl == 0 {
				pbo.Uint64FixedArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed64UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.Uint64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed64UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedArAr = pbos
			}
		}
		{
			goorl := len(aas.UintArAr)
			if goorl == 0 {
				pbo.UintArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.UintArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintArAr = pbos
			}
		}
		{
			goorl := len(aas.StrArAr)
			if goorl == 0 {
				pbo.StrArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_StringValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.StrArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = string(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_StringValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrArAr = pbos
			}
		}
		{
			goorl := len(aas.BytesArAr)
			if goorl == 0 {
				pbo.BytesArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.BytesArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												pbos2 := make([]uint8, goorl2)
												for i := 0; i < goorl2; i += 1 {
													{
														goore := goore[i]
														{
															pbos2[i] = byte(goore)
														}
													}
												}
												pbos1[i] = pbos2
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesArAr = pbos
			}
		}
		{
			goorl := len(aas.TimeArAr)
			if goorl == 0 {
				pbo.TimeArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_TimestampList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.TimeArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*timestamppb.Timestamp, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if !amino.IsEmptyTime(goore) {
												pbos1[i] = timestamppb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_TimestampList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeArAr = pbos
			}
		}
		{
			goorl := len(aas.DurationArAr)
			if goorl == 0 {
				pbo.DurationArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_DurationList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.DurationArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*durationpb.Duration, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if goore.Nanoseconds() != 0 {
												pbos1[i] = durationpb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_DurationList{Value: pbos1}
							}
						}
					}
				}
				pbo.DurationArAr = pbos
			}
		}
		{
			goorl := len(aas.EmptyArAr)
			if goorl == 0 {
				pbo.EmptyArAr = nil
			} else {
				pbos := make([]*testspb.TESTS_EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := aas.EmptyArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*testspb.EmptyStruct, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbom := proto.Message(nil)
											pbom, err = goore.ToPBMessage(cdc)
											if err != nil {
												return
											}
											pbos1[i] = pbom.(*testspb.EmptyStruct)
										}
									}
								}
								pbos[i] = &testspb.TESTS_EmptyStructList{Value: pbos1}
							}
						}
					}
				}
				pbo.EmptyArAr = pbos
			}
		}
	}
	msg = pbo
	return
}

func (aas ArraysArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ArraysArraysStruct)
	msg = pbo
	return
}

func (aas *ArraysArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysArraysStruct = msg.(*testspb.ArraysArraysStruct)
	{
		if pbo != nil {
			{
				goors := [2][2]int8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int8ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int8{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int8(int8(pboev))
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int8ArAr = goors
			}
			{
				goors := [2][2]int16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int16ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int16{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int16(int16(pboev))
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int16ArAr = goors
			}
			{
				goors := [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int32{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int32(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int32ArAr = goors
			}
			{
				goors := [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int32{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int32(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int32FixedArAr = goors
			}
			{
				goors := [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int64{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int64(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int64ArAr = goors
			}
			{
				goors := [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int64{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int64(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Int64FixedArAr = goors
			}
			{
				goors := [2][2]int{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.IntArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]int{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = int(int(pboev))
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).IntArAr = goors
			}
			{
				goors := [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.ByteArAr[i]
						{
							pboev := pboe
							goors1 := [2]uint8{}
							for i := 0; i < 2; i += 1 {
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
				(*aas).ByteArAr = goors
			}
			{
				goors := [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint8ArAr[i]
						{
							pboev := pboe
							goors1 := [2]uint8{}
							for i := 0; i < 2; i += 1 {
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
				(*aas).Uint8ArAr = goors
			}
			{
				goors := [2][2]uint16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint16ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint16{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint16(uint16(pboev))
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Uint16ArAr = goors
			}
			{
				goors := [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint32{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint32(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Uint32ArAr = goors
			}
			{
				goors := [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint32{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint32(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Uint32FixedArAr = goors
			}
			{
				goors := [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint64{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint64(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Uint64ArAr = goors
			}
			{
				goors := [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint64{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint64(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).Uint64FixedArAr = goors
			}
			{
				goors := [2][2]uint{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.UintArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]uint{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = uint(uint(pboev))
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).UintArAr = goors
			}
			{
				goors := [2][2]string{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.StrArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]string{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = string(pboev)
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).StrArAr = goors
			}
			{
				goors := [2][2][]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.BytesArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2][]uint8{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											var pbol int = 0
											if pboev != nil {
												pbol = len(pboev)
											}
											if pbol == 0 {
												goors1[i] = nil
											} else {
												goors2 := make([]uint8, pbol)
												for i := 0; i < pbol; i += 1 {
													{
														pboe := pboev[i]
														{
															pboev := pboe
															goors2[i] = uint8(uint8(pboev))
														}
													}
												}
												goors1[i] = goors2
											}
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).BytesArAr = goors
			}
			{
				goors := [2][2]time.Time{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.TimeArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]time.Time{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = pboev.AsTime()
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).TimeArAr = goors
			}
			{
				goors := [2][2]time.Duration{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.DurationArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]time.Duration{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											goors1[i] = pboev.AsDuration()
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).DurationArAr = goors
			}
			{
				goors := [2][2]EmptyStruct{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.EmptyArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								goors1 := [2]EmptyStruct{}
								for i := 0; i < 2; i += 1 {
									{
										pboe := pboev[i]
										{
											pboev := pboe
											if pboev != nil {
												err = goors1[i].FromPBMessage(cdc, pboev)
												if err != nil {
													return
												}
											}
										}
									}
								}
								goors[i] = goors1
							}
						}
					}
				}
				(*aas).EmptyArAr = goors
			}
		}
	}
	return
}

func (aas ArraysArraysStruct) GetTypeURL() (typeURL string) {
	return "/tests.ArraysArraysStruct"
}

func IsArraysArraysStructReprEmpty(goor ArraysArraysStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.IntArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64ArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.UintArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.StrArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationArAr) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyArAr) != 0 {
				return false
			}
		}
	}
	return
}

func (ss SlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesStruct
	{
		if IsSlicesStructReprEmpty(ss) {
			var pbov *testspb.SlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesStruct)
		{
			goorl := len(ss.Int8Sl)
			if goorl == 0 {
				pbo.Int8Sl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int8Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Sl = pbos
			}
		}
		{
			goorl := len(ss.Int16Sl)
			if goorl == 0 {
				pbo.Int16Sl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int16Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Sl = pbos
			}
		}
		{
			goorl := len(ss.Int32Sl)
			if goorl == 0 {
				pbo.Int32Sl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int32Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32Sl = pbos
			}
		}
		{
			goorl := len(ss.Int32FixedSl)
			if goorl == 0 {
				pbo.Int32FixedSl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int32FixedSl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32FixedSl = pbos
			}
		}
		{
			goorl := len(ss.Int64Sl)
			if goorl == 0 {
				pbo.Int64Sl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int64Sl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64Sl = pbos
			}
		}
		{
			goorl := len(ss.Int64FixedSl)
			if goorl == 0 {
				pbo.Int64FixedSl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Int64FixedSl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64FixedSl = pbos
			}
		}
		{
			goorl := len(ss.IntSl)
			if goorl == 0 {
				pbo.IntSl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.IntSl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntSl = pbos
			}
		}
		{
			goorl := len(ss.ByteSl)
			if goorl == 0 {
				pbo.ByteSl = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.ByteSl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteSl = pbos
			}
		}
		{
			goorl := len(ss.Uint8Sl)
			if goorl == 0 {
				pbo.Uint8Sl = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint8Sl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Sl = pbos
			}
		}
		{
			goorl := len(ss.Uint16Sl)
			if goorl == 0 {
				pbo.Uint16Sl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint16Sl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Sl = pbos
			}
		}
		{
			goorl := len(ss.Uint32Sl)
			if goorl == 0 {
				pbo.Uint32Sl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint32Sl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32Sl = pbos
			}
		}
		{
			goorl := len(ss.Uint32FixedSl)
			if goorl == 0 {
				pbo.Uint32FixedSl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint32FixedSl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32FixedSl = pbos
			}
		}
		{
			goorl := len(ss.Uint64Sl)
			if goorl == 0 {
				pbo.Uint64Sl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint64Sl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64Sl = pbos
			}
		}
		{
			goorl := len(ss.Uint64FixedSl)
			if goorl == 0 {
				pbo.Uint64FixedSl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.Uint64FixedSl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64FixedSl = pbos
			}
		}
		{
			goorl := len(ss.UintSl)
			if goorl == 0 {
				pbo.UintSl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.UintSl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintSl = pbos
			}
		}
		{
			goorl := len(ss.StrSl)
			if goorl == 0 {
				pbo.StrSl = nil
			} else {
				pbos := make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.StrSl[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.StrSl = pbos
			}
		}
		{
			goorl := len(ss.BytesSl)
			if goorl == 0 {
				pbo.BytesSl = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.BytesSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.BytesSl = pbos
			}
		}
		{
			goorl := len(ss.TimeSl)
			if goorl == 0 {
				pbo.TimeSl = nil
			} else {
				pbos := make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.TimeSl[i]
						{
							if !amino.IsEmptyTime(goore) {
								pbos[i] = timestamppb.New(goore)
							}
						}
					}
				}
				pbo.TimeSl = pbos
			}
		}
		{
			goorl := len(ss.DurationSl)
			if goorl == 0 {
				pbo.DurationSl = nil
			} else {
				pbos := make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.DurationSl[i]
						{
							if goore.Nanoseconds() != 0 {
								pbos[i] = durationpb.New(goore)
							}
						}
					}
				}
				pbo.DurationSl = pbos
			}
		}
		{
			goorl := len(ss.EmptySl)
			if goorl == 0 {
				pbo.EmptySl = nil
			} else {
				pbos := make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := ss.EmptySl[i]
						{
							pbom := proto.Message(nil)
							pbom, err = goore.ToPBMessage(cdc)
							if err != nil {
								return
							}
							pbos[i] = pbom.(*testspb.EmptyStruct)
						}
					}
				}
				pbo.EmptySl = pbos
			}
		}
	}
	msg = pbo
	return
}

func (ss SlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.SlicesStruct)
	msg = pbo
	return
}

func (ss *SlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesStruct = msg.(*testspb.SlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8Sl != nil {
					pbol = len(pbo.Int8Sl)
				}
				if pbol == 0 {
					(*ss).Int8Sl = nil
				} else {
					goors := make([]int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8Sl[i]
							{
								pboev := pboe
								goors[i] = int8(int8(pboev))
							}
						}
					}
					(*ss).Int8Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16Sl != nil {
					pbol = len(pbo.Int16Sl)
				}
				if pbol == 0 {
					(*ss).Int16Sl = nil
				} else {
					goors := make([]int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16Sl[i]
							{
								pboev := pboe
								goors[i] = int16(int16(pboev))
							}
						}
					}
					(*ss).Int16Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32Sl != nil {
					pbol = len(pbo.Int32Sl)
				}
				if pbol == 0 {
					(*ss).Int32Sl = nil
				} else {
					goors := make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32Sl[i]
							{
								pboev := pboe
								goors[i] = int32(pboev)
							}
						}
					}
					(*ss).Int32Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSl != nil {
					pbol = len(pbo.Int32FixedSl)
				}
				if pbol == 0 {
					(*ss).Int32FixedSl = nil
				} else {
					goors := make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedSl[i]
							{
								pboev := pboe
								goors[i] = int32(pboev)
							}
						}
					}
					(*ss).Int32FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64Sl != nil {
					pbol = len(pbo.Int64Sl)
				}
				if pbol == 0 {
					(*ss).Int64Sl = nil
				} else {
					goors := make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64Sl[i]
							{
								pboev := pboe
								goors[i] = int64(pboev)
							}
						}
					}
					(*ss).Int64Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSl != nil {
					pbol = len(pbo.Int64FixedSl)
				}
				if pbol == 0 {
					(*ss).Int64FixedSl = nil
				} else {
					goors := make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedSl[i]
							{
								pboev := pboe
								goors[i] = int64(pboev)
							}
						}
					}
					(*ss).Int64FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntSl != nil {
					pbol = len(pbo.IntSl)
				}
				if pbol == 0 {
					(*ss).IntSl = nil
				} else {
					goors := make([]int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntSl[i]
							{
								pboev := pboe
								goors[i] = int(int(pboev))
							}
						}
					}
					(*ss).IntSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSl != nil {
					pbol = len(pbo.ByteSl)
				}
				if pbol == 0 {
					(*ss).ByteSl = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ByteSl[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*ss).ByteSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8Sl != nil {
					pbol = len(pbo.Uint8Sl)
				}
				if pbol == 0 {
					(*ss).Uint8Sl = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8Sl[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*ss).Uint8Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16Sl != nil {
					pbol = len(pbo.Uint16Sl)
				}
				if pbol == 0 {
					(*ss).Uint16Sl = nil
				} else {
					goors := make([]uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16Sl[i]
							{
								pboev := pboe
								goors[i] = uint16(uint16(pboev))
							}
						}
					}
					(*ss).Uint16Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32Sl != nil {
					pbol = len(pbo.Uint32Sl)
				}
				if pbol == 0 {
					(*ss).Uint32Sl = nil
				} else {
					goors := make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32Sl[i]
							{
								pboev := pboe
								goors[i] = uint32(pboev)
							}
						}
					}
					(*ss).Uint32Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSl != nil {
					pbol = len(pbo.Uint32FixedSl)
				}
				if pbol == 0 {
					(*ss).Uint32FixedSl = nil
				} else {
					goors := make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedSl[i]
							{
								pboev := pboe
								goors[i] = uint32(pboev)
							}
						}
					}
					(*ss).Uint32FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64Sl != nil {
					pbol = len(pbo.Uint64Sl)
				}
				if pbol == 0 {
					(*ss).Uint64Sl = nil
				} else {
					goors := make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64Sl[i]
							{
								pboev := pboe
								goors[i] = uint64(pboev)
							}
						}
					}
					(*ss).Uint64Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSl != nil {
					pbol = len(pbo.Uint64FixedSl)
				}
				if pbol == 0 {
					(*ss).Uint64FixedSl = nil
				} else {
					goors := make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedSl[i]
							{
								pboev := pboe
								goors[i] = uint64(pboev)
							}
						}
					}
					(*ss).Uint64FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintSl != nil {
					pbol = len(pbo.UintSl)
				}
				if pbol == 0 {
					(*ss).UintSl = nil
				} else {
					goors := make([]uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintSl[i]
							{
								pboev := pboe
								goors[i] = uint(uint(pboev))
							}
						}
					}
					(*ss).UintSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrSl != nil {
					pbol = len(pbo.StrSl)
				}
				if pbol == 0 {
					(*ss).StrSl = nil
				} else {
					goors := make([]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrSl[i]
							{
								pboev := pboe
								goors[i] = string(pboev)
							}
						}
					}
					(*ss).StrSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSl != nil {
					pbol = len(pbo.BytesSl)
				}
				if pbol == 0 {
					(*ss).BytesSl = nil
				} else {
					goors := make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesSl[i]
							{
								pboev := pboe
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									goors[i] = nil
								} else {
									goors1 := make([]uint8, pbol1)
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
					(*ss).BytesSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSl != nil {
					pbol = len(pbo.TimeSl)
				}
				if pbol == 0 {
					(*ss).TimeSl = nil
				} else {
					goors := make([]time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimeSl[i]
							{
								pboev := pboe
								goors[i] = pboev.AsTime()
							}
						}
					}
					(*ss).TimeSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationSl != nil {
					pbol = len(pbo.DurationSl)
				}
				if pbol == 0 {
					(*ss).DurationSl = nil
				} else {
					goors := make([]time.Duration, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.DurationSl[i]
							{
								pboev := pboe
								goors[i] = pboev.AsDuration()
							}
						}
					}
					(*ss).DurationSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySl != nil {
					pbol = len(pbo.EmptySl)
				}
				if pbol == 0 {
					(*ss).EmptySl = nil
				} else {
					goors := make([]EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptySl[i]
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
					(*ss).EmptySl = goors
				}
			}
		}
	}
	return
}

func (ss SlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.SlicesStruct"
}

func IsSlicesStructReprEmpty(goor SlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntSl) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64Sl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeSl) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptySl) != 0 {
				return false
			}
		}
	}
	return
}

func (sss SlicesSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesSlicesStruct
	{
		if IsSlicesSlicesStructReprEmpty(sss) {
			var pbov *testspb.SlicesSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesSlicesStruct)
		{
			goorl := len(sss.Int8SlSl)
			if goorl == 0 {
				pbo.Int8SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int8SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8SlSl = pbos
			}
		}
		{
			goorl := len(sss.Int16SlSl)
			if goorl == 0 {
				pbo.Int16SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16SlSl = pbos
			}
		}
		{
			goorl := len(sss.Int32SlSl)
			if goorl == 0 {
				pbo.Int32SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32SlSl = pbos
			}
		}
		{
			goorl := len(sss.Int32FixedSlSl)
			if goorl == 0 {
				pbo.Int32FixedSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed32Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed32Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedSlSl = pbos
			}
		}
		{
			goorl := len(sss.Int64SlSl)
			if goorl == 0 {
				pbo.Int64SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64SlSl = pbos
			}
		}
		{
			goorl := len(sss.Int64FixedSlSl)
			if goorl == 0 {
				pbo.Int64FixedSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed64Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Int64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed64Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedSlSl = pbos
			}
		}
		{
			goorl := len(sss.IntSlSl)
			if goorl == 0 {
				pbo.IntSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.IntSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntSlSl = pbos
			}
		}
		{
			goorl := len(sss.ByteSlSl)
			if goorl == 0 {
				pbo.ByteSlSl = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.ByteSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.ByteSlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint8SlSl)
			if goorl == 0 {
				pbo.Uint8SlSl = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint8SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint8, goorl1)
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
				pbo.Uint8SlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint16SlSl)
			if goorl == 0 {
				pbo.Uint16SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16SlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint32SlSl)
			if goorl == 0 {
				pbo.Uint32SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32SlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint32FixedSlSl)
			if goorl == 0 {
				pbo.Uint32FixedSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed32UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed32UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedSlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint64SlSl)
			if goorl == 0 {
				pbo.Uint64SlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64SlSl = pbos
			}
		}
		{
			goorl := len(sss.Uint64FixedSlSl)
			if goorl == 0 {
				pbo.Uint64FixedSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_Fixed64UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.Uint64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_Fixed64UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedSlSl = pbos
			}
		}
		{
			goorl := len(sss.UintSlSl)
			if goorl == 0 {
				pbo.UintSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.UintSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintSlSl = pbos
			}
		}
		{
			goorl := len(sss.StrSlSl)
			if goorl == 0 {
				pbo.StrSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_StringValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.StrSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = string(goore)
										}
									}
								}
								pbos[i] = &testspb.TESTS_StringValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrSlSl = pbos
			}
		}
		{
			goorl := len(sss.BytesSlSl)
			if goorl == 0 {
				pbo.BytesSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.BytesSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												pbos2 := make([]uint8, goorl2)
												for i := 0; i < goorl2; i += 1 {
													{
														goore := goore[i]
														{
															pbos2[i] = byte(goore)
														}
													}
												}
												pbos1[i] = pbos2
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesSlSl = pbos
			}
		}
		{
			goorl := len(sss.TimeSlSl)
			if goorl == 0 {
				pbo.TimeSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_TimestampList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.TimeSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*timestamppb.Timestamp, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if !amino.IsEmptyTime(goore) {
												pbos1[i] = timestamppb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_TimestampList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeSlSl = pbos
			}
		}
		{
			goorl := len(sss.DurationSlSl)
			if goorl == 0 {
				pbo.DurationSlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_DurationList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.DurationSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*durationpb.Duration, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											if goore.Nanoseconds() != 0 {
												pbos1[i] = durationpb.New(goore)
											}
										}
									}
								}
								pbos[i] = &testspb.TESTS_DurationList{Value: pbos1}
							}
						}
					}
				}
				pbo.DurationSlSl = pbos
			}
		}
		{
			goorl := len(sss.EmptySlSl)
			if goorl == 0 {
				pbo.EmptySlSl = nil
			} else {
				pbos := make([]*testspb.TESTS_EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := sss.EmptySlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								pbos1 := make([]*testspb.EmptyStruct, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbom := proto.Message(nil)
											pbom, err = goore.ToPBMessage(cdc)
											if err != nil {
												return
											}
											pbos1[i] = pbom.(*testspb.EmptyStruct)
										}
									}
								}
								pbos[i] = &testspb.TESTS_EmptyStructList{Value: pbos1}
							}
						}
					}
				}
				pbo.EmptySlSl = pbos
			}
		}
	}
	msg = pbo
	return
}

func (sss SlicesSlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.SlicesSlicesStruct)
	msg = pbo
	return
}

func (sss *SlicesSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesSlicesStruct = msg.(*testspb.SlicesSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8SlSl != nil {
					pbol = len(pbo.Int8SlSl)
				}
				if pbol == 0 {
					(*sss).Int8SlSl = nil
				} else {
					goors := make([][]int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int8, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int8(int8(pboev))
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int8SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16SlSl != nil {
					pbol = len(pbo.Int16SlSl)
				}
				if pbol == 0 {
					(*sss).Int16SlSl = nil
				} else {
					goors := make([][]int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int16, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int16(int16(pboev))
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int16SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32SlSl != nil {
					pbol = len(pbo.Int32SlSl)
				}
				if pbol == 0 {
					(*sss).Int32SlSl = nil
				} else {
					goors := make([][]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int32, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int32(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int32SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSlSl != nil {
					pbol = len(pbo.Int32FixedSlSl)
				}
				if pbol == 0 {
					(*sss).Int32FixedSlSl = nil
				} else {
					goors := make([][]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int32, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int32(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int32FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64SlSl != nil {
					pbol = len(pbo.Int64SlSl)
				}
				if pbol == 0 {
					(*sss).Int64SlSl = nil
				} else {
					goors := make([][]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int64, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int64(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int64SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSlSl != nil {
					pbol = len(pbo.Int64FixedSlSl)
				}
				if pbol == 0 {
					(*sss).Int64FixedSlSl = nil
				} else {
					goors := make([][]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int64, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int64(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Int64FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntSlSl != nil {
					pbol = len(pbo.IntSlSl)
				}
				if pbol == 0 {
					(*sss).IntSlSl = nil
				} else {
					goors := make([][]int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]int, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = int(int(pboev))
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).IntSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSlSl != nil {
					pbol = len(pbo.ByteSlSl)
				}
				if pbol == 0 {
					(*sss).ByteSlSl = nil
				} else {
					goors := make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ByteSlSl[i]
							{
								pboev := pboe
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									goors[i] = nil
								} else {
									goors1 := make([]uint8, pbol1)
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
					(*sss).ByteSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8SlSl != nil {
					pbol = len(pbo.Uint8SlSl)
				}
				if pbol == 0 {
					(*sss).Uint8SlSl = nil
				} else {
					goors := make([][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8SlSl[i]
							{
								pboev := pboe
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									goors[i] = nil
								} else {
									goors1 := make([]uint8, pbol1)
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
					(*sss).Uint8SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16SlSl != nil {
					pbol = len(pbo.Uint16SlSl)
				}
				if pbol == 0 {
					(*sss).Uint16SlSl = nil
				} else {
					goors := make([][]uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint16, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint16(uint16(pboev))
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Uint16SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32SlSl != nil {
					pbol = len(pbo.Uint32SlSl)
				}
				if pbol == 0 {
					(*sss).Uint32SlSl = nil
				} else {
					goors := make([][]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint32, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint32(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Uint32SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSlSl != nil {
					pbol = len(pbo.Uint32FixedSlSl)
				}
				if pbol == 0 {
					(*sss).Uint32FixedSlSl = nil
				} else {
					goors := make([][]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint32, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint32(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Uint32FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64SlSl != nil {
					pbol = len(pbo.Uint64SlSl)
				}
				if pbol == 0 {
					(*sss).Uint64SlSl = nil
				} else {
					goors := make([][]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64SlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint64, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint64(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Uint64SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSlSl != nil {
					pbol = len(pbo.Uint64FixedSlSl)
				}
				if pbol == 0 {
					(*sss).Uint64FixedSlSl = nil
				} else {
					goors := make([][]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint64, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint64(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).Uint64FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintSlSl != nil {
					pbol = len(pbo.UintSlSl)
				}
				if pbol == 0 {
					(*sss).UintSlSl = nil
				} else {
					goors := make([][]uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]uint, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = uint(uint(pboev))
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).UintSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrSlSl != nil {
					pbol = len(pbo.StrSlSl)
				}
				if pbol == 0 {
					(*sss).StrSlSl = nil
				} else {
					goors := make([][]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]string, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = string(pboev)
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).StrSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSlSl != nil {
					pbol = len(pbo.BytesSlSl)
				}
				if pbol == 0 {
					(*sss).BytesSlSl = nil
				} else {
					goors := make([][][]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([][]uint8, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													var pbol2 int = 0
													if pboev != nil {
														pbol2 = len(pboev)
													}
													if pbol2 == 0 {
														goors1[i] = nil
													} else {
														goors2 := make([]uint8, pbol2)
														for i := 0; i < pbol2; i += 1 {
															{
																pboe := pboev[i]
																{
																	pboev := pboe
																	goors2[i] = uint8(uint8(pboev))
																}
															}
														}
														goors1[i] = goors2
													}
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).BytesSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSlSl != nil {
					pbol = len(pbo.TimeSlSl)
				}
				if pbol == 0 {
					(*sss).TimeSlSl = nil
				} else {
					goors := make([][]time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimeSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]time.Time, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = pboev.AsTime()
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).TimeSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationSlSl != nil {
					pbol = len(pbo.DurationSlSl)
				}
				if pbol == 0 {
					(*sss).DurationSlSl = nil
				} else {
					goors := make([][]time.Duration, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.DurationSlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]time.Duration, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													goors1[i] = pboev.AsDuration()
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).DurationSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySlSl != nil {
					pbol = len(pbo.EmptySlSl)
				}
				if pbol == 0 {
					(*sss).EmptySlSl = nil
				} else {
					goors := make([][]EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptySlSl[i]
							if pboe != nil {
								{
									pboev := pboe.Value
									var pbol1 int = 0
									if pboev != nil {
										pbol1 = len(pboev)
									}
									if pbol1 == 0 {
										goors[i] = nil
									} else {
										goors1 := make([]EmptyStruct, pbol1)
										for i := 0; i < pbol1; i += 1 {
											{
												pboe := pboev[i]
												{
													pboev := pboe
													if pboev != nil {
														err = goors1[i].FromPBMessage(cdc, pboev)
														if err != nil {
															return
														}
													}
												}
											}
										}
										goors[i] = goors1
									}
								}
							}
						}
					}
					(*sss).EmptySlSl = goors
				}
			}
		}
	}
	return
}

func (sss SlicesSlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.SlicesSlicesStruct"
}

func IsSlicesSlicesStructReprEmpty(goor SlicesSlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.ByteSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64SlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimeSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationSlSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptySlSl) != 0 {
				return false
			}
		}
	}
	return
}

func (ps PointersStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointersStruct
	{
		if IsPointersStructReprEmpty(ps) {
			var pbov *testspb.PointersStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointersStruct)
		{
			if ps.Int8Pt != nil {
				dgoor := *ps.Int8Pt
				dgoor = dgoor
				pbo.Int8Pt = int32(dgoor)
			}
		}
		{
			if ps.Int16Pt != nil {
				dgoor := *ps.Int16Pt
				dgoor = dgoor
				pbo.Int16Pt = int32(dgoor)
			}
		}
		{
			if ps.Int32Pt != nil {
				dgoor := *ps.Int32Pt
				dgoor = dgoor
				pbo.Int32Pt = int32(dgoor)
			}
		}
		{
			if ps.Int32FixedPt != nil {
				dgoor := *ps.Int32FixedPt
				dgoor = dgoor
				pbo.Int32FixedPt = int32(dgoor)
			}
		}
		{
			if ps.Int64Pt != nil {
				dgoor := *ps.Int64Pt
				dgoor = dgoor
				pbo.Int64Pt = int64(dgoor)
			}
		}
		{
			if ps.Int64FixedPt != nil {
				dgoor := *ps.Int64FixedPt
				dgoor = dgoor
				pbo.Int64FixedPt = int64(dgoor)
			}
		}
		{
			if ps.IntPt != nil {
				dgoor := *ps.IntPt
				dgoor = dgoor
				pbo.IntPt = int64(dgoor)
			}
		}
		{
			if ps.BytePt != nil {
				dgoor := *ps.BytePt
				dgoor = dgoor
				pbo.BytePt = uint32(dgoor)
			}
		}
		{
			if ps.Uint8Pt != nil {
				dgoor := *ps.Uint8Pt
				dgoor = dgoor
				pbo.Uint8Pt = uint32(dgoor)
			}
		}
		{
			if ps.Uint16Pt != nil {
				dgoor := *ps.Uint16Pt
				dgoor = dgoor
				pbo.Uint16Pt = uint32(dgoor)
			}
		}
		{
			if ps.Uint32Pt != nil {
				dgoor := *ps.Uint32Pt
				dgoor = dgoor
				pbo.Uint32Pt = uint32(dgoor)
			}
		}
		{
			if ps.Uint32FixedPt != nil {
				dgoor := *ps.Uint32FixedPt
				dgoor = dgoor
				pbo.Uint32FixedPt = uint32(dgoor)
			}
		}
		{
			if ps.Uint64Pt != nil {
				dgoor := *ps.Uint64Pt
				dgoor = dgoor
				pbo.Uint64Pt = uint64(dgoor)
			}
		}
		{
			if ps.Uint64FixedPt != nil {
				dgoor := *ps.Uint64FixedPt
				dgoor = dgoor
				pbo.Uint64FixedPt = uint64(dgoor)
			}
		}
		{
			if ps.UintPt != nil {
				dgoor := *ps.UintPt
				dgoor = dgoor
				pbo.UintPt = uint64(dgoor)
			}
		}
		{
			if ps.StrPt != nil {
				dgoor := *ps.StrPt
				dgoor = dgoor
				pbo.StrPt = string(dgoor)
			}
		}
		{
			if ps.BytesPt != nil {
				dgoor := *ps.BytesPt
				dgoor = dgoor
				goorl := len(dgoor)
				if goorl == 0 {
					pbo.BytesPt = nil
				} else {
					pbos := make([]uint8, goorl)
					for i := 0; i < goorl; i += 1 {
						{
							goore := dgoor[i]
							{
								pbos[i] = byte(goore)
							}
						}
					}
					pbo.BytesPt = pbos
				}
			}
		}
		{
			if ps.TimePt != nil {
				dgoor := *ps.TimePt
				dgoor = dgoor
				pbo.TimePt = timestamppb.New(dgoor)
			}
		}
		{
			if ps.DurationPt != nil {
				dgoor := *ps.DurationPt
				dgoor = dgoor
				pbo.DurationPt = durationpb.New(dgoor)
			}
		}
		{
			if ps.EmptyPt != nil {
				pbom := proto.Message(nil)
				pbom, err = ps.EmptyPt.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.EmptyPt = pbom.(*testspb.EmptyStruct)
				if pbo.EmptyPt == nil {
					pbo.EmptyPt = new(testspb.EmptyStruct)
				}
			}
		}
	}
	msg = pbo
	return
}

func (ps PointersStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PointersStruct)
	msg = pbo
	return
}

func (ps *PointersStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointersStruct = msg.(*testspb.PointersStruct)
	{
		if pbo != nil {
			{
				(*ps).Int8Pt = new(int8)
				*(*ps).Int8Pt = int8(int8(pbo.Int8Pt))
			}
			{
				(*ps).Int16Pt = new(int16)
				*(*ps).Int16Pt = int16(int16(pbo.Int16Pt))
			}
			{
				(*ps).Int32Pt = new(int32)
				*(*ps).Int32Pt = int32(pbo.Int32Pt)
			}
			{
				(*ps).Int32FixedPt = new(int32)
				*(*ps).Int32FixedPt = int32(pbo.Int32FixedPt)
			}
			{
				(*ps).Int64Pt = new(int64)
				*(*ps).Int64Pt = int64(pbo.Int64Pt)
			}
			{
				(*ps).Int64FixedPt = new(int64)
				*(*ps).Int64FixedPt = int64(pbo.Int64FixedPt)
			}
			{
				(*ps).IntPt = new(int)
				*(*ps).IntPt = int(int(pbo.IntPt))
			}
			{
				(*ps).BytePt = new(uint8)
				*(*ps).BytePt = uint8(uint8(pbo.BytePt))
			}
			{
				(*ps).Uint8Pt = new(uint8)
				*(*ps).Uint8Pt = uint8(uint8(pbo.Uint8Pt))
			}
			{
				(*ps).Uint16Pt = new(uint16)
				*(*ps).Uint16Pt = uint16(uint16(pbo.Uint16Pt))
			}
			{
				(*ps).Uint32Pt = new(uint32)
				*(*ps).Uint32Pt = uint32(pbo.Uint32Pt)
			}
			{
				(*ps).Uint32FixedPt = new(uint32)
				*(*ps).Uint32FixedPt = uint32(pbo.Uint32FixedPt)
			}
			{
				(*ps).Uint64Pt = new(uint64)
				*(*ps).Uint64Pt = uint64(pbo.Uint64Pt)
			}
			{
				(*ps).Uint64FixedPt = new(uint64)
				*(*ps).Uint64FixedPt = uint64(pbo.Uint64FixedPt)
			}
			{
				(*ps).UintPt = new(uint)
				*(*ps).UintPt = uint(uint(pbo.UintPt))
			}
			{
				(*ps).StrPt = new(string)
				*(*ps).StrPt = string(pbo.StrPt)
			}
			{
				(*ps).BytesPt = new([]uint8)
				var pbol int = 0
				if pbo.BytesPt != nil {
					pbol = len(pbo.BytesPt)
				}
				if pbol == 0 {
					*(*ps).BytesPt = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesPt[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					*(*ps).BytesPt = goors
				}
			}
			{
				(*ps).TimePt = new(time.Time)
				*(*ps).TimePt = pbo.TimePt.AsTime()
			}
			{
				(*ps).DurationPt = new(time.Duration)
				*(*ps).DurationPt = pbo.DurationPt.AsDuration()
			}
			{
				if pbo.EmptyPt != nil {
					(*ps).EmptyPt = new(EmptyStruct)
					err = (*ps).EmptyPt.FromPBMessage(cdc, pbo.EmptyPt)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (ps PointersStruct) GetTypeURL() (typeURL string) {
	return "/tests.PointersStruct"
}

func IsPointersStructReprEmpty(goor PointersStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8Pt != nil {
				dgoor := *goor.Int8Pt
				dgoor = dgoor
				if dgoor != int8(0) {
					return false
				}
			}
		}
		{
			if goor.Int16Pt != nil {
				dgoor := *goor.Int16Pt
				dgoor = dgoor
				if dgoor != int16(0) {
					return false
				}
			}
		}
		{
			if goor.Int32Pt != nil {
				dgoor := *goor.Int32Pt
				dgoor = dgoor
				if dgoor != int32(0) {
					return false
				}
			}
		}
		{
			if goor.Int32FixedPt != nil {
				dgoor := *goor.Int32FixedPt
				dgoor = dgoor
				if dgoor != int32(0) {
					return false
				}
			}
		}
		{
			if goor.Int64Pt != nil {
				dgoor := *goor.Int64Pt
				dgoor = dgoor
				if dgoor != int64(0) {
					return false
				}
			}
		}
		{
			if goor.Int64FixedPt != nil {
				dgoor := *goor.Int64FixedPt
				dgoor = dgoor
				if dgoor != int64(0) {
					return false
				}
			}
		}
		{
			if goor.IntPt != nil {
				dgoor := *goor.IntPt
				dgoor = dgoor
				if dgoor != int(0) {
					return false
				}
			}
		}
		{
			if goor.BytePt != nil {
				dgoor := *goor.BytePt
				dgoor = dgoor
				if dgoor != uint8(0) {
					return false
				}
			}
		}
		{
			if goor.Uint8Pt != nil {
				dgoor := *goor.Uint8Pt
				dgoor = dgoor
				if dgoor != uint8(0) {
					return false
				}
			}
		}
		{
			if goor.Uint16Pt != nil {
				dgoor := *goor.Uint16Pt
				dgoor = dgoor
				if dgoor != uint16(0) {
					return false
				}
			}
		}
		{
			if goor.Uint32Pt != nil {
				dgoor := *goor.Uint32Pt
				dgoor = dgoor
				if dgoor != uint32(0) {
					return false
				}
			}
		}
		{
			if goor.Uint32FixedPt != nil {
				dgoor := *goor.Uint32FixedPt
				dgoor = dgoor
				if dgoor != uint32(0) {
					return false
				}
			}
		}
		{
			if goor.Uint64Pt != nil {
				dgoor := *goor.Uint64Pt
				dgoor = dgoor
				if dgoor != uint64(0) {
					return false
				}
			}
		}
		{
			if goor.Uint64FixedPt != nil {
				dgoor := *goor.Uint64FixedPt
				dgoor = dgoor
				if dgoor != uint64(0) {
					return false
				}
			}
		}
		{
			if goor.UintPt != nil {
				dgoor := *goor.UintPt
				dgoor = dgoor
				if dgoor != uint(0) {
					return false
				}
			}
		}
		{
			if goor.StrPt != nil {
				dgoor := *goor.StrPt
				dgoor = dgoor
				if dgoor != string("") {
					return false
				}
			}
		}
		{
			if goor.BytesPt != nil {
				dgoor := *goor.BytesPt
				dgoor = dgoor
				if len(dgoor) != 0 {
					return false
				}
			}
		}
		{
			if goor.TimePt != nil {
				return false
			}
		}
		{
			if goor.DurationPt != nil {
				dgoor := *goor.DurationPt
				dgoor = dgoor
				if dgoor != time.Duration(0) {
					return false
				}
			}
		}
		{
			if goor.EmptyPt != nil {
				return false
			}
		}
	}
	return
}

func (pss PointerSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointerSlicesStruct
	{
		if IsPointerSlicesStructReprEmpty(pss) {
			var pbov *testspb.PointerSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointerSlicesStruct)
		{
			goorl := len(pss.Int8PtSl)
			if goorl == 0 {
				pbo.Int8PtSl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int8PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int8PtSl = pbos
			}
		}
		{
			goorl := len(pss.Int16PtSl)
			if goorl == 0 {
				pbo.Int16PtSl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int16PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int16PtSl = pbos
			}
		}
		{
			goorl := len(pss.Int32PtSl)
			if goorl == 0 {
				pbo.Int32PtSl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int32PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int32PtSl = pbos
			}
		}
		{
			goorl := len(pss.Int32FixedPtSl)
			if goorl == 0 {
				pbo.Int32FixedPtSl = nil
			} else {
				pbos := make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int32FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int32(dgoor)
							}
						}
					}
				}
				pbo.Int32FixedPtSl = pbos
			}
		}
		{
			goorl := len(pss.Int64PtSl)
			if goorl == 0 {
				pbo.Int64PtSl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int64PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int64(dgoor)
							}
						}
					}
				}
				pbo.Int64PtSl = pbos
			}
		}
		{
			goorl := len(pss.Int64FixedPtSl)
			if goorl == 0 {
				pbo.Int64FixedPtSl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Int64FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int64(dgoor)
							}
						}
					}
				}
				pbo.Int64FixedPtSl = pbos
			}
		}
		{
			goorl := len(pss.IntPtSl)
			if goorl == 0 {
				pbo.IntPtSl = nil
			} else {
				pbos := make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.IntPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = int64(dgoor)
							}
						}
					}
				}
				pbo.IntPtSl = pbos
			}
		}
		{
			goorl := len(pss.BytePtSl)
			if goorl == 0 {
				pbo.BytePtSl = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.BytePtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = byte(dgoor)
							}
						}
					}
				}
				pbo.BytePtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint8PtSl)
			if goorl == 0 {
				pbo.Uint8PtSl = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint8PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = byte(dgoor)
							}
						}
					}
				}
				pbo.Uint8PtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint16PtSl)
			if goorl == 0 {
				pbo.Uint16PtSl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint16PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint32(dgoor)
							}
						}
					}
				}
				pbo.Uint16PtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint32PtSl)
			if goorl == 0 {
				pbo.Uint32PtSl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint32PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint32(dgoor)
							}
						}
					}
				}
				pbo.Uint32PtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint32FixedPtSl)
			if goorl == 0 {
				pbo.Uint32FixedPtSl = nil
			} else {
				pbos := make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint32FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint32(dgoor)
							}
						}
					}
				}
				pbo.Uint32FixedPtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint64PtSl)
			if goorl == 0 {
				pbo.Uint64PtSl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint64PtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint64(dgoor)
							}
						}
					}
				}
				pbo.Uint64PtSl = pbos
			}
		}
		{
			goorl := len(pss.Uint64FixedPtSl)
			if goorl == 0 {
				pbo.Uint64FixedPtSl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.Uint64FixedPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint64(dgoor)
							}
						}
					}
				}
				pbo.Uint64FixedPtSl = pbos
			}
		}
		{
			goorl := len(pss.UintPtSl)
			if goorl == 0 {
				pbo.UintPtSl = nil
			} else {
				pbos := make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.UintPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = uint64(dgoor)
							}
						}
					}
				}
				pbo.UintPtSl = pbos
			}
		}
		{
			goorl := len(pss.StrPtSl)
			if goorl == 0 {
				pbo.StrPtSl = nil
			} else {
				pbos := make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.StrPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = string(dgoor)
							}
						}
					}
				}
				pbo.StrPtSl = pbos
			}
		}
		{
			goorl := len(pss.BytesPtSl)
			if goorl == 0 {
				pbo.BytesPtSl = nil
			} else {
				pbos := make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.BytesPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								goorl1 := len(dgoor)
								if goorl1 == 0 {
									pbos[i] = nil
								} else {
									pbos1 := make([]uint8, goorl1)
									for i := 0; i < goorl1; i += 1 {
										{
											goore := dgoor[i]
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
				}
				pbo.BytesPtSl = pbos
			}
		}
		{
			goorl := len(pss.TimePtSl)
			if goorl == 0 {
				pbo.TimePtSl = nil
			} else {
				pbos := make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.TimePtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = timestamppb.New(dgoor)
							}
						}
					}
				}
				pbo.TimePtSl = pbos
			}
		}
		{
			goorl := len(pss.DurationPtSl)
			if goorl == 0 {
				pbo.DurationPtSl = nil
			} else {
				pbos := make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.DurationPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								pbos[i] = durationpb.New(dgoor)
							}
						}
					}
				}
				pbo.DurationPtSl = pbos
			}
		}
		{
			goorl := len(pss.EmptyPtSl)
			if goorl == 0 {
				pbo.EmptyPtSl = nil
			} else {
				pbos := make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := pss.EmptyPtSl[i]
						{
							if goore != nil {
								pbom := proto.Message(nil)
								pbom, err = goore.ToPBMessage(cdc)
								if err != nil {
									return
								}
								pbos[i] = pbom.(*testspb.EmptyStruct)
								if pbos[i] == nil {
									pbos[i] = new(testspb.EmptyStruct)
								}
							}
						}
					}
				}
				pbo.EmptyPtSl = pbos
			}
		}
	}
	msg = pbo
	return
}

func (pss PointerSlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PointerSlicesStruct)
	msg = pbo
	return
}

func (pss *PointerSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointerSlicesStruct = msg.(*testspb.PointerSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8PtSl != nil {
					pbol = len(pbo.Int8PtSl)
				}
				if pbol == 0 {
					(*pss).Int8PtSl = nil
				} else {
					goors := make([]*int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8PtSl[i]
							{
								pboev := pboe
								goors[i] = new(int8)
								*goors[i] = int8(int8(pboev))
							}
						}
					}
					(*pss).Int8PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16PtSl != nil {
					pbol = len(pbo.Int16PtSl)
				}
				if pbol == 0 {
					(*pss).Int16PtSl = nil
				} else {
					goors := make([]*int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16PtSl[i]
							{
								pboev := pboe
								goors[i] = new(int16)
								*goors[i] = int16(int16(pboev))
							}
						}
					}
					(*pss).Int16PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32PtSl != nil {
					pbol = len(pbo.Int32PtSl)
				}
				if pbol == 0 {
					(*pss).Int32PtSl = nil
				} else {
					goors := make([]*int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32PtSl[i]
							{
								pboev := pboe
								goors[i] = new(int32)
								*goors[i] = int32(pboev)
							}
						}
					}
					(*pss).Int32PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedPtSl != nil {
					pbol = len(pbo.Int32FixedPtSl)
				}
				if pbol == 0 {
					(*pss).Int32FixedPtSl = nil
				} else {
					goors := make([]*int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedPtSl[i]
							{
								pboev := pboe
								goors[i] = new(int32)
								*goors[i] = int32(pboev)
							}
						}
					}
					(*pss).Int32FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64PtSl != nil {
					pbol = len(pbo.Int64PtSl)
				}
				if pbol == 0 {
					(*pss).Int64PtSl = nil
				} else {
					goors := make([]*int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64PtSl[i]
							{
								pboev := pboe
								goors[i] = new(int64)
								*goors[i] = int64(pboev)
							}
						}
					}
					(*pss).Int64PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedPtSl != nil {
					pbol = len(pbo.Int64FixedPtSl)
				}
				if pbol == 0 {
					(*pss).Int64FixedPtSl = nil
				} else {
					goors := make([]*int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedPtSl[i]
							{
								pboev := pboe
								goors[i] = new(int64)
								*goors[i] = int64(pboev)
							}
						}
					}
					(*pss).Int64FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntPtSl != nil {
					pbol = len(pbo.IntPtSl)
				}
				if pbol == 0 {
					(*pss).IntPtSl = nil
				} else {
					goors := make([]*int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntPtSl[i]
							{
								pboev := pboe
								goors[i] = new(int)
								*goors[i] = int(int(pboev))
							}
						}
					}
					(*pss).IntPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytePtSl != nil {
					pbol = len(pbo.BytePtSl)
				}
				if pbol == 0 {
					(*pss).BytePtSl = nil
				} else {
					goors := make([]*uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytePtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint8)
								*goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*pss).BytePtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8PtSl != nil {
					pbol = len(pbo.Uint8PtSl)
				}
				if pbol == 0 {
					(*pss).Uint8PtSl = nil
				} else {
					goors := make([]*uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8PtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint8)
								*goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*pss).Uint8PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16PtSl != nil {
					pbol = len(pbo.Uint16PtSl)
				}
				if pbol == 0 {
					(*pss).Uint16PtSl = nil
				} else {
					goors := make([]*uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16PtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint16)
								*goors[i] = uint16(uint16(pboev))
							}
						}
					}
					(*pss).Uint16PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32PtSl != nil {
					pbol = len(pbo.Uint32PtSl)
				}
				if pbol == 0 {
					(*pss).Uint32PtSl = nil
				} else {
					goors := make([]*uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32PtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint32)
								*goors[i] = uint32(pboev)
							}
						}
					}
					(*pss).Uint32PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedPtSl != nil {
					pbol = len(pbo.Uint32FixedPtSl)
				}
				if pbol == 0 {
					(*pss).Uint32FixedPtSl = nil
				} else {
					goors := make([]*uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedPtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint32)
								*goors[i] = uint32(pboev)
							}
						}
					}
					(*pss).Uint32FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64PtSl != nil {
					pbol = len(pbo.Uint64PtSl)
				}
				if pbol == 0 {
					(*pss).Uint64PtSl = nil
				} else {
					goors := make([]*uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64PtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint64)
								*goors[i] = uint64(pboev)
							}
						}
					}
					(*pss).Uint64PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedPtSl != nil {
					pbol = len(pbo.Uint64FixedPtSl)
				}
				if pbol == 0 {
					(*pss).Uint64FixedPtSl = nil
				} else {
					goors := make([]*uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedPtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint64)
								*goors[i] = uint64(pboev)
							}
						}
					}
					(*pss).Uint64FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintPtSl != nil {
					pbol = len(pbo.UintPtSl)
				}
				if pbol == 0 {
					(*pss).UintPtSl = nil
				} else {
					goors := make([]*uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintPtSl[i]
							{
								pboev := pboe
								goors[i] = new(uint)
								*goors[i] = uint(uint(pboev))
							}
						}
					}
					(*pss).UintPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrPtSl != nil {
					pbol = len(pbo.StrPtSl)
				}
				if pbol == 0 {
					(*pss).StrPtSl = nil
				} else {
					goors := make([]*string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrPtSl[i]
							{
								pboev := pboe
								goors[i] = new(string)
								*goors[i] = string(pboev)
							}
						}
					}
					(*pss).StrPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesPtSl != nil {
					pbol = len(pbo.BytesPtSl)
				}
				if pbol == 0 {
					(*pss).BytesPtSl = nil
				} else {
					goors := make([]*[]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesPtSl[i]
							{
								pboev := pboe
								goors[i] = new([]uint8)
								var pbol1 int = 0
								if pboev != nil {
									pbol1 = len(pboev)
								}
								if pbol1 == 0 {
									*goors[i] = nil
								} else {
									goors1 := make([]uint8, pbol1)
									for i := 0; i < pbol1; i += 1 {
										{
											pboe := pboev[i]
											{
												pboev := pboe
												goors1[i] = uint8(uint8(pboev))
											}
										}
									}
									*goors[i] = goors1
								}
							}
						}
					}
					(*pss).BytesPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimePtSl != nil {
					pbol = len(pbo.TimePtSl)
				}
				if pbol == 0 {
					(*pss).TimePtSl = nil
				} else {
					goors := make([]*time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimePtSl[i]
							{
								pboev := pboe
								goors[i] = new(time.Time)
								*goors[i] = pboev.AsTime()
							}
						}
					}
					(*pss).TimePtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationPtSl != nil {
					pbol = len(pbo.DurationPtSl)
				}
				if pbol == 0 {
					(*pss).DurationPtSl = nil
				} else {
					goors := make([]*time.Duration, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.DurationPtSl[i]
							{
								pboev := pboe
								goors[i] = new(time.Duration)
								*goors[i] = pboev.AsDuration()
							}
						}
					}
					(*pss).DurationPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptyPtSl != nil {
					pbol = len(pbo.EmptyPtSl)
				}
				if pbol == 0 {
					(*pss).EmptyPtSl = nil
				} else {
					goors := make([]*EmptyStruct, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.EmptyPtSl[i]
							{
								pboev := pboe
								if pboev != nil {
									goors[i] = new(EmptyStruct)
									err = goors[i].FromPBMessage(cdc, pboev)
									if err != nil {
										return
									}
								}
							}
						}
					}
					(*pss).EmptyPtSl = goors
				}
			}
		}
	}
	return
}

func (pss PointerSlicesStruct) GetTypeURL() (typeURL string) {
	return "/tests.PointerSlicesStruct"
}

func IsPointerSlicesStructReprEmpty(goor PointerSlicesStruct) (empty bool) {
	{
		empty = true
		{
			if len(goor.Int8PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int16PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int32FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Int64FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.IntPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytePtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint8PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint16PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint32FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64PtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.Uint64FixedPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.UintPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.StrPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.BytesPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.TimePtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.DurationPtSl) != 0 {
				return false
			}
		}
		{
			if len(goor.EmptyPtSl) != 0 {
				return false
			}
		}
	}
	return
}

func (cs ComplexSt) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ComplexSt
	{
		if IsComplexStReprEmpty(cs) {
			var pbov *testspb.ComplexSt
			msg = pbov
			return
		}
		pbo = new(testspb.ComplexSt)
		{
			pbom := proto.Message(nil)
			pbom, err = cs.PrField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrField = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = cs.ArField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArField = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = cs.SlField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlField = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = cs.PtField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PtField = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}

func (cs ComplexSt) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ComplexSt)
	msg = pbo
	return
}

func (cs *ComplexSt) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ComplexSt = msg.(*testspb.ComplexSt)
	{
		if pbo != nil {
			{
				if pbo.PrField != nil {
					err = (*cs).PrField.FromPBMessage(cdc, pbo.PrField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArField != nil {
					err = (*cs).ArField.FromPBMessage(cdc, pbo.ArField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlField != nil {
					err = (*cs).SlField.FromPBMessage(cdc, pbo.SlField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PtField != nil {
					err = (*cs).PtField.FromPBMessage(cdc, pbo.PtField)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (cs ComplexSt) GetTypeURL() (typeURL string) {
	return "/tests.ComplexSt"
}

func IsComplexStReprEmpty(goor ComplexSt) (empty bool) {
	{
		empty = true
		{
			e := IsPrimitivesStructReprEmpty(goor.PrField)
			if e == false {
				return false
			}
		}
		{
			e := IsArraysStructReprEmpty(goor.ArField)
			if e == false {
				return false
			}
		}
		{
			e := IsSlicesStructReprEmpty(goor.SlField)
			if e == false {
				return false
			}
		}
		{
			e := IsPointersStructReprEmpty(goor.PtField)
			if e == false {
				return false
			}
		}
	}
	return
}

func (es EmbeddedSt1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt1
	{
		if IsEmbeddedSt1ReprEmpty(es) {
			var pbov *testspb.EmbeddedSt1
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt1)
		{
			pbom := proto.Message(nil)
			pbom, err = es.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
	}
	msg = pbo
	return
}

func (es EmbeddedSt1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt1)
	msg = pbo
	return
}

func (es *EmbeddedSt1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt1 = msg.(*testspb.EmbeddedSt1)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = (*es).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (es EmbeddedSt1) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt1"
}

func IsEmbeddedSt1ReprEmpty(goor EmbeddedSt1) (empty bool) {
	{
		empty = true
		{
			e := IsPrimitivesStructReprEmpty(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
	}
	return
}

func (es EmbeddedSt2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt2
	{
		if IsEmbeddedSt2ReprEmpty(es) {
			var pbov *testspb.EmbeddedSt2
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt2)
		{
			pbom := proto.Message(nil)
			pbom, err = es.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.ArraysStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStruct = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.PointersStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStruct = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}

func (es EmbeddedSt2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt2)
	msg = pbo
	return
}

func (es *EmbeddedSt2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt2 = msg.(*testspb.EmbeddedSt2)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = (*es).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					err = (*es).ArraysStruct.FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = (*es).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					err = (*es).PointersStruct.FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (es EmbeddedSt2) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt2"
}

func IsEmbeddedSt2ReprEmpty(goor EmbeddedSt2) (empty bool) {
	{
		empty = true
		{
			e := IsPrimitivesStructReprEmpty(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
		{
			e := IsArraysStructReprEmpty(goor.ArraysStruct)
			if e == false {
				return false
			}
		}
		{
			e := IsSlicesStructReprEmpty(goor.SlicesStruct)
			if e == false {
				return false
			}
		}
		{
			e := IsPointersStructReprEmpty(goor.PointersStruct)
			if e == false {
				return false
			}
		}
	}
	return
}

func (es EmbeddedSt3) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt3
	{
		if IsEmbeddedSt3ReprEmpty(es) {
			var pbov *testspb.EmbeddedSt3
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt3)
		{
			if es.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.PrimitivesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
				if pbo.PrimitivesStruct == nil {
					pbo.PrimitivesStruct = new(testspb.PrimitivesStruct)
				}
			}
		}
		{
			if es.ArraysStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.ArraysStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ArraysStruct = pbom.(*testspb.ArraysStruct)
				if pbo.ArraysStruct == nil {
					pbo.ArraysStruct = new(testspb.ArraysStruct)
				}
			}
		}
		{
			if es.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.SlicesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
				if pbo.SlicesStruct == nil {
					pbo.SlicesStruct = new(testspb.SlicesStruct)
				}
			}
		}
		{
			if es.PointersStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.PointersStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PointersStruct = pbom.(*testspb.PointersStruct)
				if pbo.PointersStruct == nil {
					pbo.PointersStruct = new(testspb.PointersStruct)
				}
			}
		}
		{
			if es.EmptyStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.EmptyStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.EmptyStruct = pbom.(*testspb.EmptyStruct)
				if pbo.EmptyStruct == nil {
					pbo.EmptyStruct = new(testspb.EmptyStruct)
				}
			}
		}
	}
	msg = pbo
	return
}

func (es EmbeddedSt3) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt3)
	msg = pbo
	return
}

func (es *EmbeddedSt3) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt3 = msg.(*testspb.EmbeddedSt3)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					(*es).PrimitivesStruct = new(PrimitivesStruct)
					err = (*es).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					(*es).ArraysStruct = new(ArraysStruct)
					err = (*es).ArraysStruct.FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					(*es).SlicesStruct = new(SlicesStruct)
					err = (*es).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					(*es).PointersStruct = new(PointersStruct)
					err = (*es).PointersStruct.FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.EmptyStruct != nil {
					(*es).EmptyStruct = new(EmptyStruct)
					err = (*es).EmptyStruct.FromPBMessage(cdc, pbo.EmptyStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (es EmbeddedSt3) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt3"
}

func IsEmbeddedSt3ReprEmpty(goor EmbeddedSt3) (empty bool) {
	{
		empty = true
		{
			if goor.PrimitivesStruct != nil {
				return false
			}
		}
		{
			if goor.ArraysStruct != nil {
				return false
			}
		}
		{
			if goor.SlicesStruct != nil {
				return false
			}
		}
		{
			if goor.PointersStruct != nil {
				return false
			}
		}
		{
			if goor.EmptyStruct != nil {
				return false
			}
		}
	}
	return
}

func (es EmbeddedSt4) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt4
	{
		if IsEmbeddedSt4ReprEmpty(es) {
			var pbov *testspb.EmbeddedSt4
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt4)
		{
			pbo.Foo1 = int64(es.Foo1)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbo.Foo2 = string(es.Foo2)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.ArraysStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStructField = pbom.(*testspb.ArraysStruct)
		}
		{
			goorl := len(es.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := es.Foo3[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Foo3 = pbos
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbo.Foo4 = bool(es.Foo4)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = es.PointersStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStructField = pbom.(*testspb.PointersStruct)
		}
		{
			pbo.Foo5 = uint64(es.Foo5)
		}
	}
	msg = pbo
	return
}

func (es EmbeddedSt4) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt4)
	msg = pbo
	return
}

func (es *EmbeddedSt4) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt4 = msg.(*testspb.EmbeddedSt4)
	{
		if pbo != nil {
			{
				(*es).Foo1 = int(int(pbo.Foo1))
			}
			{
				if pbo.PrimitivesStruct != nil {
					err = (*es).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo2 = string(pbo.Foo2)
			}
			{
				if pbo.ArraysStructField != nil {
					err = (*es).ArraysStructField.FromPBMessage(cdc, pbo.ArraysStructField)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Foo3 != nil {
					pbol = len(pbo.Foo3)
				}
				if pbol == 0 {
					(*es).Foo3 = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*es).Foo3 = goors
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = (*es).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo4 = bool(pbo.Foo4)
			}
			{
				if pbo.PointersStructField != nil {
					err = (*es).PointersStructField.FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo5 = uint(uint(pbo.Foo5))
			}
		}
	}
	return
}

func (es EmbeddedSt4) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt4"
}

func IsEmbeddedSt4ReprEmpty(goor EmbeddedSt4) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != int(0) {
				return false
			}
		}
		{
			e := IsPrimitivesStructReprEmpty(goor.PrimitivesStruct)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo2 != string("") {
				return false
			}
		}
		{
			e := IsArraysStructReprEmpty(goor.ArraysStructField)
			if e == false {
				return false
			}
		}
		{
			if len(goor.Foo3) != 0 {
				return false
			}
		}
		{
			e := IsSlicesStructReprEmpty(goor.SlicesStruct)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo4 != bool(false) {
				return false
			}
		}
		{
			e := IsPointersStructReprEmpty(goor.PointersStructField)
			if e == false {
				return false
			}
		}
		{
			if goor.Foo5 != uint(0) {
				return false
			}
		}
	}
	return
}

func (es EmbeddedSt5) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt5NameOverride
	{
		if IsEmbeddedSt5NameOverrideReprEmpty(es) {
			var pbov *testspb.EmbeddedSt5NameOverride
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt5NameOverride)
		{
			pbo.Foo1 = int64(es.Foo1)
		}
		{
			if es.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.PrimitivesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
				if pbo.PrimitivesStruct == nil {
					pbo.PrimitivesStruct = new(testspb.PrimitivesStruct)
				}
			}
		}
		{
			pbo.Foo2 = string(es.Foo2)
		}
		{
			if es.ArraysStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = es.ArraysStructField.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.ArraysStructField = pbom.(*testspb.ArraysStruct)
				if pbo.ArraysStructField == nil {
					pbo.ArraysStructField = new(testspb.ArraysStruct)
				}
			}
		}
		{
			goorl := len(es.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := es.Foo3[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Foo3 = pbos
			}
		}
		{
			if es.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = es.SlicesStruct.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
				if pbo.SlicesStruct == nil {
					pbo.SlicesStruct = new(testspb.SlicesStruct)
				}
			}
		}
		{
			pbo.Foo4 = bool(es.Foo4)
		}
		{
			if es.PointersStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = es.PointersStructField.ToPBMessage(cdc)
				if err != nil {
					return
				}
				pbo.PointersStructField = pbom.(*testspb.PointersStruct)
				if pbo.PointersStructField == nil {
					pbo.PointersStructField = new(testspb.PointersStruct)
				}
			}
		}
		{
			pbo.Foo5 = uint64(es.Foo5)
		}
	}
	msg = pbo
	return
}

func (es EmbeddedSt5) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt5NameOverride)
	msg = pbo
	return
}

func (es *EmbeddedSt5) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt5NameOverride = msg.(*testspb.EmbeddedSt5NameOverride)
	{
		if pbo != nil {
			{
				(*es).Foo1 = int(int(pbo.Foo1))
			}
			{
				if pbo.PrimitivesStruct != nil {
					(*es).PrimitivesStruct = new(PrimitivesStruct)
					err = (*es).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo2 = string(pbo.Foo2)
			}
			{
				if pbo.ArraysStructField != nil {
					(*es).ArraysStructField = new(ArraysStruct)
					err = (*es).ArraysStructField.FromPBMessage(cdc, pbo.ArraysStructField)
					if err != nil {
						return
					}
				}
			}
			{
				var pbol int = 0
				if pbo.Foo3 != nil {
					pbol = len(pbo.Foo3)
				}
				if pbol == 0 {
					(*es).Foo3 = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*es).Foo3 = goors
				}
			}
			{
				if pbo.SlicesStruct != nil {
					(*es).SlicesStruct = new(SlicesStruct)
					err = (*es).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo4 = bool(pbo.Foo4)
			}
			{
				if pbo.PointersStructField != nil {
					(*es).PointersStructField = new(PointersStruct)
					err = (*es).PointersStructField.FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				(*es).Foo5 = uint(uint(pbo.Foo5))
			}
		}
	}
	return
}

func (es EmbeddedSt5) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt5NameOverride"
}

func IsEmbeddedSt5NameOverrideReprEmpty(goor EmbeddedSt5) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != int(0) {
				return false
			}
		}
		{
			if goor.PrimitivesStruct != nil {
				return false
			}
		}
		{
			if goor.Foo2 != string("") {
				return false
			}
		}
		{
			if goor.ArraysStructField != nil {
				return false
			}
		}
		{
			if len(goor.Foo3) != 0 {
				return false
			}
		}
		{
			if goor.SlicesStruct != nil {
				return false
			}
		}
		{
			if goor.Foo4 != bool(false) {
				return false
			}
		}
		{
			if goor.PointersStructField != nil {
				return false
			}
		}
		{
			if goor.Foo5 != uint(0) {
				return false
			}
		}
	}
	return
}

func (ams AminoMarshalerStruct1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct1
	{
		goor, err1 := ams.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerStruct1ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerStruct1
			msg = pbov
			return
		}
		pbo = new(testspb.AminoMarshalerStruct1)
		{
			pbo.C = int64(goor.C)
		}
		{
			pbo.D = int64(goor.D)
		}
	}
	msg = pbo
	return
}

func (ams AminoMarshalerStruct1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct1)
	msg = pbo
	return
}

func (ams *AminoMarshalerStruct1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct1 = msg.(*testspb.AminoMarshalerStruct1)
	{
		if pbo != nil {
			var goor ReprStruct1
			{
				goor.C = int64(pbo.C)
			}
			{
				goor.D = int64(pbo.D)
			}
			err = ams.UnmarshalAmino(goor)
			if err != nil {
				return
			}
		}
	}
	return
}

func (ams AminoMarshalerStruct1) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct1"
}

func IsAminoMarshalerStruct1ReprEmpty(goor ReprStruct1) (empty bool) {
	{
		empty = true
		{
			if goor.C != int64(0) {
				return false
			}
		}
		{
			if goor.D != int64(0) {
				return false
			}
		}
	}
	return
}

func (rs ReprStruct1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprStruct1
	{
		if IsReprStruct1ReprEmpty(rs) {
			var pbov *testspb.ReprStruct1
			msg = pbov
			return
		}
		pbo = new(testspb.ReprStruct1)
		{
			pbo.C = int64(rs.C)
		}
		{
			pbo.D = int64(rs.D)
		}
	}
	msg = pbo
	return
}

func (rs ReprStruct1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprStruct1)
	msg = pbo
	return
}

func (rs *ReprStruct1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprStruct1 = msg.(*testspb.ReprStruct1)
	{
		if pbo != nil {
			{
				(*rs).C = int64(pbo.C)
			}
			{
				(*rs).D = int64(pbo.D)
			}
		}
	}
	return
}

func (rs ReprStruct1) GetTypeURL() (typeURL string) {
	return "/tests.ReprStruct1"
}

func IsReprStruct1ReprEmpty(goor ReprStruct1) (empty bool) {
	{
		empty = true
		{
			if goor.C != int64(0) {
				return false
			}
		}
		{
			if goor.D != int64(0) {
				return false
			}
		}
	}
	return
}

func (ams AminoMarshalerStruct2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct2
	{
		goor, err1 := ams.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerStruct2ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerStruct2
			msg = pbov
			return
		}
		goorl := len(goor)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]*testspb.ReprElem2, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goor[i]
					{
						pbom := proto.Message(nil)
						pbom, err = goore.ToPBMessage(cdc)
						if err != nil {
							return
						}
						pbos[i] = pbom.(*testspb.ReprElem2)
					}
				}
			}
			pbo = &testspb.AminoMarshalerStruct2{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ams AminoMarshalerStruct2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct2)
	msg = pbo
	return
}

func (ams *AminoMarshalerStruct2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct2 = msg.(*testspb.AminoMarshalerStruct2)
	{
		var goor []ReprElem2
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			goor = nil
		} else {
			goors := make([]ReprElem2, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
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
			goor = goors
		}
		err = ams.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (ams AminoMarshalerStruct2) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct2"
}

func IsAminoMarshalerStruct2ReprEmpty(goor []ReprElem2) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (re ReprElem2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprElem2
	{
		if IsReprElem2ReprEmpty(re) {
			var pbov *testspb.ReprElem2
			msg = pbov
			return
		}
		pbo = new(testspb.ReprElem2)
		{
			pbo.Key = string(re.Key)
		}
		{
			if re.Value != nil {
				typeUrl := cdc.GetTypeURL(re.Value)
				bz := []byte(nil)
				bz, err = cdc.Marshal(re.Value)
				if err != nil {
					return
				}
				pbo.Value = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
	}
	msg = pbo
	return
}

func (re ReprElem2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprElem2)
	msg = pbo
	return
}

func (re *ReprElem2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprElem2 = msg.(*testspb.ReprElem2)
	{
		if pbo != nil {
			{
				(*re).Key = string(pbo.Key)
			}
			{
				typeUrl := pbo.Value.TypeUrl
				bz := pbo.Value.Value
				goorp := &(*re).Value
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (re ReprElem2) GetTypeURL() (typeURL string) {
	return "/tests.ReprElem2"
}

func IsReprElem2ReprEmpty(goor ReprElem2) (empty bool) {
	{
		empty = true
		{
			if goor.Key != string("") {
				return false
			}
		}
		{
			if goor.Value != nil {
				return false
			}
		}
	}
	return
}

func (ams AminoMarshalerStruct3) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct3
	{
		goor, err1 := ams.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerStruct3ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerStruct3
			msg = pbov
			return
		}
		pbo = &testspb.AminoMarshalerStruct3{Value: int32(goor)}
	}
	msg = pbo
	return
}

func (ams AminoMarshalerStruct3) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct3)
	msg = pbo
	return
}

func (ams *AminoMarshalerStruct3) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct3 = msg.(*testspb.AminoMarshalerStruct3)
	{
		var goor int32
		goor = int32(pbo.Value)
		err = ams.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (ams AminoMarshalerStruct3) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct3"
}

func IsAminoMarshalerStruct3ReprEmpty(goor int32) (empty bool) {
	{
		empty = true
		if goor != int32(0) {
			return false
		}
	}
	return
}

func (am AminoMarshalerInt4) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerInt4
	{
		goor, err1 := am.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerInt4ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerInt4
			msg = pbov
			return
		}
		pbo = new(testspb.AminoMarshalerInt4)
		{
			pbo.A = int32(goor.A)
		}
	}
	msg = pbo
	return
}

func (am AminoMarshalerInt4) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerInt4)
	msg = pbo
	return
}

func (am *AminoMarshalerInt4) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerInt4 = msg.(*testspb.AminoMarshalerInt4)
	{
		if pbo != nil {
			var goor ReprStruct4
			{
				goor.A = int32(pbo.A)
			}
			err = am.UnmarshalAmino(goor)
			if err != nil {
				return
			}
		}
	}
	return
}

func (am AminoMarshalerInt4) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerInt4"
}

func IsAminoMarshalerInt4ReprEmpty(goor ReprStruct4) (empty bool) {
	{
		empty = true
		{
			if goor.A != int32(0) {
				return false
			}
		}
	}
	return
}

func (am AminoMarshalerInt5) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerInt5
	{
		goor, err1 := am.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerInt5ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerInt5
			msg = pbov
			return
		}
		pbo = &testspb.AminoMarshalerInt5{Value: string(goor)}
	}
	msg = pbo
	return
}

func (am AminoMarshalerInt5) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerInt5)
	msg = pbo
	return
}

func (am *AminoMarshalerInt5) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerInt5 = msg.(*testspb.AminoMarshalerInt5)
	{
		var goor string
		goor = string(pbo.Value)
		err = am.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (am AminoMarshalerInt5) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerInt5"
}

func IsAminoMarshalerInt5ReprEmpty(goor string) (empty bool) {
	{
		empty = true
		if goor != string("") {
			return false
		}
	}
	return
}

func (ams AminoMarshalerStruct6) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct6
	{
		goor, err1 := ams.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerStruct6ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerStruct6
			msg = pbov
			return
		}
		goorl := len(goor)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]*testspb.AminoMarshalerStruct1, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goor[i]
					{
						pbom := proto.Message(nil)
						pbom, err = goore.ToPBMessage(cdc)
						if err != nil {
							return
						}
						pbos[i] = pbom.(*testspb.AminoMarshalerStruct1)
					}
				}
			}
			pbo = &testspb.AminoMarshalerStruct6{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ams AminoMarshalerStruct6) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct6)
	msg = pbo
	return
}

func (ams *AminoMarshalerStruct6) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct6 = msg.(*testspb.AminoMarshalerStruct6)
	{
		var goor []AminoMarshalerStruct1
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			goor = nil
		} else {
			goors := make([]AminoMarshalerStruct1, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
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
			goor = goors
		}
		err = ams.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (ams AminoMarshalerStruct6) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct6"
}

func IsAminoMarshalerStruct6ReprEmpty(goor []AminoMarshalerStruct1) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (ams AminoMarshalerStruct7) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct7
	{
		goor, err1 := ams.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsAminoMarshalerStruct7ReprEmpty(goor) {
			var pbov *testspb.AminoMarshalerStruct7
			msg = pbov
			return
		}
		goorl := len(goor)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goor[i]
					{
						goor1, err2 := goore.MarshalAmino()
						if err2 != nil {
							return nil, err2
						}
						if !IsReprElem7ReprEmpty(goor1) {
							pbos[i] = byte(goor1)
						}
					}
				}
			}
			pbo = &testspb.AminoMarshalerStruct7{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ams AminoMarshalerStruct7) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct7)
	msg = pbo
	return
}

func (ams *AminoMarshalerStruct7) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct7 = msg.(*testspb.AminoMarshalerStruct7)
	{
		var goor []ReprElem7
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			goor = nil
		} else {
			goors := make([]ReprElem7, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
					{
						pboev := pboe
						var goor1 uint8
						goor1 = uint8(uint8(pboev))
						err = goors[i].UnmarshalAmino(goor1)
						if err != nil {
							return
						}
					}
				}
			}
			goor = goors
		}
		err = ams.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (ams AminoMarshalerStruct7) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct7"
}

func IsAminoMarshalerStruct7ReprEmpty(goor []ReprElem7) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (re ReprElem7) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprElem7
	{
		goor, err1 := re.MarshalAmino()
		if err1 != nil {
			return nil, err1
		}
		if IsReprElem7ReprEmpty(goor) {
			var pbov *testspb.ReprElem7
			msg = pbov
			return
		}
		pbo = &testspb.ReprElem7{Value: uint32(goor)}
	}
	msg = pbo
	return
}

func (re ReprElem7) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprElem7)
	msg = pbo
	return
}

func (re *ReprElem7) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprElem7 = msg.(*testspb.ReprElem7)
	{
		var goor uint8
		goor = uint8(uint8(pbo.Value))
		err = re.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}

func (re ReprElem7) GetTypeURL() (typeURL string) {
	return "/tests.ReprElem7"
}

func IsReprElem7ReprEmpty(goor uint8) (empty bool) {
	{
		empty = true
		if goor != uint8(0) {
			return false
		}
	}
	return
}

func (id IntDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntDef
	{
		if IsIntDefReprEmpty(id) {
			var pbov *testspb.IntDef
			msg = pbov
			return
		}
		pbo = &testspb.IntDef{Value: int64(id)}
	}
	msg = pbo
	return
}

func (id IntDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntDef)
	msg = pbo
	return
}

func (id *IntDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntDef = msg.(*testspb.IntDef)
	{
		*id = IntDef(int(pbo.Value))
	}
	return
}

func (id IntDef) GetTypeURL() (typeURL string) {
	return "/tests.IntDef"
}

func IsIntDefReprEmpty(goor IntDef) (empty bool) {
	{
		empty = true
		if goor != IntDef(0) {
			return false
		}
	}
	return
}

func (ia IntAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntAr
	{
		if IsIntArReprEmpty(ia) {
			var pbov *testspb.IntAr
			msg = pbov
			return
		}
		goorl := len(ia)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]int64, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := ia[i]
					{
						pbos[i] = int64(goore)
					}
				}
			}
			pbo = &testspb.IntAr{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ia IntAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntAr)
	msg = pbo
	return
}

func (ia *IntAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntAr = msg.(*testspb.IntAr)
	{
		goors := [4]int{}
		for i := 0; i < 4; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = int(int(pboev))
				}
			}
		}
		*ia = goors
	}
	return
}

func (ia IntAr) GetTypeURL() (typeURL string) {
	return "/tests.IntAr"
}

func IsIntArReprEmpty(goor IntAr) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (is IntSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntSl
	{
		if IsIntSlReprEmpty(is) {
			var pbov *testspb.IntSl
			msg = pbov
			return
		}
		goorl := len(is)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]int64, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := is[i]
					{
						pbos[i] = int64(goore)
					}
				}
			}
			pbo = &testspb.IntSl{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (is IntSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntSl)
	msg = pbo
	return
}

func (is *IntSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntSl = msg.(*testspb.IntSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*is = nil
		} else {
			goors := make([]int, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
					{
						pboev := pboe
						goors[i] = int(int(pboev))
					}
				}
			}
			*is = goors
		}
	}
	return
}

func (is IntSl) GetTypeURL() (typeURL string) {
	return "/tests.IntSl"
}

func IsIntSlReprEmpty(goor IntSl) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (ba ByteAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ByteAr
	{
		if IsByteArReprEmpty(ba) {
			var pbov *testspb.ByteAr
			msg = pbov
			return
		}
		goorl := len(ba)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := ba[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &testspb.ByteAr{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ba ByteAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ByteAr)
	msg = pbo
	return
}

func (ba *ByteAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ByteAr = msg.(*testspb.ByteAr)
	{
		goors := [4]uint8{}
		for i := 0; i < 4; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = uint8(uint8(pboev))
				}
			}
		}
		*ba = goors
	}
	return
}

func (ba ByteAr) GetTypeURL() (typeURL string) {
	return "/tests.ByteAr"
}

func IsByteArReprEmpty(goor ByteAr) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (bs ByteSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ByteSl
	{
		if IsByteSlReprEmpty(bs) {
			var pbov *testspb.ByteSl
			msg = pbov
			return
		}
		goorl := len(bs)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := bs[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &testspb.ByteSl{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (bs ByteSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ByteSl)
	msg = pbo
	return
}

func (bs *ByteSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ByteSl = msg.(*testspb.ByteSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*bs = nil
		} else {
			goors := make([]uint8, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
					{
						pboev := pboe
						goors[i] = uint8(uint8(pboev))
					}
				}
			}
			*bs = goors
		}
	}
	return
}

func (bs ByteSl) GetTypeURL() (typeURL string) {
	return "/tests.ByteSl"
}

func IsByteSlReprEmpty(goor ByteSl) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (psd PrimitivesStructDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructDef
	{
		if IsPrimitivesStructDefReprEmpty(psd) {
			var pbov *testspb.PrimitivesStructDef
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStructDef)
		{
			pbo.Int8 = int32(psd.Int8)
		}
		{
			pbo.Int16 = int32(psd.Int16)
		}
		{
			pbo.Int32 = int32(psd.Int32)
		}
		{
			pbo.Int32Fixed = int32(psd.Int32Fixed)
		}
		{
			pbo.Int64 = int64(psd.Int64)
		}
		{
			pbo.Int64Fixed = int64(psd.Int64Fixed)
		}
		{
			pbo.Int = int64(psd.Int)
		}
		{
			pbo.Byte = uint32(psd.Byte)
		}
		{
			pbo.Uint8 = uint32(psd.Uint8)
		}
		{
			pbo.Uint16 = uint32(psd.Uint16)
		}
		{
			pbo.Uint32 = uint32(psd.Uint32)
		}
		{
			pbo.Uint32Fixed = uint32(psd.Uint32Fixed)
		}
		{
			pbo.Uint64 = uint64(psd.Uint64)
		}
		{
			pbo.Uint64Fixed = uint64(psd.Uint64Fixed)
		}
		{
			pbo.Uint = uint64(psd.Uint)
		}
		{
			pbo.Str = string(psd.Str)
		}
		{
			goorl := len(psd.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := psd.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(psd.Time) {
				pbo.Time = timestamppb.New(psd.Time)
			}
		}
		{
			if psd.Duration.Nanoseconds() != 0 {
				pbo.Duration = durationpb.New(psd.Duration)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = psd.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}

func (psd PrimitivesStructDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructDef)
	msg = pbo
	return
}

func (psd *PrimitivesStructDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructDef = msg.(*testspb.PrimitivesStructDef)
	{
		if pbo != nil {
			{
				(*psd).Int8 = int8(int8(pbo.Int8))
			}
			{
				(*psd).Int16 = int16(int16(pbo.Int16))
			}
			{
				(*psd).Int32 = int32(pbo.Int32)
			}
			{
				(*psd).Int32Fixed = int32(pbo.Int32Fixed)
			}
			{
				(*psd).Int64 = int64(pbo.Int64)
			}
			{
				(*psd).Int64Fixed = int64(pbo.Int64Fixed)
			}
			{
				(*psd).Int = int(int(pbo.Int))
			}
			{
				(*psd).Byte = uint8(uint8(pbo.Byte))
			}
			{
				(*psd).Uint8 = uint8(uint8(pbo.Uint8))
			}
			{
				(*psd).Uint16 = uint16(uint16(pbo.Uint16))
			}
			{
				(*psd).Uint32 = uint32(pbo.Uint32)
			}
			{
				(*psd).Uint32Fixed = uint32(pbo.Uint32Fixed)
			}
			{
				(*psd).Uint64 = uint64(pbo.Uint64)
			}
			{
				(*psd).Uint64Fixed = uint64(pbo.Uint64Fixed)
			}
			{
				(*psd).Uint = uint(uint(pbo.Uint))
			}
			{
				(*psd).Str = string(pbo.Str)
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					(*psd).Bytes = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*psd).Bytes = goors
				}
			}
			{
				(*psd).Time = pbo.Time.AsTime()
			}
			{
				(*psd).Duration = pbo.Duration.AsDuration()
			}
			{
				if pbo.Empty != nil {
					err = (*psd).Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (psd PrimitivesStructDef) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStructDef"
}

func IsPrimitivesStructDefReprEmpty(goor PrimitivesStructDef) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != int8(0) {
				return false
			}
		}
		{
			if goor.Int16 != int16(0) {
				return false
			}
		}
		{
			if goor.Int32 != int32(0) {
				return false
			}
		}
		{
			if goor.Int32Fixed != int32(0) {
				return false
			}
		}
		{
			if goor.Int64 != int64(0) {
				return false
			}
		}
		{
			if goor.Int64Fixed != int64(0) {
				return false
			}
		}
		{
			if goor.Int != int(0) {
				return false
			}
		}
		{
			if goor.Byte != uint8(0) {
				return false
			}
		}
		{
			if goor.Uint8 != uint8(0) {
				return false
			}
		}
		{
			if goor.Uint16 != uint16(0) {
				return false
			}
		}
		{
			if goor.Uint32 != uint32(0) {
				return false
			}
		}
		{
			if goor.Uint32Fixed != uint32(0) {
				return false
			}
		}
		{
			if goor.Uint64 != uint64(0) {
				return false
			}
		}
		{
			if goor.Uint64Fixed != uint64(0) {
				return false
			}
		}
		{
			if goor.Uint != uint(0) {
				return false
			}
		}
		{
			if goor.Str != string("") {
				return false
			}
		}
		{
			if len(goor.Bytes) != 0 {
				return false
			}
		}
		{
			if !amino.IsEmptyTime(goor.Time) {
				return false
			}
		}
		{
			if goor.Duration != time.Duration(0) {
				return false
			}
		}
		{
			e := IsEmptyStructReprEmpty(goor.Empty)
			if e == false {
				return false
			}
		}
	}
	return
}

func (ps PrimitivesStructSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructSl
	{
		if IsPrimitivesStructSlReprEmpty(ps) {
			var pbov *testspb.PrimitivesStructSl
			msg = pbov
			return
		}
		goorl := len(ps)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]*testspb.PrimitivesStruct, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := ps[i]
					{
						pbom := proto.Message(nil)
						pbom, err = goore.ToPBMessage(cdc)
						if err != nil {
							return
						}
						pbos[i] = pbom.(*testspb.PrimitivesStruct)
					}
				}
			}
			pbo = &testspb.PrimitivesStructSl{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ps PrimitivesStructSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructSl)
	msg = pbo
	return
}

func (ps *PrimitivesStructSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructSl = msg.(*testspb.PrimitivesStructSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*ps = nil
		} else {
			goors := make([]PrimitivesStruct, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
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
			*ps = goors
		}
	}
	return
}

func (ps PrimitivesStructSl) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStructSl"
}

func IsPrimitivesStructSlReprEmpty(goor PrimitivesStructSl) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (ps PrimitivesStructAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructAr
	{
		if IsPrimitivesStructArReprEmpty(ps) {
			var pbov *testspb.PrimitivesStructAr
			msg = pbov
			return
		}
		goorl := len(ps)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]*testspb.PrimitivesStruct, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := ps[i]
					{
						pbom := proto.Message(nil)
						pbom, err = goore.ToPBMessage(cdc)
						if err != nil {
							return
						}
						pbos[i] = pbom.(*testspb.PrimitivesStruct)
					}
				}
			}
			pbo = &testspb.PrimitivesStructAr{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ps PrimitivesStructAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructAr)
	msg = pbo
	return
}

func (ps *PrimitivesStructAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructAr = msg.(*testspb.PrimitivesStructAr)
	{
		goors := [2]PrimitivesStruct{}
		for i := 0; i < 2; i += 1 {
			{
				pboe := pbo.Value[i]
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
		*ps = goors
	}
	return
}

func (ps PrimitivesStructAr) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStructAr"
}

func IsPrimitivesStructArReprEmpty(goor PrimitivesStructAr) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (c Concrete1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete1
	{
		if IsConcrete1ReprEmpty(c) {
			var pbov *testspb.Concrete1
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete1)
	}
	msg = pbo
	return
}

func (c Concrete1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.Concrete1)
	msg = pbo
	return
}

func (c *Concrete1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete1 = msg.(*testspb.Concrete1)
	{
		if pbo != nil {
		}
	}
	return
}

func (c Concrete1) GetTypeURL() (typeURL string) {
	return "/tests.Concrete1"
}

func IsConcrete1ReprEmpty(goor Concrete1) (empty bool) {
	{
		empty = true
	}
	return
}

func (c Concrete2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete2
	{
		if IsConcrete2ReprEmpty(c) {
			var pbov *testspb.Concrete2
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete2)
	}
	msg = pbo
	return
}

func (c Concrete2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.Concrete2)
	msg = pbo
	return
}

func (c *Concrete2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete2 = msg.(*testspb.Concrete2)
	{
		if pbo != nil {
		}
	}
	return
}

func (c Concrete2) GetTypeURL() (typeURL string) {
	return "/tests.Concrete2"
}

func IsConcrete2ReprEmpty(goor Concrete2) (empty bool) {
	{
		empty = true
	}
	return
}

func (ctd ConcreteTypeDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ConcreteTypeDef
	{
		if IsConcreteTypeDefReprEmpty(ctd) {
			var pbov *testspb.ConcreteTypeDef
			msg = pbov
			return
		}
		goorl := len(ctd)
		if goorl == 0 {
			pbo = nil
		} else {
			pbos := make([]uint8, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := ctd[i]
					{
						pbos[i] = byte(goore)
					}
				}
			}
			pbo = &testspb.ConcreteTypeDef{Value: pbos}
		}
	}
	msg = pbo
	return
}

func (ctd ConcreteTypeDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ConcreteTypeDef)
	msg = pbo
	return
}

func (ctd *ConcreteTypeDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ConcreteTypeDef = msg.(*testspb.ConcreteTypeDef)
	{
		goors := [4]uint8{}
		for i := 0; i < 4; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = uint8(uint8(pboev))
				}
			}
		}
		*ctd = goors
	}
	return
}

func (ctd ConcreteTypeDef) GetTypeURL() (typeURL string) {
	return "/tests.ConcreteTypeDef"
}

func IsConcreteTypeDefReprEmpty(goor ConcreteTypeDef) (empty bool) {
	{
		empty = true
		if len(goor) != 0 {
			return false
		}
	}
	return
}

func (cwb ConcreteWrappedBytes) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ConcreteWrappedBytes
	{
		if IsConcreteWrappedBytesReprEmpty(cwb) {
			var pbov *testspb.ConcreteWrappedBytes
			msg = pbov
			return
		}
		pbo = new(testspb.ConcreteWrappedBytes)
		{
			goorl := len(cwb.Value)
			if goorl == 0 {
				pbo.Value = nil
			} else {
				pbos := make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := cwb.Value[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Value = pbos
			}
		}
	}
	msg = pbo
	return
}

func (cwb ConcreteWrappedBytes) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ConcreteWrappedBytes)
	msg = pbo
	return
}

func (cwb *ConcreteWrappedBytes) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ConcreteWrappedBytes = msg.(*testspb.ConcreteWrappedBytes)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Value != nil {
					pbol = len(pbo.Value)
				}
				if pbol == 0 {
					(*cwb).Value = nil
				} else {
					goors := make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Value[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*cwb).Value = goors
				}
			}
		}
	}
	return
}

func (cwb ConcreteWrappedBytes) GetTypeURL() (typeURL string) {
	return "/tests.ConcreteWrappedBytes"
}

func IsConcreteWrappedBytesReprEmpty(goor ConcreteWrappedBytes) (empty bool) {
	{
		empty = true
		{
			if len(goor.Value) != 0 {
				return false
			}
		}
	}
	return
}

func (ifs InterfaceFieldsStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.InterfaceFieldsStruct
	{
		if IsInterfaceFieldsStructReprEmpty(ifs) {
			var pbov *testspb.InterfaceFieldsStruct
			msg = pbov
			return
		}
		pbo = new(testspb.InterfaceFieldsStruct)
		{
			if ifs.F1 != nil {
				typeUrl := cdc.GetTypeURL(ifs.F1)
				bz := []byte(nil)
				bz, err = cdc.Marshal(ifs.F1)
				if err != nil {
					return
				}
				pbo.F1 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if ifs.F2 != nil {
				typeUrl := cdc.GetTypeURL(ifs.F2)
				bz := []byte(nil)
				bz, err = cdc.Marshal(ifs.F2)
				if err != nil {
					return
				}
				pbo.F2 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if ifs.F3 != nil {
				typeUrl := cdc.GetTypeURL(ifs.F3)
				bz := []byte(nil)
				bz, err = cdc.Marshal(ifs.F3)
				if err != nil {
					return
				}
				pbo.F3 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if ifs.F4 != nil {
				typeUrl := cdc.GetTypeURL(ifs.F4)
				bz := []byte(nil)
				bz, err = cdc.Marshal(ifs.F4)
				if err != nil {
					return
				}
				pbo.F4 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
	}
	msg = pbo
	return
}

func (ifs InterfaceFieldsStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.InterfaceFieldsStruct)
	msg = pbo
	return
}

func (ifs *InterfaceFieldsStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.InterfaceFieldsStruct = msg.(*testspb.InterfaceFieldsStruct)
	{
		if pbo != nil {
			{
				typeUrl := pbo.F1.TypeUrl
				bz := pbo.F1.Value
				goorp := &(*ifs).F1
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F2.TypeUrl
				bz := pbo.F2.Value
				goorp := &(*ifs).F2
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F3.TypeUrl
				bz := pbo.F3.Value
				goorp := &(*ifs).F3
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F4.TypeUrl
				bz := pbo.F4.Value
				goorp := &(*ifs).F4
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (ifs InterfaceFieldsStruct) GetTypeURL() (typeURL string) {
	return "/tests.InterfaceFieldsStruct"
}

func IsInterfaceFieldsStructReprEmpty(goor InterfaceFieldsStruct) (empty bool) {
	{
		empty = true
		{
			if goor.F1 != nil {
				return false
			}
		}
		{
			if goor.F2 != nil {
				return false
			}
		}
		{
			if goor.F3 != nil {
				return false
			}
		}
		{
			if goor.F4 != nil {
				return false
			}
		}
	}
	return
}
