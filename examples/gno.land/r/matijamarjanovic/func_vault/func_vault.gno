package func_vault

type FuncRecord struct {
    Name string
    Function interface{}
    Description string
    Creator std.Address
    ID int64
}

func (fr FuncRecord) Less(than Record) bool {
    other, ok := than.(FuncRecord)
    if !ok {
        panic("Cannot compare FuncRecord with non-FuncRecord type")
    }
    return fr.ID < other.ID
}

type FuncVault struct {
    tree *btree.BTree
}

var funcVault = &FuncVault{
    tree: btree.New(),
}
