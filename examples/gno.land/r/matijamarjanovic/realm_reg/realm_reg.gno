package realm_reg

import (
//	"std"
//	"time"
	"gno.land/p/demo/ufmt"
	"strings"
	"gno.land/p/demo/avl"
)

var fileRegistry = avl.NewTree() // filePath -> description
var test = ""

func Register(input string) {

	parts := strings.Split(input, ";") 
	for _, part := range parts {
		kv := strings.SplitN(part, "|", 2) 
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])   
			value := strings.TrimSpace(kv[1]) 
			fileRegistry.Set(key, value)              
		}
	}

}

func Render(path string) string {
	var output strings.Builder 
	output.WriteString(test)
	output.WriteString("\n")

	size := fileRegistry.Size() 

	for i := 0; i < size ; i++ {
		key, value := fileRegistry.GetByIndex(i) 
		output.WriteString(ufmt.Sprintf("- %s | %s\n\n", key, value.(string))) 
	}

	return output.String() 
}
