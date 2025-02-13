// Package params provides a lightweight implementation inspired by the x/params
// module of the Cosmos SDK.
//
// It includes a keeper for managing key-value pairs with module identifiers as
// prefixes, along with a global querier for retrieving any key from any module.
//
// The Params Module provides functionalities for caching and persistent access
// to parameters across the entire chain.
//
// It manages both module parameters and arbitrary parameters.
// Module parameters are sourced from all other keepers, such as AuthKeeper,
// BankKeeper, and VMKeeper.Each keeper registers its keeper keys with ParamKeeper
// using the ParamfulKeeper interface.WillSetParam() is called whenever module
// parameters need to be updated.
//
//
// NOTE: important to not use local cached fields unless they are synchronously
// stored to the underlying store. This optimization generally only belongs in paramk.GetParams().
// users of paramk.GetParams() generally should not cache anything and
// instead rely on the efficiency of paramk.GetParams().
//
// In otherwords, ParamKeeper is the only component responsible for caching and storing parameters.
// Other keepers should neither cache nor maintain these parameters as state variables.
// While store access is synchronized, keeper access is not.
//
// Here is the interval key specs that is stored

// /pv/<module>:(<realm>".")?<key>
//
//
// ParamKeeper.SetParams(module_prefix, k, v) is used by each registered module keeper to set the
// module parameter struct.
// ParamKeeper.SetParamXXX() is used to set arbitrary parameters as single primitive values.
//
// A prefix, ValueStoreKeyPrefix (/pv/), is added to each key before it is stored as the internal key:
// Module parameter keys follow this format:
// /pv/<module>:<key>

// For arbitrary parameter keys that set from realm.
// /pv/vm:<realm>.<key>

// The method for querying parameters follows this pattern:
// To query module parameters:
// gnokey query params/<module_prefix>:<params_key>
//
// Since a module parameter is a struct, a simple key "p" is used by the each module
// For example, to query the Auth module's parameters:
// gnokey query params/auth:p
//
// To query arbitrary parameters:
// gnokey query params/<params_key>
//
// For example:
// gnokey query params/gno.land/r/myrealm.foo

// XXX: removes isAlphaNum validation for keys.
// (isAlphaNum for request router validation not sure we want to change it)
package params
