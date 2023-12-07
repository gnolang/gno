Context:

The problem is from []...


`package main

import (
"errors"
"strconv"
)

type Error int64

func (e Error) Error() string {
return "error: " + strconv.Itoa(int(e))
}

var errCmp = errors.New("XXXX")

// specil case:
// one is interface
func main() {
if Error(1) == errCmp {
println("what the firetruck?")
} else {
println("something else")
}
}`



first,why this compiles? the reason for this is Error(1) satisfies interface of error, which indicates Error(1) can be assigned to errCmp, Error(1) and  errCmp is comparable.
"lhs is assignable to rhs, or vice versa", according to spec.

But it gives out incorrect result. in the code above, it should give out :// something else
but gives out: what the firetruck?

The cause for this is about type check, in the case, the Error(1) and errCmp is both mistakenly treated as int64, which
is the underlying type if Error(1), the value of LHS after evaluation is 1 and the RHS is 0, so the == check will give false. 
as a simple prove, if you check this: Error(0) = errCmp, the result will be true.

In the right way, the LHS and RHS has different underlying type, so the result should be false.

It's a corner of the iceberg after some more digging:

Type mix check missing 

`// both typed(different) const
func main() {
println(int(1) == int8(1))
}

in this case, it should not compile for the mismatch of type, but it works(as unexpected).
the reason for this is the missing of a regular type check, the type is cast forcibly while it should not。


Operators check missing or incorrect
`package main

// one untyped const, one typed const
func main() {
println(1 / "a")
}

// Error:
// main/files/types/4b2_filetest.gno:5: operator / not defined on: <untyped> string`
it should give out this error, but gives ...

The reason for this is in golang,  binary expression, unary expression, and INC/DEC stmt are close related to operators. e.g. for ADD, a + b, a and b must be both numericOrString,
   for Sub, a - b, where a and b must be both numeric.

   The current situation is there check around operators happens in runtime, while they should be executed in preprocess stage.
   this non-trivial as for performance, we want these check happens in compile time



3.unamed to named conversion missed

`package main

type word uint
type nat []word

// receiver
func (n nat) add() bool {
return true
}

func Gen() nat {
n := []word{0}
return n
}

// mapLit
func main() {
r := Gen()
switch r.(type) {
case nat:
println("nat")
println(r.add())
default:
println("should not happen")
}
}

// Output:
// nat
// true`

unamed (composite) literals should be converted to named implicitly in preprocess time.
    

Flow:
    checkOp for binary expr, unary expr and inc/dec stmt
        comparable, == !=

        arith + - ...
        isNumericOrString

    checkOperand for special case, in / and %, divisor should not be zero.

    regular type check for const, with nativeType excluded
    regular type check for others, check assignable





    binaryExpression/unaryExpression
    check comparison:
        assignableTo(type), LHS or RHS is assignable to the other 
            isIdentical
                primitive, struct, map, ...
            untyped -> typed
            unnamed -> named
            type(Impl)    -> interface

        EQU NEQ
            LHS, RHS both comparable

        LSS, GTR,,,
            LHS, RHS ordered

        NOTE: above no requre for match type, e,g. Main.Error and error to compare

        // else ops
        first check (if typed) LHS, RHS, panic mismatch directly
        check predicates
        check special case like zero divisor
        check implicit convertable, same logic assignableTo(Type)



        
            
        

Scenarios:
    binary expression:
    left expression [Op] right expression

        // both const
        const [Op] const
            left typed Op right typed
            left untyped Op right typed
            left typed Op right untyped
            left untyped Op right untyped   // println(1.0 * 0)
        // one is const
        Not const [Op] const
            Not const [Op] untyped const
            Not const [Op] typed const
        const [Op] Not const
            untyped const [Op] const
            typed const [Op] const

        // both not const
        Not const [Op] Not const

Untyped Convert rule:

        here describes convert rules on both prime or composite values

        // untyped value can be converted to its correspondence value when:
        // a) it's (untyped) const, and conform to convert rules, 
           e.g. println(int(0) + 1) will work
           e.g. println(int(0) + "a") will not work

        // b) it's declared type of value, like type Error int8, 0 will be able to converted to Error
            e.g. 
            type Error int64
            func main() {
                println(Error(0) == "0") // will work
                println(Error(0) == "a") // not work
                println(Error(0) == []string{"hello"}) // not work
            }
        
        actually, declard type is a special case of convert rule

Special case:

        when one of LHS or RHS is interface, and the other implement the interface, it will comform some 
        `equlaity` rule, which means they are treated as equal in type-check in compile time, but do the
        strict check in runtime. TODO: understand wo golang really do in runtime

        This is pretty similar with the `declared type rule` as above, the both conforms some specific constraint

Operators:
    
    // general rule, arith operators has higher precedence, required `strict` check
    // In contrast, equlity == or != has a lower precedence
    // in case of +=, its precedence is a combination of + and =, it's given + precedence

	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	BAND     // &
	BOR      // |
	XOR      // ^
	SHL      // <<
	SHR      // >>
	BAND_NOT // &^

    LAND  // &&
    LOR   // ||
    ARROW // <-
    INC   // ++
    DEC   // --

    -------------------------------------------------------

	ADD_ASSIGN      // +=
	SUB_ASSIGN      // -=
	MUL_ASSIGN      // *=
	QUO_ASSIGN      // /=
	REM_ASSIGN      // %=
	BAND_ASSIGN     // &=
	BOR_ASSIGN      // |=
	XOR_ASSIGN      // ^=
	SHL_ASSIGN      // <<=
	SHR_ASSIGN      // >>=
	BAND_NOT_ASSIGN // &^=

    -------------------------------------------------------

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ    // !=
	LEQ    // <=
	GEQ    // >=
	DEFINE // :=


others:
    callExpr
        params, return value(multi)
    
    untyped composite value is not const, but a unamed type?

thoughts:
    rule based code gen, Peg?


TODOs: 

    // TODO: dec value representation
    // TODO: Current flow : check op operand,  check type convertable, and convert, and check type match again,  means, this kind of check should still in preprocess
    // TODO: preCheck->Convert->postCheck, all in `checkOrConvertType`


    this is for arith and comparable operators

    specical case: bigInt to gonative time.Month. skipped
    
    mix of numeric types


NOTE: 
    
    // in 13f1, indicates that, in preCheck, 
    // if dt is not match with op
    // else if LHS,RHS is typed and mismatched, panic
    // else, check untyped(const), unnamed(composite) cases
    *** is straight forward, if we have right and left types, use it. how about interface and declared types? if LHS and RHS in one of this? so not only untyped passthrough, these latter two needs passthrouh from preCheck too. ***
   
    // wired that == accepts LHS impl RHS or visa versa, while += not accpet this
    

    // only change for binary unary ,assign ,etc,  op related, using another func like checkOperand, else checkOrConvert remains 