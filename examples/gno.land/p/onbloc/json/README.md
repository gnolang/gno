# JSON Parser

The JSON parser is a package that provides functionality for parsing and processing JSON strings. This package accepts JSON strings as byte slices.

Currently, gno does not [support the `reflect` package](https://docs.gno.land/resources/effective-gno#reflection-is-never-clear), so it cannot retrieve type information at runtime. Therefore, it is designed to infer and handle type information when parsing JSON strings using a state machine approach.

After passing through the state machine, JSON strings are represented as the `Node` type. The `Node` type represents nodes for JSON data, including various types such as `ObjectNode`, `ArrayNode`, `StringNode`, `NumberNode`, `BoolNode`, and `NullNode`.

This package provides methods for manipulating, searching, and extracting the Node type.

## State Machine

To parse JSON strings, a [finite state machine](https://en.wikipedia.org/wiki/Finite-state_machine) approach is used. The state machine transitions to the next state based on the current state and the input character while parsing the JSON string. Through this method, type information can be inferred and processed without reflect, and the amount of parser code can be significantly reduced.

The image below shows the state transitions of the state machine according to the states and input characters.

```mermaid
stateDiagram-v2
    [*] --> __: Start
    __ --> ST: String
    __ --> MI: Number
    __ --> ZE: Zero
    __ --> IN: Integer
    __ --> T1: Boolean (true)
    __ --> F1: Boolean (false)
    __ --> N1: Null
    __ --> ec: Empty Object End
    __ --> cc: Object End
    __ --> bc: Array End
    __ --> co: Object Begin
    __ --> bo: Array Begin
    __ --> cm: Comma
    __ --> cl: Colon
    __ --> OK: Success/End
    ST --> OK: String Complete
    MI --> OK: Number Complete
    ZE --> OK: Zero Complete
    IN --> OK: Integer Complete
    T1 --> OK: True Complete
    F1 --> OK: False Complete
    N1 --> OK: Null Complete
    ec --> OK: Empty Object Complete
    cc --> OK: Object Complete
    bc --> OK: Array Complete
    co --> OB: Inside Object
    bo --> AR: Inside Array
    cm --> KE: Expecting New Key
    cm --> VA: Expecting New Value
    cl --> VA: Expecting Value
    OB --> ST: String in Object (Key)
    OB --> ec: Empty Object
    OB --> cc: End Object
    AR --> ST: String in Array
    AR --> bc: End Array
    KE --> ST: String as Key
    VA --> ST: String as Value
    VA --> MI: Number as Value
    VA --> T1: True as Value
    VA --> F1: False as Value
    VA --> N1: Null as Value
    OK --> [*]: End
```

## Examples

This package provides parsing functionality along with encoding and decoding functionality. The following examples demonstrate how to use this package.

### Decoding

Decoding (or Unmarshaling) is the functionality that converts an input byte slice JSON string into a `Node` type.

The converted `Node` type allows you to modify the JSON data or search and extract data that meets specific conditions.

```go
package main

import (
    "gno.land/p/demo/json"
    "gno.land/p/nt/ufmt/v0"
)

func main() {
    node, err := json.Unmarshal([]byte(`{"foo": "var"}`))
    if err != nil {
        ufmt.Errorf("error: %v", err)
    }

    ufmt.Sprintf("node: %v", node)
}
```

### Encoding

Encoding (or Marshaling) is the functionality that converts JSON data represented as a Node type into a byte slice JSON string.

> ⚠️ Caution: Converting a large `Node` type into a JSON string may _impact performance_. or might be cause _unexpected behavior_.

```go
package main

import (
    "gno.land/p/demo/json"
    "gno.land/p/nt/ufmt/v0"
)

func main() {
    node := ObjectNode("", map[string]*Node{
        "foo": StringNode("foo", "bar"),
        "baz": NumberNode("baz", 100500),
        "qux": NullNode("qux"),
    })

    b, err := json.Marshal(node)
    if err != nil {
        ufmt.Errorf("error: %v", err)
    }

    ufmt.Sprintf("json: %s", string(b))
}
```

### Searching

Once the JSON data converted into a `Node` type, you can **search** and **extract** data that satisfy specific conditions. For example, you can find data with a specific type or data with a specific key.

To use this functionality, you can use methods in the `GetXXX` prefixed methods. The `MustXXX` methods also provide the same functionality as the former methods, but they will **panic** if data doesn't satisfies the condition.

Here is an example of finding data with a specific key. For more examples, please refer to the [node.gno](node.gno) file.

```go
package main

import (
    "gno.land/p/demo/json"
    "gno.land/p/nt/ufmt/v0"
)

func main() {
    root, err := Unmarshal([]byte(`{"foo": true, "bar": null}`))
    if err != nil {
        ufmt.Errorf("error: %v", err)
    }

    value, err := root.GetKey("foo")
    if err != nil {
        ufmt.Errorf("error occurred while getting key, %s", err)
    }

    if value.MustBool() != true {
        ufmt.Errorf("value is not true")
    }

    value, err = root.GetKey("bar")
    if err != nil {
        t.Errorf("error occurred while getting key, %s", err)
    }

    _, err = root.GetKey("baz")
    if err == nil {
        t.Errorf("key baz is not exist. must be failed")
    }
}
```

## Contributing

Please submit any issues or pull requests for this package through the GitHub repository at [gnolang/gno](<https://github.com/gnolang/gno>).
