Design choices for Actions
 * Action must be in string serializable form for human review.
 * Object references in args should be disallowed for simplicity;
 an Action, like an HTTP Request, is discrete structure.
 * Is "unmarshalling" opinionated? No, let people choose encoding.

Secure Gno: (move elsewhere)
 1. An unexposed (lowercase) declaration can be used by anyone who holds it.
 1. Unexposed fields of any struct can still be copied by assignment.
 1. You can also copy an unexposed struct's unexposed field and get a
    reference.  `x := external.MakePrivateStructPtr(); y := *x; z := &y`
 1. You could *maybe* prevent the above by only returning interface
    values, and generally preventing the holding of an unexposed declaration,
    but this also depends on whether reflection supports instantiation, and the
    user would still need to check that the type is what they expect it to be.
 1. In other words, don't expect to prevent creation of new references for
    security.
 1. You can tell whether a reference was copied or not by checking the value of
    a private field that was originally set to reference.
    `x := &unexposedStruct{ptr:nil}; x.ptr = x`
