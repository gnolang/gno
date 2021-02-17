// PKGPATH: gno.land/r/chat
// I am playing.  I am not certain what the result of this will be.
// I think I am starting to get what Jae's been on about the whole time, "but you don't need graphics man"
// "A gameboy would be fine"
// It is very difficult to communicate technical concepts, especially new ones (even if they're old)
// Sometimes you just need to build them to even let people know what you're on about.
// This is going to be fun.

package chat  // This should make me gno.land/r/chat
var root interface{}  //I believe that this gets us talking to the Merkle Trie

type ChatMessage struct {
    OwnerID
    CreateTime
    Text string
    Attachments []Attachment
}


type Attachment struct {
    Type string
    Data []byte
}

func UpdateRoot(...) error {
  root = ...
}



package main

var root Node

type Node interface{}
type Key interface{}
type Value interface{}

type InnerNode struct {
	Key   Key
	Left  Node `gno:owned`
	Right Node `gno:owned`
}

type LeafNode struct {
	Key   Key
	Value Value
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	println("gno.land initializing...")
	node1 := LeafNode{
		Key:   "left",
		Value: "left value",
	}
	node2 := LeafNode{
		Key:   "right",
		Value: "right value",
	}
	root = InnerNode{
		Key:   "new",
		Left:  node1,
		Right: node2,
	}
	println("gno.land ready!")
}

