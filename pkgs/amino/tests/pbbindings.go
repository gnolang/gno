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

func (goo EmptyStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmptyStruct
	{
		if IsEmptyStructReprEmpty(goo) {
			var pbov *testspb.EmptyStruct
			msg = pbov
			return
		}
		pbo = new(testspb.EmptyStruct)
	}
	msg = pbo
	return
}
func (goo EmptyStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmptyStruct)
	msg = pbo
	return
}
func (goo *EmptyStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmptyStruct = msg.(*testspb.EmptyStruct)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ EmptyStruct) GetTypeURL() (typeURL string) {
	return "/tests.EmptyStruct"
}
func IsEmptyStructReprEmpty(goor EmptyStruct) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo PrimitivesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStruct
	{
		if IsPrimitivesStructReprEmpty(goo) {
			var pbov *testspb.PrimitivesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStruct)
		{
			pbo.Int8 = int32(goo.Int8)
		}
		{
			pbo.Int16 = int32(goo.Int16)
		}
		{
			pbo.Int32 = int32(goo.Int32)
		}
		{
			pbo.Int32Fixed = int32(goo.Int32Fixed)
		}
		{
			pbo.Int64 = int64(goo.Int64)
		}
		{
			pbo.Int64Fixed = int64(goo.Int64Fixed)
		}
		{
			pbo.Int = int64(goo.Int)
		}
		{
			pbo.Byte = uint32(goo.Byte)
		}
		{
			pbo.Uint8 = uint32(goo.Uint8)
		}
		{
			pbo.Uint16 = uint32(goo.Uint16)
		}
		{
			pbo.Uint32 = uint32(goo.Uint32)
		}
		{
			pbo.Uint32Fixed = uint32(goo.Uint32Fixed)
		}
		{
			pbo.Uint64 = uint64(goo.Uint64)
		}
		{
			pbo.Uint64Fixed = uint64(goo.Uint64Fixed)
		}
		{
			pbo.Uint = uint64(goo.Uint)
		}
		{
			pbo.Str = string(goo.Str)
		}
		{
			goorl := len(goo.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			if goo.Duration.Nanoseconds() != 0 {
				pbo.Duration = durationpb.New(goo.Duration)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}
func (goo PrimitivesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStruct)
	msg = pbo
	return
}
func (goo *PrimitivesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStruct = msg.(*testspb.PrimitivesStruct)
	{
		if pbo != nil {
			{
				(*goo).Int8 = int8(int8(pbo.Int8))
			}
			{
				(*goo).Int16 = int16(int16(pbo.Int16))
			}
			{
				(*goo).Int32 = int32(pbo.Int32)
			}
			{
				(*goo).Int32Fixed = int32(pbo.Int32Fixed)
			}
			{
				(*goo).Int64 = int64(pbo.Int64)
			}
			{
				(*goo).Int64Fixed = int64(pbo.Int64Fixed)
			}
			{
				(*goo).Int = int(int(pbo.Int))
			}
			{
				(*goo).Byte = uint8(uint8(pbo.Byte))
			}
			{
				(*goo).Uint8 = uint8(uint8(pbo.Uint8))
			}
			{
				(*goo).Uint16 = uint16(uint16(pbo.Uint16))
			}
			{
				(*goo).Uint32 = uint32(pbo.Uint32)
			}
			{
				(*goo).Uint32Fixed = uint32(pbo.Uint32Fixed)
			}
			{
				(*goo).Uint64 = uint64(pbo.Uint64)
			}
			{
				(*goo).Uint64Fixed = uint64(pbo.Uint64Fixed)
			}
			{
				(*goo).Uint = uint(uint(pbo.Uint))
			}
			{
				(*goo).Str = string(pbo.Str)
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					(*goo).Bytes = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Bytes = goors
				}
			}
			{
				(*goo).Time = pbo.Time.AsTime()
			}
			{
				(*goo).Duration = pbo.Duration.AsDuration()
			}
			{
				if pbo.Empty != nil {
					err = (*goo).Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PrimitivesStruct) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStruct"
}
func IsPrimitivesStructReprEmpty(goor PrimitivesStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != 0 {
				return false
			}
		}
		{
			if goor.Int16 != 0 {
				return false
			}
		}
		{
			if goor.Int32 != 0 {
				return false
			}
		}
		{
			if goor.Int32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int64 != 0 {
				return false
			}
		}
		{
			if goor.Int64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int != 0 {
				return false
			}
		}
		{
			if goor.Byte != 0 {
				return false
			}
		}
		{
			if goor.Uint8 != 0 {
				return false
			}
		}
		{
			if goor.Uint16 != 0 {
				return false
			}
		}
		{
			if goor.Uint32 != 0 {
				return false
			}
		}
		{
			if goor.Uint32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint64 != 0 {
				return false
			}
		}
		{
			if goor.Uint64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint != 0 {
				return false
			}
		}
		{
			if goor.Str != "" {
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
			if goor.Duration != 0 {
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
func (goo ShortArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ShortArraysStruct
	{
		if IsShortArraysStructReprEmpty(goo) {
			var pbov *testspb.ShortArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ShortArraysStruct)
		{
			goorl := len(goo.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeAr[i]
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
			goorl := len(goo.DurationAr)
			if goorl == 0 {
				pbo.DurationAr = nil
			} else {
				var pbos = make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationAr[i]
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
func (goo ShortArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ShortArraysStruct)
	msg = pbo
	return
}
func (goo *ShortArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ShortArraysStruct = msg.(*testspb.ShortArraysStruct)
	{
		if pbo != nil {
			{
				var goors = [0]time.Time{}
				for i := 0; i < 0; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsTime()
						}
					}
				}
				(*goo).TimeAr = goors
			}
			{
				var goors = [0]time.Duration{}
				for i := 0; i < 0; i += 1 {
					{
						pboe := pbo.DurationAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsDuration()
						}
					}
				}
				(*goo).DurationAr = goors
			}
		}
	}
	return
}
func (_ ShortArraysStruct) GetTypeURL() (typeURL string) {
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
func (goo ArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysStruct
	{
		if IsArraysStructReprEmpty(goo) {
			var pbov *testspb.ArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysStruct)
		{
			goorl := len(goo.Int8Ar)
			if goorl == 0 {
				pbo.Int8Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Ar = pbos
			}
		}
		{
			goorl := len(goo.Int16Ar)
			if goorl == 0 {
				pbo.Int16Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Ar = pbos
			}
		}
		{
			goorl := len(goo.Int32Ar)
			if goorl == 0 {
				pbo.Int32Ar = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32Ar[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32Ar = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedAr)
			if goorl == 0 {
				pbo.Int32FixedAr = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedAr[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32FixedAr = pbos
			}
		}
		{
			goorl := len(goo.Int64Ar)
			if goorl == 0 {
				pbo.Int64Ar = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64Ar[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64Ar = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedAr)
			if goorl == 0 {
				pbo.Int64FixedAr = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedAr[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64FixedAr = pbos
			}
		}
		{
			goorl := len(goo.IntAr)
			if goorl == 0 {
				pbo.IntAr = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntAr[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntAr = pbos
			}
		}
		{
			goorl := len(goo.ByteAr)
			if goorl == 0 {
				pbo.ByteAr = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteAr[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteAr = pbos
			}
		}
		{
			goorl := len(goo.Uint8Ar)
			if goorl == 0 {
				pbo.Uint8Ar = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8Ar[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint16Ar)
			if goorl == 0 {
				pbo.Uint16Ar = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16Ar[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint32Ar)
			if goorl == 0 {
				pbo.Uint32Ar = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32Ar[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedAr)
			if goorl == 0 {
				pbo.Uint32FixedAr = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedAr[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32FixedAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64Ar)
			if goorl == 0 {
				pbo.Uint64Ar = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64Ar[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64Ar = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedAr)
			if goorl == 0 {
				pbo.Uint64FixedAr = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedAr[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64FixedAr = pbos
			}
		}
		{
			goorl := len(goo.UintAr)
			if goorl == 0 {
				pbo.UintAr = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintAr[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintAr = pbos
			}
		}
		{
			goorl := len(goo.StrAr)
			if goorl == 0 {
				pbo.StrAr = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrAr[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.StrAr = pbos
			}
		}
		{
			goorl := len(goo.BytesAr)
			if goorl == 0 {
				pbo.BytesAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesAr[i]
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
				pbo.BytesAr = pbos
			}
		}
		{
			goorl := len(goo.TimeAr)
			if goorl == 0 {
				pbo.TimeAr = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeAr[i]
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
			goorl := len(goo.DurationAr)
			if goorl == 0 {
				pbo.DurationAr = nil
			} else {
				var pbos = make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationAr[i]
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
			goorl := len(goo.EmptyAr)
			if goorl == 0 {
				pbo.EmptyAr = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyAr[i]
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
func (goo ArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ArraysStruct)
	msg = pbo
	return
}
func (goo *ArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysStruct = msg.(*testspb.ArraysStruct)
	{
		if pbo != nil {
			{
				var goors = [4]int8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int8Ar[i]
						{
							pboev := pboe
							goors[i] = int8(int8(pboev))
						}
					}
				}
				(*goo).Int8Ar = goors
			}
			{
				var goors = [4]int16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int16Ar[i]
						{
							pboev := pboe
							goors[i] = int16(int16(pboev))
						}
					}
				}
				(*goo).Int16Ar = goors
			}
			{
				var goors = [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32Ar[i]
						{
							pboev := pboe
							goors[i] = int32(pboev)
						}
					}
				}
				(*goo).Int32Ar = goors
			}
			{
				var goors = [4]int32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int32FixedAr[i]
						{
							pboev := pboe
							goors[i] = int32(pboev)
						}
					}
				}
				(*goo).Int32FixedAr = goors
			}
			{
				var goors = [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64Ar[i]
						{
							pboev := pboe
							goors[i] = int64(pboev)
						}
					}
				}
				(*goo).Int64Ar = goors
			}
			{
				var goors = [4]int64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Int64FixedAr[i]
						{
							pboev := pboe
							goors[i] = int64(pboev)
						}
					}
				}
				(*goo).Int64FixedAr = goors
			}
			{
				var goors = [4]int{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.IntAr[i]
						{
							pboev := pboe
							goors[i] = int(int(pboev))
						}
					}
				}
				(*goo).IntAr = goors
			}
			{
				var goors = [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.ByteAr[i]
						{
							pboev := pboe
							goors[i] = uint8(uint8(pboev))
						}
					}
				}
				(*goo).ByteAr = goors
			}
			{
				var goors = [4]uint8{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint8Ar[i]
						{
							pboev := pboe
							goors[i] = uint8(uint8(pboev))
						}
					}
				}
				(*goo).Uint8Ar = goors
			}
			{
				var goors = [4]uint16{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint16Ar[i]
						{
							pboev := pboe
							goors[i] = uint16(uint16(pboev))
						}
					}
				}
				(*goo).Uint16Ar = goors
			}
			{
				var goors = [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32Ar[i]
						{
							pboev := pboe
							goors[i] = uint32(pboev)
						}
					}
				}
				(*goo).Uint32Ar = goors
			}
			{
				var goors = [4]uint32{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint32FixedAr[i]
						{
							pboev := pboe
							goors[i] = uint32(pboev)
						}
					}
				}
				(*goo).Uint32FixedAr = goors
			}
			{
				var goors = [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64Ar[i]
						{
							pboev := pboe
							goors[i] = uint64(pboev)
						}
					}
				}
				(*goo).Uint64Ar = goors
			}
			{
				var goors = [4]uint64{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.Uint64FixedAr[i]
						{
							pboev := pboe
							goors[i] = uint64(pboev)
						}
					}
				}
				(*goo).Uint64FixedAr = goors
			}
			{
				var goors = [4]uint{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.UintAr[i]
						{
							pboev := pboe
							goors[i] = uint(uint(pboev))
						}
					}
				}
				(*goo).UintAr = goors
			}
			{
				var goors = [4]string{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.StrAr[i]
						{
							pboev := pboe
							goors[i] = string(pboev)
						}
					}
				}
				(*goo).StrAr = goors
			}
			{
				var goors = [4][]uint8{}
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
								var goors1 = make([]uint8, pbol)
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
				(*goo).BytesAr = goors
			}
			{
				var goors = [4]time.Time{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.TimeAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsTime()
						}
					}
				}
				(*goo).TimeAr = goors
			}
			{
				var goors = [4]time.Duration{}
				for i := 0; i < 4; i += 1 {
					{
						pboe := pbo.DurationAr[i]
						{
							pboev := pboe
							goors[i] = pboev.AsDuration()
						}
					}
				}
				(*goo).DurationAr = goors
			}
			{
				var goors = [4]EmptyStruct{}
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
				(*goo).EmptyAr = goors
			}
		}
	}
	return
}
func (_ ArraysStruct) GetTypeURL() (typeURL string) {
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
func (goo ArraysArraysStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ArraysArraysStruct
	{
		if IsArraysArraysStructReprEmpty(goo) {
			var pbov *testspb.ArraysArraysStruct
			msg = pbov
			return
		}
		pbo = new(testspb.ArraysArraysStruct)
		{
			goorl := len(goo.Int8ArAr)
			if goorl == 0 {
				pbo.Int8ArAr = nil
			} else {
				var pbos = make([]*testspb.Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int16ArAr)
			if goorl == 0 {
				pbo.Int16ArAr = nil
			} else {
				var pbos = make([]*testspb.Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int32ArAr)
			if goorl == 0 {
				pbo.Int32ArAr = nil
			} else {
				var pbos = make([]*testspb.Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedArAr)
			if goorl == 0 {
				pbo.Int32FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed32Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed32Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.Int64ArAr)
			if goorl == 0 {
				pbo.Int64ArAr = nil
			} else {
				var pbos = make([]*testspb.Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64ArAr = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedArAr)
			if goorl == 0 {
				pbo.Int64FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed64Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed64Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.IntArAr)
			if goorl == 0 {
				pbo.IntArAr = nil
			} else {
				var pbos = make([]*testspb.Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntArAr = pbos
			}
		}
		{
			goorl := len(goo.ByteArAr)
			if goorl == 0 {
				pbo.ByteArAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteArAr[i]
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
				pbo.ByteArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint8ArAr)
			if goorl == 0 {
				pbo.Uint8ArAr = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8ArAr[i]
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
				pbo.Uint8ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint16ArAr)
			if goorl == 0 {
				pbo.Uint16ArAr = nil
			} else {
				var pbos = make([]*testspb.UInt16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint32ArAr)
			if goorl == 0 {
				pbo.Uint32ArAr = nil
			} else {
				var pbos = make([]*testspb.UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedArAr)
			if goorl == 0 {
				pbo.Uint32FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed32UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed32UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64ArAr)
			if goorl == 0 {
				pbo.Uint64ArAr = nil
			} else {
				var pbos = make([]*testspb.UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64ArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64ArAr = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedArAr)
			if goorl == 0 {
				pbo.Uint64FixedArAr = nil
			} else {
				var pbos = make([]*testspb.Fixed64UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed64UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedArAr = pbos
			}
		}
		{
			goorl := len(goo.UintArAr)
			if goorl == 0 {
				pbo.UintArAr = nil
			} else {
				var pbos = make([]*testspb.UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintArAr = pbos
			}
		}
		{
			goorl := len(goo.StrArAr)
			if goorl == 0 {
				pbo.StrArAr = nil
			} else {
				var pbos = make([]*testspb.StringValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = string(goore)
										}
									}
								}
								pbos[i] = &testspb.StringValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrArAr = pbos
			}
		}
		{
			goorl := len(goo.BytesArAr)
			if goorl == 0 {
				pbo.BytesArAr = nil
			} else {
				var pbos = make([]*testspb.BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												var pbos2 = make([]uint8, goorl2)
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
								pbos[i] = &testspb.BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesArAr = pbos
			}
		}
		{
			goorl := len(goo.TimeArAr)
			if goorl == 0 {
				pbo.TimeArAr = nil
			} else {
				var pbos = make([]*testspb.TimestampList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*timestamppb.Timestamp, goorl1)
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
								pbos[i] = &testspb.TimestampList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeArAr = pbos
			}
		}
		{
			goorl := len(goo.DurationArAr)
			if goorl == 0 {
				pbo.DurationArAr = nil
			} else {
				var pbos = make([]*testspb.DurationList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*durationpb.Duration, goorl1)
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
								pbos[i] = &testspb.DurationList{Value: pbos1}
							}
						}
					}
				}
				pbo.DurationArAr = pbos
			}
		}
		{
			goorl := len(goo.EmptyArAr)
			if goorl == 0 {
				pbo.EmptyArAr = nil
			} else {
				var pbos = make([]*testspb.EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyArAr[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*testspb.EmptyStruct, goorl1)
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
								pbos[i] = &testspb.EmptyStructList{Value: pbos1}
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
func (goo ArraysArraysStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ArraysArraysStruct)
	msg = pbo
	return
}
func (goo *ArraysArraysStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ArraysArraysStruct = msg.(*testspb.ArraysArraysStruct)
	{
		if pbo != nil {
			{
				var goors = [2][2]int8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int8ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int8{}
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
				(*goo).Int8ArAr = goors
			}
			{
				var goors = [2][2]int16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int16ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int16{}
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
				(*goo).Int16ArAr = goors
			}
			{
				var goors = [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int32{}
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
				(*goo).Int32ArAr = goors
			}
			{
				var goors = [2][2]int32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int32FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int32{}
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
				(*goo).Int32FixedArAr = goors
			}
			{
				var goors = [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int64{}
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
				(*goo).Int64ArAr = goors
			}
			{
				var goors = [2][2]int64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Int64FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int64{}
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
				(*goo).Int64FixedArAr = goors
			}
			{
				var goors = [2][2]int{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.IntArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]int{}
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
				(*goo).IntArAr = goors
			}
			{
				var goors = [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.ByteArAr[i]
						{
							pboev := pboe
							var goors1 = [2]uint8{}
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
				(*goo).ByteArAr = goors
			}
			{
				var goors = [2][2]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint8ArAr[i]
						{
							pboev := pboe
							var goors1 = [2]uint8{}
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
				(*goo).Uint8ArAr = goors
			}
			{
				var goors = [2][2]uint16{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint16ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint16{}
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
				(*goo).Uint16ArAr = goors
			}
			{
				var goors = [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint32{}
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
				(*goo).Uint32ArAr = goors
			}
			{
				var goors = [2][2]uint32{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint32FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint32{}
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
				(*goo).Uint32FixedArAr = goors
			}
			{
				var goors = [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64ArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint64{}
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
				(*goo).Uint64ArAr = goors
			}
			{
				var goors = [2][2]uint64{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.Uint64FixedArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint64{}
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
				(*goo).Uint64FixedArAr = goors
			}
			{
				var goors = [2][2]uint{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.UintArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]uint{}
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
				(*goo).UintArAr = goors
			}
			{
				var goors = [2][2]string{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.StrArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]string{}
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
				(*goo).StrArAr = goors
			}
			{
				var goors = [2][2][]uint8{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.BytesArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2][]uint8{}
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
												var goors2 = make([]uint8, pbol)
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
				(*goo).BytesArAr = goors
			}
			{
				var goors = [2][2]time.Time{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.TimeArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]time.Time{}
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
				(*goo).TimeArAr = goors
			}
			{
				var goors = [2][2]time.Duration{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.DurationArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]time.Duration{}
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
				(*goo).DurationArAr = goors
			}
			{
				var goors = [2][2]EmptyStruct{}
				for i := 0; i < 2; i += 1 {
					{
						pboe := pbo.EmptyArAr[i]
						if pboe != nil {
							{
								pboev := pboe.Value
								var goors1 = [2]EmptyStruct{}
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
				(*goo).EmptyArAr = goors
			}
		}
	}
	return
}
func (_ ArraysArraysStruct) GetTypeURL() (typeURL string) {
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
func (goo SlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesStruct
	{
		if IsSlicesStructReprEmpty(goo) {
			var pbov *testspb.SlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesStruct)
		{
			goorl := len(goo.Int8Sl)
			if goorl == 0 {
				pbo.Int8Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int8Sl = pbos
			}
		}
		{
			goorl := len(goo.Int16Sl)
			if goorl == 0 {
				pbo.Int16Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int16Sl = pbos
			}
		}
		{
			goorl := len(goo.Int32Sl)
			if goorl == 0 {
				pbo.Int32Sl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32Sl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32Sl = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedSl)
			if goorl == 0 {
				pbo.Int32FixedSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedSl[i]
						{
							pbos[i] = int32(goore)
						}
					}
				}
				pbo.Int32FixedSl = pbos
			}
		}
		{
			goorl := len(goo.Int64Sl)
			if goorl == 0 {
				pbo.Int64Sl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64Sl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64Sl = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedSl)
			if goorl == 0 {
				pbo.Int64FixedSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedSl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.Int64FixedSl = pbos
			}
		}
		{
			goorl := len(goo.IntSl)
			if goorl == 0 {
				pbo.IntSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntSl[i]
						{
							pbos[i] = int64(goore)
						}
					}
				}
				pbo.IntSl = pbos
			}
		}
		{
			goorl := len(goo.ByteSl)
			if goorl == 0 {
				pbo.ByteSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteSl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.ByteSl = pbos
			}
		}
		{
			goorl := len(goo.Uint8Sl)
			if goorl == 0 {
				pbo.Uint8Sl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8Sl[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Uint8Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint16Sl)
			if goorl == 0 {
				pbo.Uint16Sl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16Sl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint16Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint32Sl)
			if goorl == 0 {
				pbo.Uint32Sl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32Sl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedSl)
			if goorl == 0 {
				pbo.Uint32FixedSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedSl[i]
						{
							pbos[i] = uint32(goore)
						}
					}
				}
				pbo.Uint32FixedSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64Sl)
			if goorl == 0 {
				pbo.Uint64Sl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64Sl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64Sl = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedSl)
			if goorl == 0 {
				pbo.Uint64FixedSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedSl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.Uint64FixedSl = pbos
			}
		}
		{
			goorl := len(goo.UintSl)
			if goorl == 0 {
				pbo.UintSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintSl[i]
						{
							pbos[i] = uint64(goore)
						}
					}
				}
				pbo.UintSl = pbos
			}
		}
		{
			goorl := len(goo.StrSl)
			if goorl == 0 {
				pbo.StrSl = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrSl[i]
						{
							pbos[i] = string(goore)
						}
					}
				}
				pbo.StrSl = pbos
			}
		}
		{
			goorl := len(goo.BytesSl)
			if goorl == 0 {
				pbo.BytesSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesSl[i]
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
				pbo.BytesSl = pbos
			}
		}
		{
			goorl := len(goo.TimeSl)
			if goorl == 0 {
				pbo.TimeSl = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeSl[i]
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
			goorl := len(goo.DurationSl)
			if goorl == 0 {
				pbo.DurationSl = nil
			} else {
				var pbos = make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationSl[i]
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
			goorl := len(goo.EmptySl)
			if goorl == 0 {
				pbo.EmptySl = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptySl[i]
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
func (goo SlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.SlicesStruct)
	msg = pbo
	return
}
func (goo *SlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesStruct = msg.(*testspb.SlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8Sl != nil {
					pbol = len(pbo.Int8Sl)
				}
				if pbol == 0 {
					(*goo).Int8Sl = nil
				} else {
					var goors = make([]int8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int8Sl[i]
							{
								pboev := pboe
								goors[i] = int8(int8(pboev))
							}
						}
					}
					(*goo).Int8Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16Sl != nil {
					pbol = len(pbo.Int16Sl)
				}
				if pbol == 0 {
					(*goo).Int16Sl = nil
				} else {
					var goors = make([]int16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int16Sl[i]
							{
								pboev := pboe
								goors[i] = int16(int16(pboev))
							}
						}
					}
					(*goo).Int16Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32Sl != nil {
					pbol = len(pbo.Int32Sl)
				}
				if pbol == 0 {
					(*goo).Int32Sl = nil
				} else {
					var goors = make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32Sl[i]
							{
								pboev := pboe
								goors[i] = int32(pboev)
							}
						}
					}
					(*goo).Int32Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSl != nil {
					pbol = len(pbo.Int32FixedSl)
				}
				if pbol == 0 {
					(*goo).Int32FixedSl = nil
				} else {
					var goors = make([]int32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int32FixedSl[i]
							{
								pboev := pboe
								goors[i] = int32(pboev)
							}
						}
					}
					(*goo).Int32FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64Sl != nil {
					pbol = len(pbo.Int64Sl)
				}
				if pbol == 0 {
					(*goo).Int64Sl = nil
				} else {
					var goors = make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64Sl[i]
							{
								pboev := pboe
								goors[i] = int64(pboev)
							}
						}
					}
					(*goo).Int64Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSl != nil {
					pbol = len(pbo.Int64FixedSl)
				}
				if pbol == 0 {
					(*goo).Int64FixedSl = nil
				} else {
					var goors = make([]int64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Int64FixedSl[i]
							{
								pboev := pboe
								goors[i] = int64(pboev)
							}
						}
					}
					(*goo).Int64FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntSl != nil {
					pbol = len(pbo.IntSl)
				}
				if pbol == 0 {
					(*goo).IntSl = nil
				} else {
					var goors = make([]int, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.IntSl[i]
							{
								pboev := pboe
								goors[i] = int(int(pboev))
							}
						}
					}
					(*goo).IntSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSl != nil {
					pbol = len(pbo.ByteSl)
				}
				if pbol == 0 {
					(*goo).ByteSl = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.ByteSl[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).ByteSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8Sl != nil {
					pbol = len(pbo.Uint8Sl)
				}
				if pbol == 0 {
					(*goo).Uint8Sl = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint8Sl[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Uint8Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16Sl != nil {
					pbol = len(pbo.Uint16Sl)
				}
				if pbol == 0 {
					(*goo).Uint16Sl = nil
				} else {
					var goors = make([]uint16, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint16Sl[i]
							{
								pboev := pboe
								goors[i] = uint16(uint16(pboev))
							}
						}
					}
					(*goo).Uint16Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32Sl != nil {
					pbol = len(pbo.Uint32Sl)
				}
				if pbol == 0 {
					(*goo).Uint32Sl = nil
				} else {
					var goors = make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32Sl[i]
							{
								pboev := pboe
								goors[i] = uint32(pboev)
							}
						}
					}
					(*goo).Uint32Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSl != nil {
					pbol = len(pbo.Uint32FixedSl)
				}
				if pbol == 0 {
					(*goo).Uint32FixedSl = nil
				} else {
					var goors = make([]uint32, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint32FixedSl[i]
							{
								pboev := pboe
								goors[i] = uint32(pboev)
							}
						}
					}
					(*goo).Uint32FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64Sl != nil {
					pbol = len(pbo.Uint64Sl)
				}
				if pbol == 0 {
					(*goo).Uint64Sl = nil
				} else {
					var goors = make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64Sl[i]
							{
								pboev := pboe
								goors[i] = uint64(pboev)
							}
						}
					}
					(*goo).Uint64Sl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSl != nil {
					pbol = len(pbo.Uint64FixedSl)
				}
				if pbol == 0 {
					(*goo).Uint64FixedSl = nil
				} else {
					var goors = make([]uint64, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Uint64FixedSl[i]
							{
								pboev := pboe
								goors[i] = uint64(pboev)
							}
						}
					}
					(*goo).Uint64FixedSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintSl != nil {
					pbol = len(pbo.UintSl)
				}
				if pbol == 0 {
					(*goo).UintSl = nil
				} else {
					var goors = make([]uint, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.UintSl[i]
							{
								pboev := pboe
								goors[i] = uint(uint(pboev))
							}
						}
					}
					(*goo).UintSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrSl != nil {
					pbol = len(pbo.StrSl)
				}
				if pbol == 0 {
					(*goo).StrSl = nil
				} else {
					var goors = make([]string, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.StrSl[i]
							{
								pboev := pboe
								goors[i] = string(pboev)
							}
						}
					}
					(*goo).StrSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSl != nil {
					pbol = len(pbo.BytesSl)
				}
				if pbol == 0 {
					(*goo).BytesSl = nil
				} else {
					var goors = make([][]uint8, pbol)
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
					(*goo).BytesSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSl != nil {
					pbol = len(pbo.TimeSl)
				}
				if pbol == 0 {
					(*goo).TimeSl = nil
				} else {
					var goors = make([]time.Time, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.TimeSl[i]
							{
								pboev := pboe
								goors[i] = pboev.AsTime()
							}
						}
					}
					(*goo).TimeSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationSl != nil {
					pbol = len(pbo.DurationSl)
				}
				if pbol == 0 {
					(*goo).DurationSl = nil
				} else {
					var goors = make([]time.Duration, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.DurationSl[i]
							{
								pboev := pboe
								goors[i] = pboev.AsDuration()
							}
						}
					}
					(*goo).DurationSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySl != nil {
					pbol = len(pbo.EmptySl)
				}
				if pbol == 0 {
					(*goo).EmptySl = nil
				} else {
					var goors = make([]EmptyStruct, pbol)
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
					(*goo).EmptySl = goors
				}
			}
		}
	}
	return
}
func (_ SlicesStruct) GetTypeURL() (typeURL string) {
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
func (goo SlicesSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.SlicesSlicesStruct
	{
		if IsSlicesSlicesStructReprEmpty(goo) {
			var pbov *testspb.SlicesSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.SlicesSlicesStruct)
		{
			goorl := len(goo.Int8SlSl)
			if goorl == 0 {
				pbo.Int8SlSl = nil
			} else {
				var pbos = make([]*testspb.Int8List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int8List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int8SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int16SlSl)
			if goorl == 0 {
				pbo.Int16SlSl = nil
			} else {
				var pbos = make([]*testspb.Int16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Int16SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int32SlSl)
			if goorl == 0 {
				pbo.Int32SlSl = nil
			} else {
				var pbos = make([]*testspb.Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int32FixedSlSl)
			if goorl == 0 {
				pbo.Int32FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed32Int32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int32(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed32Int32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int32FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.Int64SlSl)
			if goorl == 0 {
				pbo.Int64SlSl = nil
			} else {
				var pbos = make([]*testspb.Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64SlSl = pbos
			}
		}
		{
			goorl := len(goo.Int64FixedSlSl)
			if goorl == 0 {
				pbo.Int64FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed64Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed64Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Int64FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.IntSlSl)
			if goorl == 0 {
				pbo.IntSlSl = nil
			} else {
				var pbos = make([]*testspb.Int64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]int64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = int64(goore)
										}
									}
								}
								pbos[i] = &testspb.Int64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.IntSlSl = pbos
			}
		}
		{
			goorl := len(goo.ByteSlSl)
			if goorl == 0 {
				pbo.ByteSlSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.ByteSlSl[i]
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
				pbo.ByteSlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint8SlSl)
			if goorl == 0 {
				pbo.Uint8SlSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8SlSl[i]
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
				pbo.Uint8SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint16SlSl)
			if goorl == 0 {
				pbo.Uint16SlSl = nil
			} else {
				var pbos = make([]*testspb.UInt16List, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt16List{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint16SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32SlSl)
			if goorl == 0 {
				pbo.Uint32SlSl = nil
			} else {
				var pbos = make([]*testspb.UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint32FixedSlSl)
			if goorl == 0 {
				pbo.Uint32FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed32UInt32ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint32, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint32(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed32UInt32ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint32FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64SlSl)
			if goorl == 0 {
				pbo.Uint64SlSl = nil
			} else {
				var pbos = make([]*testspb.UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64SlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64SlSl = pbos
			}
		}
		{
			goorl := len(goo.Uint64FixedSlSl)
			if goorl == 0 {
				pbo.Uint64FixedSlSl = nil
			} else {
				var pbos = make([]*testspb.Fixed64UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.Fixed64UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.Uint64FixedSlSl = pbos
			}
		}
		{
			goorl := len(goo.UintSlSl)
			if goorl == 0 {
				pbo.UintSlSl = nil
			} else {
				var pbos = make([]*testspb.UInt64ValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]uint64, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = uint64(goore)
										}
									}
								}
								pbos[i] = &testspb.UInt64ValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.UintSlSl = pbos
			}
		}
		{
			goorl := len(goo.StrSlSl)
			if goorl == 0 {
				pbo.StrSlSl = nil
			} else {
				var pbos = make([]*testspb.StringValueList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]string, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											pbos1[i] = string(goore)
										}
									}
								}
								pbos[i] = &testspb.StringValueList{Value: pbos1}
							}
						}
					}
				}
				pbo.StrSlSl = pbos
			}
		}
		{
			goorl := len(goo.BytesSlSl)
			if goorl == 0 {
				pbo.BytesSlSl = nil
			} else {
				var pbos = make([]*testspb.BytesList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([][]byte, goorl1)
								for i := 0; i < goorl1; i += 1 {
									{
										goore := goore[i]
										{
											goorl2 := len(goore)
											if goorl2 == 0 {
												pbos1[i] = nil
											} else {
												var pbos2 = make([]uint8, goorl2)
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
								pbos[i] = &testspb.BytesList{Value: pbos1}
							}
						}
					}
				}
				pbo.BytesSlSl = pbos
			}
		}
		{
			goorl := len(goo.TimeSlSl)
			if goorl == 0 {
				pbo.TimeSlSl = nil
			} else {
				var pbos = make([]*testspb.TimestampList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimeSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*timestamppb.Timestamp, goorl1)
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
								pbos[i] = &testspb.TimestampList{Value: pbos1}
							}
						}
					}
				}
				pbo.TimeSlSl = pbos
			}
		}
		{
			goorl := len(goo.DurationSlSl)
			if goorl == 0 {
				pbo.DurationSlSl = nil
			} else {
				var pbos = make([]*testspb.DurationList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationSlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*durationpb.Duration, goorl1)
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
								pbos[i] = &testspb.DurationList{Value: pbos1}
							}
						}
					}
				}
				pbo.DurationSlSl = pbos
			}
		}
		{
			goorl := len(goo.EmptySlSl)
			if goorl == 0 {
				pbo.EmptySlSl = nil
			} else {
				var pbos = make([]*testspb.EmptyStructList, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptySlSl[i]
						{
							goorl1 := len(goore)
							if goorl1 == 0 {
								pbos[i] = nil
							} else {
								var pbos1 = make([]*testspb.EmptyStruct, goorl1)
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
								pbos[i] = &testspb.EmptyStructList{Value: pbos1}
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
func (goo SlicesSlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.SlicesSlicesStruct)
	msg = pbo
	return
}
func (goo *SlicesSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.SlicesSlicesStruct = msg.(*testspb.SlicesSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8SlSl != nil {
					pbol = len(pbo.Int8SlSl)
				}
				if pbol == 0 {
					(*goo).Int8SlSl = nil
				} else {
					var goors = make([][]int8, pbol)
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
										var goors1 = make([]int8, pbol1)
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
					(*goo).Int8SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16SlSl != nil {
					pbol = len(pbo.Int16SlSl)
				}
				if pbol == 0 {
					(*goo).Int16SlSl = nil
				} else {
					var goors = make([][]int16, pbol)
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
										var goors1 = make([]int16, pbol1)
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
					(*goo).Int16SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32SlSl != nil {
					pbol = len(pbo.Int32SlSl)
				}
				if pbol == 0 {
					(*goo).Int32SlSl = nil
				} else {
					var goors = make([][]int32, pbol)
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
										var goors1 = make([]int32, pbol1)
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
					(*goo).Int32SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedSlSl != nil {
					pbol = len(pbo.Int32FixedSlSl)
				}
				if pbol == 0 {
					(*goo).Int32FixedSlSl = nil
				} else {
					var goors = make([][]int32, pbol)
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
										var goors1 = make([]int32, pbol1)
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
					(*goo).Int32FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64SlSl != nil {
					pbol = len(pbo.Int64SlSl)
				}
				if pbol == 0 {
					(*goo).Int64SlSl = nil
				} else {
					var goors = make([][]int64, pbol)
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
										var goors1 = make([]int64, pbol1)
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
					(*goo).Int64SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedSlSl != nil {
					pbol = len(pbo.Int64FixedSlSl)
				}
				if pbol == 0 {
					(*goo).Int64FixedSlSl = nil
				} else {
					var goors = make([][]int64, pbol)
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
										var goors1 = make([]int64, pbol1)
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
					(*goo).Int64FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntSlSl != nil {
					pbol = len(pbo.IntSlSl)
				}
				if pbol == 0 {
					(*goo).IntSlSl = nil
				} else {
					var goors = make([][]int, pbol)
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
										var goors1 = make([]int, pbol1)
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
					(*goo).IntSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.ByteSlSl != nil {
					pbol = len(pbo.ByteSlSl)
				}
				if pbol == 0 {
					(*goo).ByteSlSl = nil
				} else {
					var goors = make([][]uint8, pbol)
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
					(*goo).ByteSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8SlSl != nil {
					pbol = len(pbo.Uint8SlSl)
				}
				if pbol == 0 {
					(*goo).Uint8SlSl = nil
				} else {
					var goors = make([][]uint8, pbol)
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
					(*goo).Uint8SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16SlSl != nil {
					pbol = len(pbo.Uint16SlSl)
				}
				if pbol == 0 {
					(*goo).Uint16SlSl = nil
				} else {
					var goors = make([][]uint16, pbol)
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
										var goors1 = make([]uint16, pbol1)
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
					(*goo).Uint16SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32SlSl != nil {
					pbol = len(pbo.Uint32SlSl)
				}
				if pbol == 0 {
					(*goo).Uint32SlSl = nil
				} else {
					var goors = make([][]uint32, pbol)
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
										var goors1 = make([]uint32, pbol1)
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
					(*goo).Uint32SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedSlSl != nil {
					pbol = len(pbo.Uint32FixedSlSl)
				}
				if pbol == 0 {
					(*goo).Uint32FixedSlSl = nil
				} else {
					var goors = make([][]uint32, pbol)
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
										var goors1 = make([]uint32, pbol1)
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
					(*goo).Uint32FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64SlSl != nil {
					pbol = len(pbo.Uint64SlSl)
				}
				if pbol == 0 {
					(*goo).Uint64SlSl = nil
				} else {
					var goors = make([][]uint64, pbol)
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
										var goors1 = make([]uint64, pbol1)
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
					(*goo).Uint64SlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedSlSl != nil {
					pbol = len(pbo.Uint64FixedSlSl)
				}
				if pbol == 0 {
					(*goo).Uint64FixedSlSl = nil
				} else {
					var goors = make([][]uint64, pbol)
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
										var goors1 = make([]uint64, pbol1)
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
					(*goo).Uint64FixedSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintSlSl != nil {
					pbol = len(pbo.UintSlSl)
				}
				if pbol == 0 {
					(*goo).UintSlSl = nil
				} else {
					var goors = make([][]uint, pbol)
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
										var goors1 = make([]uint, pbol1)
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
					(*goo).UintSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrSlSl != nil {
					pbol = len(pbo.StrSlSl)
				}
				if pbol == 0 {
					(*goo).StrSlSl = nil
				} else {
					var goors = make([][]string, pbol)
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
										var goors1 = make([]string, pbol1)
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
					(*goo).StrSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesSlSl != nil {
					pbol = len(pbo.BytesSlSl)
				}
				if pbol == 0 {
					(*goo).BytesSlSl = nil
				} else {
					var goors = make([][][]uint8, pbol)
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
										var goors1 = make([][]uint8, pbol1)
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
														var goors2 = make([]uint8, pbol2)
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
					(*goo).BytesSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimeSlSl != nil {
					pbol = len(pbo.TimeSlSl)
				}
				if pbol == 0 {
					(*goo).TimeSlSl = nil
				} else {
					var goors = make([][]time.Time, pbol)
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
										var goors1 = make([]time.Time, pbol1)
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
					(*goo).TimeSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationSlSl != nil {
					pbol = len(pbo.DurationSlSl)
				}
				if pbol == 0 {
					(*goo).DurationSlSl = nil
				} else {
					var goors = make([][]time.Duration, pbol)
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
										var goors1 = make([]time.Duration, pbol1)
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
					(*goo).DurationSlSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptySlSl != nil {
					pbol = len(pbo.EmptySlSl)
				}
				if pbol == 0 {
					(*goo).EmptySlSl = nil
				} else {
					var goors = make([][]EmptyStruct, pbol)
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
										var goors1 = make([]EmptyStruct, pbol1)
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
					(*goo).EmptySlSl = goors
				}
			}
		}
	}
	return
}
func (_ SlicesSlicesStruct) GetTypeURL() (typeURL string) {
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
func (goo PointersStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointersStruct
	{
		if IsPointersStructReprEmpty(goo) {
			var pbov *testspb.PointersStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointersStruct)
		{
			if goo.Int8Pt != nil {
				dgoor := *goo.Int8Pt
				dgoor = dgoor
				pbo.Int8Pt = int32(dgoor)
			}
		}
		{
			if goo.Int16Pt != nil {
				dgoor := *goo.Int16Pt
				dgoor = dgoor
				pbo.Int16Pt = int32(dgoor)
			}
		}
		{
			if goo.Int32Pt != nil {
				dgoor := *goo.Int32Pt
				dgoor = dgoor
				pbo.Int32Pt = int32(dgoor)
			}
		}
		{
			if goo.Int32FixedPt != nil {
				dgoor := *goo.Int32FixedPt
				dgoor = dgoor
				pbo.Int32FixedPt = int32(dgoor)
			}
		}
		{
			if goo.Int64Pt != nil {
				dgoor := *goo.Int64Pt
				dgoor = dgoor
				pbo.Int64Pt = int64(dgoor)
			}
		}
		{
			if goo.Int64FixedPt != nil {
				dgoor := *goo.Int64FixedPt
				dgoor = dgoor
				pbo.Int64FixedPt = int64(dgoor)
			}
		}
		{
			if goo.IntPt != nil {
				dgoor := *goo.IntPt
				dgoor = dgoor
				pbo.IntPt = int64(dgoor)
			}
		}
		{
			if goo.BytePt != nil {
				dgoor := *goo.BytePt
				dgoor = dgoor
				pbo.BytePt = uint32(dgoor)
			}
		}
		{
			if goo.Uint8Pt != nil {
				dgoor := *goo.Uint8Pt
				dgoor = dgoor
				pbo.Uint8Pt = uint32(dgoor)
			}
		}
		{
			if goo.Uint16Pt != nil {
				dgoor := *goo.Uint16Pt
				dgoor = dgoor
				pbo.Uint16Pt = uint32(dgoor)
			}
		}
		{
			if goo.Uint32Pt != nil {
				dgoor := *goo.Uint32Pt
				dgoor = dgoor
				pbo.Uint32Pt = uint32(dgoor)
			}
		}
		{
			if goo.Uint32FixedPt != nil {
				dgoor := *goo.Uint32FixedPt
				dgoor = dgoor
				pbo.Uint32FixedPt = uint32(dgoor)
			}
		}
		{
			if goo.Uint64Pt != nil {
				dgoor := *goo.Uint64Pt
				dgoor = dgoor
				pbo.Uint64Pt = uint64(dgoor)
			}
		}
		{
			if goo.Uint64FixedPt != nil {
				dgoor := *goo.Uint64FixedPt
				dgoor = dgoor
				pbo.Uint64FixedPt = uint64(dgoor)
			}
		}
		{
			if goo.UintPt != nil {
				dgoor := *goo.UintPt
				dgoor = dgoor
				pbo.UintPt = uint64(dgoor)
			}
		}
		{
			if goo.StrPt != nil {
				dgoor := *goo.StrPt
				dgoor = dgoor
				pbo.StrPt = string(dgoor)
			}
		}
		{
			if goo.BytesPt != nil {
				dgoor := *goo.BytesPt
				dgoor = dgoor
				goorl := len(dgoor)
				if goorl == 0 {
					pbo.BytesPt = nil
				} else {
					var pbos = make([]uint8, goorl)
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
			if goo.TimePt != nil {
				dgoor := *goo.TimePt
				dgoor = dgoor
				pbo.TimePt = timestamppb.New(dgoor)
			}
		}
		{
			if goo.DurationPt != nil {
				dgoor := *goo.DurationPt
				dgoor = dgoor
				pbo.DurationPt = durationpb.New(dgoor)
			}
		}
		{
			if goo.EmptyPt != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.EmptyPt.ToPBMessage(cdc)
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
func (goo PointersStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PointersStruct)
	msg = pbo
	return
}
func (goo *PointersStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointersStruct = msg.(*testspb.PointersStruct)
	{
		if pbo != nil {
			{
				(*goo).Int8Pt = new(int8)
				*(*goo).Int8Pt = int8(int8(pbo.Int8Pt))
			}
			{
				(*goo).Int16Pt = new(int16)
				*(*goo).Int16Pt = int16(int16(pbo.Int16Pt))
			}
			{
				(*goo).Int32Pt = new(int32)
				*(*goo).Int32Pt = int32(pbo.Int32Pt)
			}
			{
				(*goo).Int32FixedPt = new(int32)
				*(*goo).Int32FixedPt = int32(pbo.Int32FixedPt)
			}
			{
				(*goo).Int64Pt = new(int64)
				*(*goo).Int64Pt = int64(pbo.Int64Pt)
			}
			{
				(*goo).Int64FixedPt = new(int64)
				*(*goo).Int64FixedPt = int64(pbo.Int64FixedPt)
			}
			{
				(*goo).IntPt = new(int)
				*(*goo).IntPt = int(int(pbo.IntPt))
			}
			{
				(*goo).BytePt = new(uint8)
				*(*goo).BytePt = uint8(uint8(pbo.BytePt))
			}
			{
				(*goo).Uint8Pt = new(uint8)
				*(*goo).Uint8Pt = uint8(uint8(pbo.Uint8Pt))
			}
			{
				(*goo).Uint16Pt = new(uint16)
				*(*goo).Uint16Pt = uint16(uint16(pbo.Uint16Pt))
			}
			{
				(*goo).Uint32Pt = new(uint32)
				*(*goo).Uint32Pt = uint32(pbo.Uint32Pt)
			}
			{
				(*goo).Uint32FixedPt = new(uint32)
				*(*goo).Uint32FixedPt = uint32(pbo.Uint32FixedPt)
			}
			{
				(*goo).Uint64Pt = new(uint64)
				*(*goo).Uint64Pt = uint64(pbo.Uint64Pt)
			}
			{
				(*goo).Uint64FixedPt = new(uint64)
				*(*goo).Uint64FixedPt = uint64(pbo.Uint64FixedPt)
			}
			{
				(*goo).UintPt = new(uint)
				*(*goo).UintPt = uint(uint(pbo.UintPt))
			}
			{
				(*goo).StrPt = new(string)
				*(*goo).StrPt = string(pbo.StrPt)
			}
			{
				(*goo).BytesPt = new([]uint8)
				var pbol int = 0
				if pbo.BytesPt != nil {
					pbol = len(pbo.BytesPt)
				}
				if pbol == 0 {
					*(*goo).BytesPt = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.BytesPt[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					*(*goo).BytesPt = goors
				}
			}
			{
				(*goo).TimePt = new(time.Time)
				*(*goo).TimePt = pbo.TimePt.AsTime()
			}
			{
				(*goo).DurationPt = new(time.Duration)
				*(*goo).DurationPt = pbo.DurationPt.AsDuration()
			}
			{
				if pbo.EmptyPt != nil {
					(*goo).EmptyPt = new(EmptyStruct)
					err = (*goo).EmptyPt.FromPBMessage(cdc, pbo.EmptyPt)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PointersStruct) GetTypeURL() (typeURL string) {
	return "/tests.PointersStruct"
}
func IsPointersStructReprEmpty(goor PointersStruct) (empty bool) {
	{
		empty = true
		{
			if goor.Int8Pt != nil {
				dgoor := *goor.Int8Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int16Pt != nil {
				dgoor := *goor.Int16Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int32Pt != nil {
				dgoor := *goor.Int32Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int32FixedPt != nil {
				dgoor := *goor.Int32FixedPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int64Pt != nil {
				dgoor := *goor.Int64Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Int64FixedPt != nil {
				dgoor := *goor.Int64FixedPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.IntPt != nil {
				dgoor := *goor.IntPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.BytePt != nil {
				dgoor := *goor.BytePt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint8Pt != nil {
				dgoor := *goor.Uint8Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint16Pt != nil {
				dgoor := *goor.Uint16Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint32Pt != nil {
				dgoor := *goor.Uint32Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint32FixedPt != nil {
				dgoor := *goor.Uint32FixedPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint64Pt != nil {
				dgoor := *goor.Uint64Pt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.Uint64FixedPt != nil {
				dgoor := *goor.Uint64FixedPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.UintPt != nil {
				dgoor := *goor.UintPt
				dgoor = dgoor
				if dgoor != 0 {
					return false
				}
			}
		}
		{
			if goor.StrPt != nil {
				dgoor := *goor.StrPt
				dgoor = dgoor
				if dgoor != "" {
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
				if dgoor != 0 {
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
func (goo PointerSlicesStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PointerSlicesStruct
	{
		if IsPointerSlicesStructReprEmpty(goo) {
			var pbov *testspb.PointerSlicesStruct
			msg = pbov
			return
		}
		pbo = new(testspb.PointerSlicesStruct)
		{
			goorl := len(goo.Int8PtSl)
			if goorl == 0 {
				pbo.Int8PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int8PtSl[i]
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
			goorl := len(goo.Int16PtSl)
			if goorl == 0 {
				pbo.Int16PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int16PtSl[i]
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
			goorl := len(goo.Int32PtSl)
			if goorl == 0 {
				pbo.Int32PtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32PtSl[i]
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
			goorl := len(goo.Int32FixedPtSl)
			if goorl == 0 {
				pbo.Int32FixedPtSl = nil
			} else {
				var pbos = make([]int32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int32FixedPtSl[i]
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
			goorl := len(goo.Int64PtSl)
			if goorl == 0 {
				pbo.Int64PtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64PtSl[i]
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
			goorl := len(goo.Int64FixedPtSl)
			if goorl == 0 {
				pbo.Int64FixedPtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Int64FixedPtSl[i]
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
			goorl := len(goo.IntPtSl)
			if goorl == 0 {
				pbo.IntPtSl = nil
			} else {
				var pbos = make([]int64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.IntPtSl[i]
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
			goorl := len(goo.BytePtSl)
			if goorl == 0 {
				pbo.BytePtSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytePtSl[i]
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
			goorl := len(goo.Uint8PtSl)
			if goorl == 0 {
				pbo.Uint8PtSl = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint8PtSl[i]
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
			goorl := len(goo.Uint16PtSl)
			if goorl == 0 {
				pbo.Uint16PtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint16PtSl[i]
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
			goorl := len(goo.Uint32PtSl)
			if goorl == 0 {
				pbo.Uint32PtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32PtSl[i]
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
			goorl := len(goo.Uint32FixedPtSl)
			if goorl == 0 {
				pbo.Uint32FixedPtSl = nil
			} else {
				var pbos = make([]uint32, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint32FixedPtSl[i]
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
			goorl := len(goo.Uint64PtSl)
			if goorl == 0 {
				pbo.Uint64PtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64PtSl[i]
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
			goorl := len(goo.Uint64FixedPtSl)
			if goorl == 0 {
				pbo.Uint64FixedPtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Uint64FixedPtSl[i]
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
			goorl := len(goo.UintPtSl)
			if goorl == 0 {
				pbo.UintPtSl = nil
			} else {
				var pbos = make([]uint64, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.UintPtSl[i]
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
			goorl := len(goo.StrPtSl)
			if goorl == 0 {
				pbo.StrPtSl = nil
			} else {
				var pbos = make([]string, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.StrPtSl[i]
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
			goorl := len(goo.BytesPtSl)
			if goorl == 0 {
				pbo.BytesPtSl = nil
			} else {
				var pbos = make([][]byte, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.BytesPtSl[i]
						{
							if goore != nil {
								dgoor := *goore
								dgoor = dgoor
								goorl1 := len(dgoor)
								if goorl1 == 0 {
									pbos[i] = nil
								} else {
									var pbos1 = make([]uint8, goorl1)
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
			goorl := len(goo.TimePtSl)
			if goorl == 0 {
				pbo.TimePtSl = nil
			} else {
				var pbos = make([]*timestamppb.Timestamp, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.TimePtSl[i]
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
			goorl := len(goo.DurationPtSl)
			if goorl == 0 {
				pbo.DurationPtSl = nil
			} else {
				var pbos = make([]*durationpb.Duration, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.DurationPtSl[i]
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
			goorl := len(goo.EmptyPtSl)
			if goorl == 0 {
				pbo.EmptyPtSl = nil
			} else {
				var pbos = make([]*testspb.EmptyStruct, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.EmptyPtSl[i]
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
func (goo PointerSlicesStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PointerSlicesStruct)
	msg = pbo
	return
}
func (goo *PointerSlicesStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PointerSlicesStruct = msg.(*testspb.PointerSlicesStruct)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Int8PtSl != nil {
					pbol = len(pbo.Int8PtSl)
				}
				if pbol == 0 {
					(*goo).Int8PtSl = nil
				} else {
					var goors = make([]*int8, pbol)
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
					(*goo).Int8PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int16PtSl != nil {
					pbol = len(pbo.Int16PtSl)
				}
				if pbol == 0 {
					(*goo).Int16PtSl = nil
				} else {
					var goors = make([]*int16, pbol)
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
					(*goo).Int16PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32PtSl != nil {
					pbol = len(pbo.Int32PtSl)
				}
				if pbol == 0 {
					(*goo).Int32PtSl = nil
				} else {
					var goors = make([]*int32, pbol)
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
					(*goo).Int32PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int32FixedPtSl != nil {
					pbol = len(pbo.Int32FixedPtSl)
				}
				if pbol == 0 {
					(*goo).Int32FixedPtSl = nil
				} else {
					var goors = make([]*int32, pbol)
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
					(*goo).Int32FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64PtSl != nil {
					pbol = len(pbo.Int64PtSl)
				}
				if pbol == 0 {
					(*goo).Int64PtSl = nil
				} else {
					var goors = make([]*int64, pbol)
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
					(*goo).Int64PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Int64FixedPtSl != nil {
					pbol = len(pbo.Int64FixedPtSl)
				}
				if pbol == 0 {
					(*goo).Int64FixedPtSl = nil
				} else {
					var goors = make([]*int64, pbol)
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
					(*goo).Int64FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.IntPtSl != nil {
					pbol = len(pbo.IntPtSl)
				}
				if pbol == 0 {
					(*goo).IntPtSl = nil
				} else {
					var goors = make([]*int, pbol)
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
					(*goo).IntPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytePtSl != nil {
					pbol = len(pbo.BytePtSl)
				}
				if pbol == 0 {
					(*goo).BytePtSl = nil
				} else {
					var goors = make([]*uint8, pbol)
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
					(*goo).BytePtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint8PtSl != nil {
					pbol = len(pbo.Uint8PtSl)
				}
				if pbol == 0 {
					(*goo).Uint8PtSl = nil
				} else {
					var goors = make([]*uint8, pbol)
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
					(*goo).Uint8PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint16PtSl != nil {
					pbol = len(pbo.Uint16PtSl)
				}
				if pbol == 0 {
					(*goo).Uint16PtSl = nil
				} else {
					var goors = make([]*uint16, pbol)
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
					(*goo).Uint16PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32PtSl != nil {
					pbol = len(pbo.Uint32PtSl)
				}
				if pbol == 0 {
					(*goo).Uint32PtSl = nil
				} else {
					var goors = make([]*uint32, pbol)
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
					(*goo).Uint32PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint32FixedPtSl != nil {
					pbol = len(pbo.Uint32FixedPtSl)
				}
				if pbol == 0 {
					(*goo).Uint32FixedPtSl = nil
				} else {
					var goors = make([]*uint32, pbol)
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
					(*goo).Uint32FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64PtSl != nil {
					pbol = len(pbo.Uint64PtSl)
				}
				if pbol == 0 {
					(*goo).Uint64PtSl = nil
				} else {
					var goors = make([]*uint64, pbol)
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
					(*goo).Uint64PtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.Uint64FixedPtSl != nil {
					pbol = len(pbo.Uint64FixedPtSl)
				}
				if pbol == 0 {
					(*goo).Uint64FixedPtSl = nil
				} else {
					var goors = make([]*uint64, pbol)
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
					(*goo).Uint64FixedPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.UintPtSl != nil {
					pbol = len(pbo.UintPtSl)
				}
				if pbol == 0 {
					(*goo).UintPtSl = nil
				} else {
					var goors = make([]*uint, pbol)
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
					(*goo).UintPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.StrPtSl != nil {
					pbol = len(pbo.StrPtSl)
				}
				if pbol == 0 {
					(*goo).StrPtSl = nil
				} else {
					var goors = make([]*string, pbol)
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
					(*goo).StrPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.BytesPtSl != nil {
					pbol = len(pbo.BytesPtSl)
				}
				if pbol == 0 {
					(*goo).BytesPtSl = nil
				} else {
					var goors = make([]*[]uint8, pbol)
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
									*goors[i] = goors1
								}
							}
						}
					}
					(*goo).BytesPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.TimePtSl != nil {
					pbol = len(pbo.TimePtSl)
				}
				if pbol == 0 {
					(*goo).TimePtSl = nil
				} else {
					var goors = make([]*time.Time, pbol)
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
					(*goo).TimePtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.DurationPtSl != nil {
					pbol = len(pbo.DurationPtSl)
				}
				if pbol == 0 {
					(*goo).DurationPtSl = nil
				} else {
					var goors = make([]*time.Duration, pbol)
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
					(*goo).DurationPtSl = goors
				}
			}
			{
				var pbol int = 0
				if pbo.EmptyPtSl != nil {
					pbol = len(pbo.EmptyPtSl)
				}
				if pbol == 0 {
					(*goo).EmptyPtSl = nil
				} else {
					var goors = make([]*EmptyStruct, pbol)
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
					(*goo).EmptyPtSl = goors
				}
			}
		}
	}
	return
}
func (_ PointerSlicesStruct) GetTypeURL() (typeURL string) {
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
func (goo ComplexSt) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ComplexSt
	{
		if IsComplexStReprEmpty(goo) {
			var pbov *testspb.ComplexSt
			msg = pbov
			return
		}
		pbo = new(testspb.ComplexSt)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrField = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArField = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.SlField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlField = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PtField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PtField = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}
func (goo ComplexSt) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ComplexSt)
	msg = pbo
	return
}
func (goo *ComplexSt) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ComplexSt = msg.(*testspb.ComplexSt)
	{
		if pbo != nil {
			{
				if pbo.PrField != nil {
					err = (*goo).PrField.FromPBMessage(cdc, pbo.PrField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArField != nil {
					err = (*goo).ArField.FromPBMessage(cdc, pbo.ArField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlField != nil {
					err = (*goo).SlField.FromPBMessage(cdc, pbo.SlField)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PtField != nil {
					err = (*goo).PtField.FromPBMessage(cdc, pbo.PtField)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ ComplexSt) GetTypeURL() (typeURL string) {
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
func (goo EmbeddedSt1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt1
	{
		if IsEmbeddedSt1ReprEmpty(goo) {
			var pbov *testspb.EmbeddedSt1
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt1)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
	}
	msg = pbo
	return
}
func (goo EmbeddedSt1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt1)
	msg = pbo
	return
}
func (goo *EmbeddedSt1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt1 = msg.(*testspb.EmbeddedSt1)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = (*goo).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt1) GetTypeURL() (typeURL string) {
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
func (goo EmbeddedSt2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt2
	{
		if IsEmbeddedSt2ReprEmpty(goo) {
			var pbov *testspb.EmbeddedSt2
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt2)
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArraysStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStruct = pbom.(*testspb.ArraysStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PointersStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStruct = pbom.(*testspb.PointersStruct)
		}
	}
	msg = pbo
	return
}
func (goo EmbeddedSt2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt2)
	msg = pbo
	return
}
func (goo *EmbeddedSt2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt2 = msg.(*testspb.EmbeddedSt2)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					err = (*goo).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					err = (*goo).ArraysStruct.FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = (*goo).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					err = (*goo).PointersStruct.FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt2) GetTypeURL() (typeURL string) {
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
func (goo EmbeddedSt3) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt3
	{
		if IsEmbeddedSt3ReprEmpty(goo) {
			var pbov *testspb.EmbeddedSt3
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt3)
		{
			if goo.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
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
			if goo.ArraysStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ArraysStruct.ToPBMessage(cdc)
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
			if goo.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
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
			if goo.PointersStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PointersStruct.ToPBMessage(cdc)
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
			if goo.EmptyStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.EmptyStruct.ToPBMessage(cdc)
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
func (goo EmbeddedSt3) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt3)
	msg = pbo
	return
}
func (goo *EmbeddedSt3) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt3 = msg.(*testspb.EmbeddedSt3)
	{
		if pbo != nil {
			{
				if pbo.PrimitivesStruct != nil {
					(*goo).PrimitivesStruct = new(PrimitivesStruct)
					err = (*goo).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.ArraysStruct != nil {
					(*goo).ArraysStruct = new(ArraysStruct)
					err = (*goo).ArraysStruct.FromPBMessage(cdc, pbo.ArraysStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.SlicesStruct != nil {
					(*goo).SlicesStruct = new(SlicesStruct)
					err = (*goo).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.PointersStruct != nil {
					(*goo).PointersStruct = new(PointersStruct)
					err = (*goo).PointersStruct.FromPBMessage(cdc, pbo.PointersStruct)
					if err != nil {
						return
					}
				}
			}
			{
				if pbo.EmptyStruct != nil {
					(*goo).EmptyStruct = new(EmptyStruct)
					err = (*goo).EmptyStruct.FromPBMessage(cdc, pbo.EmptyStruct)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ EmbeddedSt3) GetTypeURL() (typeURL string) {
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
func (goo EmbeddedSt4) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt4
	{
		if IsEmbeddedSt4ReprEmpty(goo) {
			var pbov *testspb.EmbeddedSt4
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt4)
		{
			pbo.Foo1 = int64(goo.Foo1)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PrimitivesStruct = pbom.(*testspb.PrimitivesStruct)
		}
		{
			pbo.Foo2 = string(goo.Foo2)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.ArraysStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.ArraysStructField = pbom.(*testspb.ArraysStruct)
		}
		{
			goorl := len(goo.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Foo3[i]
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
			pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.SlicesStruct = pbom.(*testspb.SlicesStruct)
		}
		{
			pbo.Foo4 = bool(goo.Foo4)
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.PointersStructField.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.PointersStructField = pbom.(*testspb.PointersStruct)
		}
		{
			pbo.Foo5 = uint64(goo.Foo5)
		}
	}
	msg = pbo
	return
}
func (goo EmbeddedSt4) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt4)
	msg = pbo
	return
}
func (goo *EmbeddedSt4) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt4 = msg.(*testspb.EmbeddedSt4)
	{
		if pbo != nil {
			{
				(*goo).Foo1 = int(int(pbo.Foo1))
			}
			{
				if pbo.PrimitivesStruct != nil {
					err = (*goo).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo2 = string(pbo.Foo2)
			}
			{
				if pbo.ArraysStructField != nil {
					err = (*goo).ArraysStructField.FromPBMessage(cdc, pbo.ArraysStructField)
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
					(*goo).Foo3 = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Foo3 = goors
				}
			}
			{
				if pbo.SlicesStruct != nil {
					err = (*goo).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo4 = bool(pbo.Foo4)
			}
			{
				if pbo.PointersStructField != nil {
					err = (*goo).PointersStructField.FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo5 = uint(uint(pbo.Foo5))
			}
		}
	}
	return
}
func (_ EmbeddedSt4) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt4"
}
func IsEmbeddedSt4ReprEmpty(goor EmbeddedSt4) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != 0 {
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
			if goor.Foo2 != "" {
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
			if goor.Foo4 != false {
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
			if goor.Foo5 != 0 {
				return false
			}
		}
	}
	return
}
func (goo EmbeddedSt5) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.EmbeddedSt5NameOverride
	{
		if IsEmbeddedSt5NameOverrideReprEmpty(goo) {
			var pbov *testspb.EmbeddedSt5NameOverride
			msg = pbov
			return
		}
		pbo = new(testspb.EmbeddedSt5NameOverride)
		{
			pbo.Foo1 = int64(goo.Foo1)
		}
		{
			if goo.PrimitivesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PrimitivesStruct.ToPBMessage(cdc)
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
			pbo.Foo2 = string(goo.Foo2)
		}
		{
			if goo.ArraysStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.ArraysStructField.ToPBMessage(cdc)
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
			goorl := len(goo.Foo3)
			if goorl == 0 {
				pbo.Foo3 = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Foo3[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Foo3 = pbos
			}
		}
		{
			if goo.SlicesStruct != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.SlicesStruct.ToPBMessage(cdc)
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
			pbo.Foo4 = bool(goo.Foo4)
		}
		{
			if goo.PointersStructField != nil {
				pbom := proto.Message(nil)
				pbom, err = goo.PointersStructField.ToPBMessage(cdc)
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
			pbo.Foo5 = uint64(goo.Foo5)
		}
	}
	msg = pbo
	return
}
func (goo EmbeddedSt5) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.EmbeddedSt5NameOverride)
	msg = pbo
	return
}
func (goo *EmbeddedSt5) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.EmbeddedSt5NameOverride = msg.(*testspb.EmbeddedSt5NameOverride)
	{
		if pbo != nil {
			{
				(*goo).Foo1 = int(int(pbo.Foo1))
			}
			{
				if pbo.PrimitivesStruct != nil {
					(*goo).PrimitivesStruct = new(PrimitivesStruct)
					err = (*goo).PrimitivesStruct.FromPBMessage(cdc, pbo.PrimitivesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo2 = string(pbo.Foo2)
			}
			{
				if pbo.ArraysStructField != nil {
					(*goo).ArraysStructField = new(ArraysStruct)
					err = (*goo).ArraysStructField.FromPBMessage(cdc, pbo.ArraysStructField)
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
					(*goo).Foo3 = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Foo3[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Foo3 = goors
				}
			}
			{
				if pbo.SlicesStruct != nil {
					(*goo).SlicesStruct = new(SlicesStruct)
					err = (*goo).SlicesStruct.FromPBMessage(cdc, pbo.SlicesStruct)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo4 = bool(pbo.Foo4)
			}
			{
				if pbo.PointersStructField != nil {
					(*goo).PointersStructField = new(PointersStruct)
					err = (*goo).PointersStructField.FromPBMessage(cdc, pbo.PointersStructField)
					if err != nil {
						return
					}
				}
			}
			{
				(*goo).Foo5 = uint(uint(pbo.Foo5))
			}
		}
	}
	return
}
func (_ EmbeddedSt5) GetTypeURL() (typeURL string) {
	return "/tests.EmbeddedSt5NameOverride"
}
func IsEmbeddedSt5NameOverrideReprEmpty(goor EmbeddedSt5) (empty bool) {
	{
		empty = true
		{
			if goor.Foo1 != 0 {
				return false
			}
		}
		{
			if goor.PrimitivesStruct != nil {
				return false
			}
		}
		{
			if goor.Foo2 != "" {
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
			if goor.Foo4 != false {
				return false
			}
		}
		{
			if goor.PointersStructField != nil {
				return false
			}
		}
		{
			if goor.Foo5 != 0 {
				return false
			}
		}
	}
	return
}
func (goo AminoMarshalerStruct1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct1
	{
		goor, err1 := goo.MarshalAmino()
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
func (goo AminoMarshalerStruct1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct1)
	msg = pbo
	return
}
func (goo *AminoMarshalerStruct1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
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
			err = goo.UnmarshalAmino(goor)
			if err != nil {
				return
			}
		}
	}
	return
}
func (_ AminoMarshalerStruct1) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct1"
}
func IsAminoMarshalerStruct1ReprEmpty(goor ReprStruct1) (empty bool) {
	{
		empty = true
		{
			if goor.C != 0 {
				return false
			}
		}
		{
			if goor.D != 0 {
				return false
			}
		}
	}
	return
}
func (goo ReprStruct1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprStruct1
	{
		if IsReprStruct1ReprEmpty(goo) {
			var pbov *testspb.ReprStruct1
			msg = pbov
			return
		}
		pbo = new(testspb.ReprStruct1)
		{
			pbo.C = int64(goo.C)
		}
		{
			pbo.D = int64(goo.D)
		}
	}
	msg = pbo
	return
}
func (goo ReprStruct1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprStruct1)
	msg = pbo
	return
}
func (goo *ReprStruct1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprStruct1 = msg.(*testspb.ReprStruct1)
	{
		if pbo != nil {
			{
				(*goo).C = int64(pbo.C)
			}
			{
				(*goo).D = int64(pbo.D)
			}
		}
	}
	return
}
func (_ ReprStruct1) GetTypeURL() (typeURL string) {
	return "/tests.ReprStruct1"
}
func IsReprStruct1ReprEmpty(goor ReprStruct1) (empty bool) {
	{
		empty = true
		{
			if goor.C != 0 {
				return false
			}
		}
		{
			if goor.D != 0 {
				return false
			}
		}
	}
	return
}
func (goo AminoMarshalerStruct2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct2
	{
		goor, err1 := goo.MarshalAmino()
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
			var pbos = make([]*testspb.ReprElem2, goorl)
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
func (goo AminoMarshalerStruct2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct2)
	msg = pbo
	return
}
func (goo *AminoMarshalerStruct2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
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
			var goors = make([]ReprElem2, pbol)
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
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ AminoMarshalerStruct2) GetTypeURL() (typeURL string) {
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
func (goo ReprElem2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprElem2
	{
		if IsReprElem2ReprEmpty(goo) {
			var pbov *testspb.ReprElem2
			msg = pbov
			return
		}
		pbo = new(testspb.ReprElem2)
		{
			pbo.Key = string(goo.Key)
		}
		{
			if goo.Value != nil {
				typeUrl := cdc.GetTypeURL(goo.Value)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.Value)
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
func (goo ReprElem2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprElem2)
	msg = pbo
	return
}
func (goo *ReprElem2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprElem2 = msg.(*testspb.ReprElem2)
	{
		if pbo != nil {
			{
				(*goo).Key = string(pbo.Key)
			}
			{
				typeUrl := pbo.Value.TypeUrl
				bz := pbo.Value.Value
				goorp := &(*goo).Value
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ ReprElem2) GetTypeURL() (typeURL string) {
	return "/tests.ReprElem2"
}
func IsReprElem2ReprEmpty(goor ReprElem2) (empty bool) {
	{
		empty = true
		{
			if goor.Key != "" {
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
func (goo AminoMarshalerStruct3) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct3
	{
		goor, err1 := goo.MarshalAmino()
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
func (goo AminoMarshalerStruct3) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct3)
	msg = pbo
	return
}
func (goo *AminoMarshalerStruct3) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerStruct3 = msg.(*testspb.AminoMarshalerStruct3)
	{
		var goor int32
		goor = int32(pbo.Value)
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ AminoMarshalerStruct3) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerStruct3"
}
func IsAminoMarshalerStruct3ReprEmpty(goor int32) (empty bool) {
	{
		empty = true
		if goor != 0 {
			return false
		}
	}
	return
}
func (goo AminoMarshalerInt4) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerInt4
	{
		goor, err1 := goo.MarshalAmino()
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
func (goo AminoMarshalerInt4) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerInt4)
	msg = pbo
	return
}
func (goo *AminoMarshalerInt4) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerInt4 = msg.(*testspb.AminoMarshalerInt4)
	{
		if pbo != nil {
			var goor ReprStruct4
			{
				goor.A = int32(pbo.A)
			}
			err = goo.UnmarshalAmino(goor)
			if err != nil {
				return
			}
		}
	}
	return
}
func (_ AminoMarshalerInt4) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerInt4"
}
func IsAminoMarshalerInt4ReprEmpty(goor ReprStruct4) (empty bool) {
	{
		empty = true
		{
			if goor.A != 0 {
				return false
			}
		}
	}
	return
}
func (goo AminoMarshalerInt5) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerInt5
	{
		goor, err1 := goo.MarshalAmino()
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
func (goo AminoMarshalerInt5) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerInt5)
	msg = pbo
	return
}
func (goo *AminoMarshalerInt5) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.AminoMarshalerInt5 = msg.(*testspb.AminoMarshalerInt5)
	{
		var goor string
		goor = string(pbo.Value)
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ AminoMarshalerInt5) GetTypeURL() (typeURL string) {
	return "/tests.AminoMarshalerInt5"
}
func IsAminoMarshalerInt5ReprEmpty(goor string) (empty bool) {
	{
		empty = true
		if goor != "" {
			return false
		}
	}
	return
}
func (goo AminoMarshalerStruct6) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct6
	{
		goor, err1 := goo.MarshalAmino()
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
			var pbos = make([]*testspb.AminoMarshalerStruct1, goorl)
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
func (goo AminoMarshalerStruct6) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct6)
	msg = pbo
	return
}
func (goo *AminoMarshalerStruct6) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
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
			var goors = make([]AminoMarshalerStruct1, pbol)
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
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ AminoMarshalerStruct6) GetTypeURL() (typeURL string) {
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
func (goo AminoMarshalerStruct7) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.AminoMarshalerStruct7
	{
		goor, err1 := goo.MarshalAmino()
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
			var pbos = make([]uint8, goorl)
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
func (goo AminoMarshalerStruct7) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.AminoMarshalerStruct7)
	msg = pbo
	return
}
func (goo *AminoMarshalerStruct7) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
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
			var goors = make([]ReprElem7, pbol)
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
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ AminoMarshalerStruct7) GetTypeURL() (typeURL string) {
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
func (goo ReprElem7) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ReprElem7
	{
		goor, err1 := goo.MarshalAmino()
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
func (goo ReprElem7) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ReprElem7)
	msg = pbo
	return
}
func (goo *ReprElem7) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ReprElem7 = msg.(*testspb.ReprElem7)
	{
		var goor uint8
		goor = uint8(uint8(pbo.Value))
		err = goo.UnmarshalAmino(goor)
		if err != nil {
			return
		}
	}
	return
}
func (_ ReprElem7) GetTypeURL() (typeURL string) {
	return "/tests.ReprElem7"
}
func IsReprElem7ReprEmpty(goor uint8) (empty bool) {
	{
		empty = true
		if goor != 0 {
			return false
		}
	}
	return
}
func (goo IntDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntDef
	{
		if IsIntDefReprEmpty(goo) {
			var pbov *testspb.IntDef
			msg = pbov
			return
		}
		pbo = &testspb.IntDef{Value: int64(goo)}
	}
	msg = pbo
	return
}
func (goo IntDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntDef)
	msg = pbo
	return
}
func (goo *IntDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntDef = msg.(*testspb.IntDef)
	{
		*goo = IntDef(int(pbo.Value))
	}
	return
}
func (_ IntDef) GetTypeURL() (typeURL string) {
	return "/tests.IntDef"
}
func IsIntDefReprEmpty(goor IntDef) (empty bool) {
	{
		empty = true
		if goor != 0 {
			return false
		}
	}
	return
}
func (goo IntAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntAr
	{
		if IsIntArReprEmpty(goo) {
			var pbov *testspb.IntAr
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]int64, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
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
func (goo IntAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntAr)
	msg = pbo
	return
}
func (goo *IntAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntAr = msg.(*testspb.IntAr)
	{
		var goors = [4]int{}
		for i := 0; i < 4; i += 1 {
			{
				pboe := pbo.Value[i]
				{
					pboev := pboe
					goors[i] = int(int(pboev))
				}
			}
		}
		*goo = goors
	}
	return
}
func (_ IntAr) GetTypeURL() (typeURL string) {
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
func (goo IntSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.IntSl
	{
		if IsIntSlReprEmpty(goo) {
			var pbov *testspb.IntSl
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]int64, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
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
func (goo IntSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.IntSl)
	msg = pbo
	return
}
func (goo *IntSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.IntSl = msg.(*testspb.IntSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*goo = nil
		} else {
			var goors = make([]int, pbol)
			for i := 0; i < pbol; i += 1 {
				{
					pboe := pbo.Value[i]
					{
						pboev := pboe
						goors[i] = int(int(pboev))
					}
				}
			}
			*goo = goors
		}
	}
	return
}
func (_ IntSl) GetTypeURL() (typeURL string) {
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
func (goo ByteAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ByteAr
	{
		if IsByteArReprEmpty(goo) {
			var pbov *testspb.ByteAr
			msg = pbov
			return
		}
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
			pbo = &testspb.ByteAr{Value: pbos}
		}
	}
	msg = pbo
	return
}
func (goo ByteAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ByteAr)
	msg = pbo
	return
}
func (goo *ByteAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ByteAr = msg.(*testspb.ByteAr)
	{
		var goors = [4]uint8{}
		for i := 0; i < 4; i += 1 {
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
func (_ ByteAr) GetTypeURL() (typeURL string) {
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
func (goo ByteSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ByteSl
	{
		if IsByteSlReprEmpty(goo) {
			var pbov *testspb.ByteSl
			msg = pbov
			return
		}
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
			pbo = &testspb.ByteSl{Value: pbos}
		}
	}
	msg = pbo
	return
}
func (goo ByteSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ByteSl)
	msg = pbo
	return
}
func (goo *ByteSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ByteSl = msg.(*testspb.ByteSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*goo = nil
		} else {
			var goors = make([]uint8, pbol)
			for i := 0; i < pbol; i += 1 {
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
	}
	return
}
func (_ ByteSl) GetTypeURL() (typeURL string) {
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
func (goo PrimitivesStructDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructDef
	{
		if IsPrimitivesStructDefReprEmpty(goo) {
			var pbov *testspb.PrimitivesStructDef
			msg = pbov
			return
		}
		pbo = new(testspb.PrimitivesStructDef)
		{
			pbo.Int8 = int32(goo.Int8)
		}
		{
			pbo.Int16 = int32(goo.Int16)
		}
		{
			pbo.Int32 = int32(goo.Int32)
		}
		{
			pbo.Int32Fixed = int32(goo.Int32Fixed)
		}
		{
			pbo.Int64 = int64(goo.Int64)
		}
		{
			pbo.Int64Fixed = int64(goo.Int64Fixed)
		}
		{
			pbo.Int = int64(goo.Int)
		}
		{
			pbo.Byte = uint32(goo.Byte)
		}
		{
			pbo.Uint8 = uint32(goo.Uint8)
		}
		{
			pbo.Uint16 = uint32(goo.Uint16)
		}
		{
			pbo.Uint32 = uint32(goo.Uint32)
		}
		{
			pbo.Uint32Fixed = uint32(goo.Uint32Fixed)
		}
		{
			pbo.Uint64 = uint64(goo.Uint64)
		}
		{
			pbo.Uint64Fixed = uint64(goo.Uint64Fixed)
		}
		{
			pbo.Uint = uint64(goo.Uint)
		}
		{
			pbo.Str = string(goo.Str)
		}
		{
			goorl := len(goo.Bytes)
			if goorl == 0 {
				pbo.Bytes = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Bytes[i]
						{
							pbos[i] = byte(goore)
						}
					}
				}
				pbo.Bytes = pbos
			}
		}
		{
			if !amino.IsEmptyTime(goo.Time) {
				pbo.Time = timestamppb.New(goo.Time)
			}
		}
		{
			if goo.Duration.Nanoseconds() != 0 {
				pbo.Duration = durationpb.New(goo.Duration)
			}
		}
		{
			pbom := proto.Message(nil)
			pbom, err = goo.Empty.ToPBMessage(cdc)
			if err != nil {
				return
			}
			pbo.Empty = pbom.(*testspb.EmptyStruct)
		}
	}
	msg = pbo
	return
}
func (goo PrimitivesStructDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructDef)
	msg = pbo
	return
}
func (goo *PrimitivesStructDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructDef = msg.(*testspb.PrimitivesStructDef)
	{
		if pbo != nil {
			{
				(*goo).Int8 = int8(int8(pbo.Int8))
			}
			{
				(*goo).Int16 = int16(int16(pbo.Int16))
			}
			{
				(*goo).Int32 = int32(pbo.Int32)
			}
			{
				(*goo).Int32Fixed = int32(pbo.Int32Fixed)
			}
			{
				(*goo).Int64 = int64(pbo.Int64)
			}
			{
				(*goo).Int64Fixed = int64(pbo.Int64Fixed)
			}
			{
				(*goo).Int = int(int(pbo.Int))
			}
			{
				(*goo).Byte = uint8(uint8(pbo.Byte))
			}
			{
				(*goo).Uint8 = uint8(uint8(pbo.Uint8))
			}
			{
				(*goo).Uint16 = uint16(uint16(pbo.Uint16))
			}
			{
				(*goo).Uint32 = uint32(pbo.Uint32)
			}
			{
				(*goo).Uint32Fixed = uint32(pbo.Uint32Fixed)
			}
			{
				(*goo).Uint64 = uint64(pbo.Uint64)
			}
			{
				(*goo).Uint64Fixed = uint64(pbo.Uint64Fixed)
			}
			{
				(*goo).Uint = uint(uint(pbo.Uint))
			}
			{
				(*goo).Str = string(pbo.Str)
			}
			{
				var pbol int = 0
				if pbo.Bytes != nil {
					pbol = len(pbo.Bytes)
				}
				if pbol == 0 {
					(*goo).Bytes = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Bytes[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Bytes = goors
				}
			}
			{
				(*goo).Time = pbo.Time.AsTime()
			}
			{
				(*goo).Duration = pbo.Duration.AsDuration()
			}
			{
				if pbo.Empty != nil {
					err = (*goo).Empty.FromPBMessage(cdc, pbo.Empty)
					if err != nil {
						return
					}
				}
			}
		}
	}
	return
}
func (_ PrimitivesStructDef) GetTypeURL() (typeURL string) {
	return "/tests.PrimitivesStructDef"
}
func IsPrimitivesStructDefReprEmpty(goor PrimitivesStructDef) (empty bool) {
	{
		empty = true
		{
			if goor.Int8 != 0 {
				return false
			}
		}
		{
			if goor.Int16 != 0 {
				return false
			}
		}
		{
			if goor.Int32 != 0 {
				return false
			}
		}
		{
			if goor.Int32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int64 != 0 {
				return false
			}
		}
		{
			if goor.Int64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Int != 0 {
				return false
			}
		}
		{
			if goor.Byte != 0 {
				return false
			}
		}
		{
			if goor.Uint8 != 0 {
				return false
			}
		}
		{
			if goor.Uint16 != 0 {
				return false
			}
		}
		{
			if goor.Uint32 != 0 {
				return false
			}
		}
		{
			if goor.Uint32Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint64 != 0 {
				return false
			}
		}
		{
			if goor.Uint64Fixed != 0 {
				return false
			}
		}
		{
			if goor.Uint != 0 {
				return false
			}
		}
		{
			if goor.Str != "" {
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
			if goor.Duration != 0 {
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
func (goo PrimitivesStructSl) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructSl
	{
		if IsPrimitivesStructSlReprEmpty(goo) {
			var pbov *testspb.PrimitivesStructSl
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]*testspb.PrimitivesStruct, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
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
func (goo PrimitivesStructSl) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructSl)
	msg = pbo
	return
}
func (goo *PrimitivesStructSl) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructSl = msg.(*testspb.PrimitivesStructSl)
	{
		var pbol int = 0
		if pbo != nil {
			pbol = len(pbo.Value)
		}
		if pbol == 0 {
			*goo = nil
		} else {
			var goors = make([]PrimitivesStruct, pbol)
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
			*goo = goors
		}
	}
	return
}
func (_ PrimitivesStructSl) GetTypeURL() (typeURL string) {
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
func (goo PrimitivesStructAr) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.PrimitivesStructAr
	{
		if IsPrimitivesStructArReprEmpty(goo) {
			var pbov *testspb.PrimitivesStructAr
			msg = pbov
			return
		}
		goorl := len(goo)
		if goorl == 0 {
			pbo = nil
		} else {
			var pbos = make([]*testspb.PrimitivesStruct, goorl)
			for i := 0; i < goorl; i += 1 {
				{
					goore := goo[i]
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
func (goo PrimitivesStructAr) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.PrimitivesStructAr)
	msg = pbo
	return
}
func (goo *PrimitivesStructAr) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.PrimitivesStructAr = msg.(*testspb.PrimitivesStructAr)
	{
		var goors = [2]PrimitivesStruct{}
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
		*goo = goors
	}
	return
}
func (_ PrimitivesStructAr) GetTypeURL() (typeURL string) {
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
func (goo Concrete1) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete1
	{
		if IsConcrete1ReprEmpty(goo) {
			var pbov *testspb.Concrete1
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete1)
	}
	msg = pbo
	return
}
func (goo Concrete1) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.Concrete1)
	msg = pbo
	return
}
func (goo *Concrete1) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete1 = msg.(*testspb.Concrete1)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ Concrete1) GetTypeURL() (typeURL string) {
	return "/tests.Concrete1"
}
func IsConcrete1ReprEmpty(goor Concrete1) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo Concrete2) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.Concrete2
	{
		if IsConcrete2ReprEmpty(goo) {
			var pbov *testspb.Concrete2
			msg = pbov
			return
		}
		pbo = new(testspb.Concrete2)
	}
	msg = pbo
	return
}
func (goo Concrete2) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.Concrete2)
	msg = pbo
	return
}
func (goo *Concrete2) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.Concrete2 = msg.(*testspb.Concrete2)
	{
		if pbo != nil {
		}
	}
	return
}
func (_ Concrete2) GetTypeURL() (typeURL string) {
	return "/tests.Concrete2"
}
func IsConcrete2ReprEmpty(goor Concrete2) (empty bool) {
	{
		empty = true
	}
	return
}
func (goo ConcreteTypeDef) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ConcreteTypeDef
	{
		if IsConcreteTypeDefReprEmpty(goo) {
			var pbov *testspb.ConcreteTypeDef
			msg = pbov
			return
		}
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
			pbo = &testspb.ConcreteTypeDef{Value: pbos}
		}
	}
	msg = pbo
	return
}
func (goo ConcreteTypeDef) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ConcreteTypeDef)
	msg = pbo
	return
}
func (goo *ConcreteTypeDef) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ConcreteTypeDef = msg.(*testspb.ConcreteTypeDef)
	{
		var goors = [4]uint8{}
		for i := 0; i < 4; i += 1 {
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
func (_ ConcreteTypeDef) GetTypeURL() (typeURL string) {
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
func (goo ConcreteWrappedBytes) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.ConcreteWrappedBytes
	{
		if IsConcreteWrappedBytesReprEmpty(goo) {
			var pbov *testspb.ConcreteWrappedBytes
			msg = pbov
			return
		}
		pbo = new(testspb.ConcreteWrappedBytes)
		{
			goorl := len(goo.Value)
			if goorl == 0 {
				pbo.Value = nil
			} else {
				var pbos = make([]uint8, goorl)
				for i := 0; i < goorl; i += 1 {
					{
						goore := goo.Value[i]
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
func (goo ConcreteWrappedBytes) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.ConcreteWrappedBytes)
	msg = pbo
	return
}
func (goo *ConcreteWrappedBytes) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.ConcreteWrappedBytes = msg.(*testspb.ConcreteWrappedBytes)
	{
		if pbo != nil {
			{
				var pbol int = 0
				if pbo.Value != nil {
					pbol = len(pbo.Value)
				}
				if pbol == 0 {
					(*goo).Value = nil
				} else {
					var goors = make([]uint8, pbol)
					for i := 0; i < pbol; i += 1 {
						{
							pboe := pbo.Value[i]
							{
								pboev := pboe
								goors[i] = uint8(uint8(pboev))
							}
						}
					}
					(*goo).Value = goors
				}
			}
		}
	}
	return
}
func (_ ConcreteWrappedBytes) GetTypeURL() (typeURL string) {
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
func (goo InterfaceFieldsStruct) ToPBMessage(cdc *amino.Codec) (msg proto.Message, err error) {
	var pbo *testspb.InterfaceFieldsStruct
	{
		if IsInterfaceFieldsStructReprEmpty(goo) {
			var pbov *testspb.InterfaceFieldsStruct
			msg = pbov
			return
		}
		pbo = new(testspb.InterfaceFieldsStruct)
		{
			if goo.F1 != nil {
				typeUrl := cdc.GetTypeURL(goo.F1)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.F1)
				if err != nil {
					return
				}
				pbo.F1 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F2 != nil {
				typeUrl := cdc.GetTypeURL(goo.F2)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.F2)
				if err != nil {
					return
				}
				pbo.F2 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F3 != nil {
				typeUrl := cdc.GetTypeURL(goo.F3)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.F3)
				if err != nil {
					return
				}
				pbo.F3 = &anypb.Any{TypeUrl: typeUrl, Value: bz}
			}
		}
		{
			if goo.F4 != nil {
				typeUrl := cdc.GetTypeURL(goo.F4)
				bz := []byte(nil)
				bz, err = cdc.Marshal(goo.F4)
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
func (goo InterfaceFieldsStruct) EmptyPBMessage(cdc *amino.Codec) (msg proto.Message) {
	pbo := new(testspb.InterfaceFieldsStruct)
	msg = pbo
	return
}
func (goo *InterfaceFieldsStruct) FromPBMessage(cdc *amino.Codec, msg proto.Message) (err error) {
	var pbo *testspb.InterfaceFieldsStruct = msg.(*testspb.InterfaceFieldsStruct)
	{
		if pbo != nil {
			{
				typeUrl := pbo.F1.TypeUrl
				bz := pbo.F1.Value
				goorp := &(*goo).F1
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F2.TypeUrl
				bz := pbo.F2.Value
				goorp := &(*goo).F2
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F3.TypeUrl
				bz := pbo.F3.Value
				goorp := &(*goo).F3
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
			{
				typeUrl := pbo.F4.TypeUrl
				bz := pbo.F4.Value
				goorp := &(*goo).F4
				err = cdc.UnmarshalAny2(typeUrl, bz, goorp)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
func (_ InterfaceFieldsStruct) GetTypeURL() (typeURL string) {
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
