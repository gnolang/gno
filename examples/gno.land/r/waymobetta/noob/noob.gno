package noob

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	noob = avl.NewTree()
	ownerAddr string
)

type Storage struct {
	Date int64 // block height
	Data []string
}

func init() {
	ownerAddr = std.GetOrigCaller().String()

	noob.Set(
		"foo",
		Storage{
			Date: std.GetHeight(),
			Data: []string{"bar"},
		},
	)
}

func Noob(data string) {
	var storage Storage
	var dataSlice []string

	// get existing
	existing, ok := noob.Get("foo")
	if ok {
		storage = existing.(Storage)	

		for _, existingData := range storage.Data {
			dataSlice = append(
				dataSlice,
				existingData,
			)
		}
	}

	// append new data
	dataSlice = append(
		dataSlice,
		data,
	)

	// update with new + old values
	noob.Set(
		"foo",
		Storage{
			Date: std.GetHeight(), // sets current block height as "date"
			Data: dataSlice,
		},
	)	
}

func GetNoob() string {
	var storage Storage
	var resultStr string

	existing, ok := noob.Get("foo")
	if ok {
		storage = existing.(Storage)

		for index, data := range storage.Data {
			// if last item in slice
			if index == len(storage.Data)-1 {
				resultStr += ufmt.Sprintf("%s", data)

				return ufmt.Sprintf(`
				height: [%d] 
				data: [%s]`, 
					storage.Date, 
					resultStr,
				)
			}
			resultStr += ufmt.Sprintf("%s, ", data)
		}
	}
	return ufmt.Sprintf(`
		height: [%d] 
		data: [%s]`, 
		storage.Date, 
		resultStr,
	)
}

func GetOwner() string {
	return ownerAddr
}
