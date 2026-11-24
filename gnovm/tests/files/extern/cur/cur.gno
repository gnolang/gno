package cur

// A package NAMED cur. Imported without an alias it binds `cur` in the
// importer's file scope, which zrealm_cur_decl_import_implicit.gno refuses.

type Fake struct{}

func (Fake) IsUser() bool { return true }

func Previous() Fake { return Fake{} }
