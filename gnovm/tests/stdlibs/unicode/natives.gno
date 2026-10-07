package unicode

// Custom code: IsPrint, IsGraphic, SimpleFold and IsUpper used to be declared
// here and bound to the host Go unicode package as a test-time optimization.
// That made them answer according to whatever Unicode version the host
// toolchain shipped, which silently diverged from gno's own tables once the
// two versions differed - so a realm could classify a rune one way under
// `gno test` and the other way on chain. The bindings were removed; tests now
// exercise the same table lookups production does.
