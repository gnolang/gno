package tokenhub

import (
    "std"
    "errors"
	"strings"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/grc/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/demo/ufmt"
)

var (
    registeredTokens = avl.NewTree() // symbol -> *grc20.Token
    
    ErrTokenAlreadyRegistered = errors.New("token already registered")
    ErrTokenNotFound = errors.New("token not found")
)

func RegisterToken(token *grc20.Token) error {
    symbol := token.GetSymbol()
    
    if registeredTokens.Has(symbol) {
        return ErrTokenAlreadyRegistered
    }
    
    registeredTokens.Set(symbol, token)
    return nil
}

func getBalances(addr std.Address) string {
    balances := avl.NewTree()
    var sb strings.Builder
    
    registeredTokens.Iterate("", "", func(key string, value interface{}) bool {
        token := value.(*grc20.Token)
        balance := token.BalanceOf(addr)
        balances.Set(key, balance)
        return false
    })

    sb.WriteString("Balances: ")
    balances.Iterate("", "", func(key string, value interface{}) bool {
        balance := value.(uint64)
        sb.WriteString(ufmt.Sprintf("%s: %d ", key, balance))
        return false
    })
    
    return sb.String()
}

func GetMyBalances() string {
    return getBalances(std.GetOrigCaller())
}

func Render(path string) string {
    var sb strings.Builder

    sb.WriteString(md.H1("TokenHub"))
    sb.WriteString(md.Paragraph("A central registry for GRC20 tokens on gno.land"))

    sb.WriteString(md.H2("Registered Tokens"))
    
	var tokenItems []string
    registeredTokens.Iterate("", "", func(key string, value interface{}) bool {
        token := value.(*grc20.Token)
        tokenItems = append(tokenItems, ufmt.Sprintf("%s (%s) - %d decimals", 
            token.GetName(),
            token.GetSymbol(), 
            token.GetDecimals()))
        return false
    })

    if len(tokenItems) > 0 {
        sb.WriteString(md.BulletList(tokenItems))
    } else {
        sb.WriteString(md.Italic("No tokens registered yet"))
        sb.WriteString("\n")
    }

    return sb.String()
}

