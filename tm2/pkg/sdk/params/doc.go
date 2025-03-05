// Package params provides a lightweight implementation inspired by the x/params
// module of the Cosmos SDK.
//
// It includes a keeper for managing key-value pairs with module identifiers as
// prefixes, along with a global querier for retrieving any key from any module.
//
// The Params Module provides functionalities for caching and persistent access
// to parameters across the entire chain.
//
// Keys are generally of the format <module>:<submodule>:<name>.
// Parameters are stored in the underlying store with key format
// /pv/<module>:<submodule>:<name>.
//
//  * 'module' must be an alphanumeric ASCII string.
//  * 'submodule' can be anything but cannot contain a ':'
//  * 'submodule' is set to 'p' for keeper param structs.
//  * The VM keeper of gno.land uses the submodule for realm paths.
//
// Module parameters are sourced from all other keepers, such as AuthKeeper,
// BankKeeper, and VMKeeper which must be registered with .Register().
// The ParamsKeeper ensures that the <module> is registered,
// but otherwise doesn't enforce much else about the key format.
//
// Before params are written ParamfulKeeper.WillSetParam() is called, allowing
// any custom caching to happen if needed. Any caching must be stored in 'ctx',
// and keepers must otherwise be immutable otherwise the checktx (mempool) and
// delivertx (blockchain) states would trample each other.
//
// ParamKeeper.SetStruct(module, "p", k, v) is used by each registered module
// keeper to set the module parameter struct (for type-safety).
// ParamKeeper.SetParamXXX() is used to set arbitrary primitive parameters.
//
// The method for querying parameters follows this pattern:
// gnokey query params/<module>:<submodule>:<name>.
// For example, gnokey query params/vm:gno.land/r/myuser/myrealm:foo.

package params
