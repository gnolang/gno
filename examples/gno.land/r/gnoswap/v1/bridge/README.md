# Upgradable

## Visual Representation of Upgrade Pattern

```plain
r/emission <------ BRD <----- V1 Register
                   ^ 
                   |------- V2 
                   |         |
                   |         | Update
                   |         v
                   |------- V3
                   |        ...
```

This diagram illustrates how the Real-world (`r/emission`) contract uses the bridge (`BRD`) to upgrade through versions (`V1`, `V2`, `V3`) sequentially.

## Overview

The bridge package implements an upgradable pattern. This package stores function names and function pointers as keys and values in an AVL tree, and is used by packages expected to be upgraded by calling the stored function pointers using the key.

## Upgrade Pattern

Upgrades are performed by storing and using function pointers. This approach was adopted for the following reasons:

- Once a contract is deployed, it cannot be directly modified afterward
- Functions are dependent on specific package paths
- Must be able to cross realm-boundaries

For example, let's consider a situation where a function is upgraded from `v1` to `v2`.

```go
// PATH: gno.land/r/demo/v1/calculator
package calculator

func Add(a, b int) int {
 return a + b
}
```

This is a simple addition function. Although unlikely, if a situation arises where the addition logic needs to be modified, and the logic needs to be updated to multiply the result by `2` in `v2`:

```go
// PATH: gno.land/r/demo/v2/calculator
package calculator

func Add(a, b int) int {
 return (a + b) * 2 // <- updated
}
```

If we first use the method of redeploying the contract, we would face the inconvenient situation of having to stop the chain, redeploy, and then manually migrate the data.

Additionally, even if a redeployment is made, if contracts using the `Add` function do not update their package paths, they would continue to use the original `v1.Add`, which could lead to unexpected issues. Of course, there could be a method of not versioning the package path, but in this case, an `gnokey` error saying "package already exists" would occur during deployment, so versioning had to be specified.

Anyway, for these reasons, when considering contract upgrades, the dependency on package paths also had to be considered. That's why we adopted the current approach of storing function names and function pointers as key-value pairs in an `avl` data structure.

The reason for using function pointers was due to issues related to gno's realm-boundary. Currently, there was an issue where some values could pass during realm transitions while others could not. However, functions are treated as first-class citizens and can be used across realm-boundaries. This is due to gno's internal implementation.

When gno calls a function, it handles realm transitions through the `Machine`'s `PushFrameCall` method. At this time, the current realm is stored as `LastRealm`, and it transitions to the realm of the package to which the function belongs. In this process, function pointers are stored in `PackageValue`'s `FBlocks`, so the state of the function is preserved even if realm transitions occur.

```go
// gnovm/pkg/gnolang/machine.go
func (m *Machine) PushFrameCall(cx *CallExpr, fv *FuncValue, recv TypedValue) {
 fr := &Frame{
     // ...
     LastPackage: m.Package,
     LastRealm:   m.Realm,
 }
 // ...
 pv := fv.GetPackage(m.Store)
 rlm := pv.GetRealm()
 if rlm != nil && m.Realm != rlm {
     m.Realm = rlm // enter new realm
 }
}
```

```go
// gnovm/pkg/gnolang/values.go
type PackageValue struct {
 ObjectInfo
// …
 FBlocks []Value
 Realm   *Realm
 fBlocksMap map[Name]*Block
}
```

Additionally, gno's Store type enables object sharing between realms. Objects have unique IDs that can be accessed from any realm, and mappings between realm information and packages are managed through the Store. Due to this mechanism, function pointers can be used even across realm-boundaries.

```go
// gnovm/pkg/gnolang/store.go
type Store interface {
 GetObject(oid ObjectID) Object
 SetObject(Object)
 GetPackageRealm(pkgPath string) *Realm
 SetPackageRealm(*Realm)
 // ...
}
```

Specifically, the following process occurs when a function is called:

1. Look up the function in the realm where the function pointer is stored
2. Transition to the realm of that function through `PushFrameCall`
3. Return to the original realm after executing the function

Thanks to this structure, the bridge package can store function pointers and call them without issues from other realms. This makes it possible to replace existing functions with new ones during contract upgrades, while also solving the package path dependency problem.

## Function Registration and Management

The bridge package provides two main functions for registering and updating functions:

### RegisterCallback: Register a new function

```go
func RegisterCallback(caller Address, namespace, name string, callback any) error
```

- `caller`: Address of the entity registering the function (requires admin, governance authority)
- `namespace`: Namespace to which the function will belong
- `name`: Name of the function
- `callback`: Function pointer to register

Example:

```go
adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
err := bridge.RegisterCallback(adminAddr, "calculator", "Add", Add)
```

### UpdateCallback: Update an existing function

```go
func UpdateCallback(caller Address, namespace, name string, callback any) error
```

- `caller`: Address of the entity updating the function (requires admin, governance authority)
- `namespace`: Namespace of the function to update
- `name`: Name of the function to update
- `callback`: New function pointer

Example:

```go
adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
err := bridge.UpdateCallback(adminAddr, "calculator", "Add", NewAdd)
```

### LookupCallback: Retrieve a stored function

```go
func LookupCallback(namespace, name string) (any, bool)
```

- `namespace`: Namespace of the function to retrieve
- `name`: Name of the function to retrieve
- Return value: Function pointer and existence flag

Example:

```go
cb, exists := bridge.LookupCallback("calculator", "Add")
if !exists {
 return "Add function not found"
}

addFn, ok := cb.(func(int, int) int)
if !ok {
 return "Invalid function type"
}

result := addFn(10, 20)
```

## Usage Examples

1. Initial function registration (v1):

```go
package calculator

func init() {
 adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
 err := bridge.RegisterCallback(adminAddr, "calculator", "Add", Add)
 if err != nil {
     panic(err)
 }
}

func Add(a, b int) int {
 return a + b
}
```

2. Function upgrade (v2):

```go
package calculator

func init() {
 adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
 err := bridge.UpdateCallback(adminAddr, "calculator", "Add", Add)
 if err != nil {
     panic(err)
 }
}

func Add(a, b int) int {
 return (a + b) * 2
}
```

3. Using the function from another contract:

```go
package otherpkg

func UseCalculator(a, b int) int {
 cb, exists := bridge.LookupCallback("calculator", "Add")
 if !exists {
     panic("Add function not found")
 }
    
 addFn, ok := cb.(func(int, int) int)
 if !ok {
     panic("Invalid function type")
 }
    
 return addFn(a, b)
}
```

By registering and updating functions in this way, other contracts can use the latest version of the function without needing to change the package path. Additionally, the bridge package prevents unauthorized modifications through admin privilege checks. Furthermore, if someone tries to make unauthorized modifications using gnokey for injection, currently gnokey parameters can only handle primitive types like numbers or strings, so it cannot handle sophisticated types like function pointers.

## Limitations

The bridge package implements an upgrade pattern using function pointers, but there are several key limitations:

1. Need to predict function registration

- Functions that can be upgraded must be predicted and designed in advance
- If an unpredicted function needs to be upgraded, a new bridge implementation is needed
- Managing all functions through bridge creates significant overhead

Example:

```go
// Only register functions expected to need upgrades to bridge
func init() {
 adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
 bridge.RegisterCallback(adminAddr, "calculator", "Add", Add)
 // Other functions like Sub, Mul, Div might also need upgrades
 // but if not predicted in advance, additional work will be needed later
}
```

2. Reduced readability

- Increased code complexity due to function pointer usage
- Code becomes verbose due to required type assertions
- Difficult to intuitively understand the function call flow

Example:

```go
// Normal function call
result := calculator.Add(10, 20)

// Function call through bridge
cb, exists := bridge.LookupCallback("calculator", "Add")
if !exists {
 panic("Add function not found")
}
addFn, ok := cb.(func(int, int) int)
if !ok {
 panic("Invalid function type")
}
result := addFn(10, 20)
```

3. Type safety constraints

- Function pointers require runtime type checking
- Possible panic if incorrect type assertion is made
- Partially lose the benefits of compile-time type checking

4. Performance overhead

- Additional operations required for function pointer lookup and type assertion
- Indirect calls through bridge for each invocation
- Need to store additional function information
