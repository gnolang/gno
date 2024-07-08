// Copyright 2017 Tendermint. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package amino_test

import (
	"fmt"
	"reflect"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
)

func Example() {
	type Message interface{}

	type bcMessage struct {
		Message string
		Height  int
	}

	type bcResponse struct {
		Status  int
		Message string
	}

	type bcStatus struct {
		Peers int
	}

	// amino.RegisterPackage registers globally.
	amino.RegisterPackage(
		amino.NewPackage(
			reflect.TypeOf(bcMessage{}).PkgPath(),
			"amino_test",
			amino.GetCallersDirname(),
		).
			WithTypes(&bcMessage{}, &bcResponse{}, &bcStatus{}),
	)

	bm := &bcMessage{Message: "ABC", Height: 100}
	msg := bm

	var bz []byte // the marshalled bytes.
	var err error
	bz, err = amino.MarshalAnySized(msg)
	fmt.Printf("Encoded: %X (err: %v)\n", bz, err)

	var msg2 Message
	err = amino.UnmarshalSized(bz, &msg2)
	fmt.Printf("Decoded: %v (err: %v)\n", msg2, err)
	bm2 := msg2.(*bcMessage)
	fmt.Printf("Decoded successfully: %v\n", *bm == *bm2)

	// Output:
	// Encoded: 210A152F616D696E6F5F746573742E62634D65737361676512080A0341424310C801 (err: <nil>)
	// Decoded: &{ABC 100} (err: <nil>)
	// Decoded successfully: true
}
