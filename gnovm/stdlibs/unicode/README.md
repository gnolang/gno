Directly copied from go1.27rc2 (Unicode 17.0.0). BSD license.
TestCalibrate() commented out for import reasons.

`tables.gno`, `casetables.gno`, `letter.gno`, `graphic.gno` and `digit.gno` are
verbatim copies of the upstream files. When refreshing them, copy
`src/strconv/isprint.go` in the same change: it is generated from the same
Unicode tables and will otherwise disagree with `unicode` about which code
points are printable.
