Context:

    mix of numeric types

    comparable

    special case



    

Flow:

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


NOTE: 
    
    // in 13f1, indicates that, in preCheck, 
    // if dt is not match with op
    // else if LHS,RHS is typed and mismatched, panic
    // else, check untyped(const), unnamed(composite) cases
    *** is straight forward, if we have right and left types, use it. how about interface and declared types? if LHS and RHS in one of this? so not only untyped passthrough, these latter two needs passthrouh from preCheck too. ***
   
    // wired that == accepts LHS impl RHS or visa versa, while += not accpet this
    

    // only change for binary unary ,assign ,etc,  op related, using another func like checkOperand, else checkOrConvert remains 