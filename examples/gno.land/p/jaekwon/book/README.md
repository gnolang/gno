Design choices:
 * Action must be in string serializable form for human review.
 * Object references in args should be disallowed for simplicity;
 an Action, like an HTTP Request, is discrete structure.
 * Is "unmarshalling" opinionated? No, let people choose encoding.

Secure Gno:
 * An unexposed (lowercase) declaration can still be used by anyone who holds it.
 * Unexposed fields of any struct can still be copied by assignment.
 * You can also copy an unexposed struct's unexposed field and get a reference.
   `x := external.MakePrivateStructPtr(); y := *x; z := &y`
 * You can tell whether a reference was copied or not by checking the value of
   a private field that was originally set to reference.
   `x := &unexposedStruct{ptr:nil}; x.ptr = x`
 * You can prevent the aforementioned by only returning interface values, and
   generally preventing the holding of an unexposed declaration, but this also
   depends on whether reflection supports instantiation. (VERIFY)
