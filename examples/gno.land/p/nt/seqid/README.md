# seqid

```
package seqid // import "gno.land/p/nt/seqid"

Package seqid provides a simple way to have sequential IDs which will be ordered
correctly when inserted in an AVL tree.

Sample usage:

    var id seqid.ID
    var users avl.Tree

    func NewUser() {
    	users.Set(id.Next().Binary(), &User{ ... })
    }

TYPES

type ID uint64
    An ID is a simple sequential ID generator.

func FromBinary(b string) (ID, bool)
    FromBinary creates a new ID from the given string.

func (i ID) Binary() string
    Binary returns a big-endian binary representation of the ID, suitable to be
    used as an AVL key.

func (i *ID) Next() ID
    Next advances the ID i. It will panic if increasing ID would overflow.

func (i *ID) TryNext() (ID, bool)
    TryNext increases i by 1 and returns its value. It returns true if
    successful, or false if the increment would result in an overflow.
```
