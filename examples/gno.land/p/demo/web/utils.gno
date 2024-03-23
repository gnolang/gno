package web

import (
	"std"
	"strings"
	"strconv"
	"regexp"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
)

// PUBLIC

//// INT=>INT CONVERSIONS (+ Stringer shorthand helper function)

func Str(str_ interface{}) string {
	return Stringer(str_)
}

func Int(int_ interface{}) int {
	return int(int_)
}

func UI64(int_ int) uint64 {
	return uint64(int_)
}

func I64(int_ int) int64 {
	return int64(int_)
}

func UI32(int_ int) uint32 {
	return uint32(int_)
}

func I32(int_ int) int32 {
	return int32(int_)
}

func UI16(int_ int) uint16 {
	return uint16(int_)
}

func I16(int_ int) int16 {
	return int16(int_)
}

func UI8(int_ int) uint8 {
	return uint8(int_)
}

func I8(int_ int) int8 {
	return int8(int_)
}


//// STR=>INT CONVERSION

func StrInt(str_ string) int {
	_num, _err := strconv.Atoi(str_)
	if _err != nil {
		panic(ufmt.Sprintf("Error converting '%s' to int",str_))
	}
	return _num
}

func StrUI64(str_ string) uint64 {
	return UI64(StrInt(str_))	
}

func StrI64(str_ string) int64 {
	return I64(StrInt(str_))	
}

func StrUI32(str_ string) uint32 {
	return UI32(StrInt(str_))	
}

func StrI32(str_ string) int32 {
	return I32(StrInt(str_))	
}

func StrUI16(str_ string) uint16 {
	return UI16(StrInt(str_))	
}

func StrI16(str_ string) int16 {
	return I16(StrInt(str_))	
}

func StrUI8(str_ string) uint8 {
	return UI8(StrInt(str_))	
}

func StrI8(str_ string) int8 {
	return I8(StrInt(str_))	
}

// INTERNAL

func attify(pattern_ string) string {
	if pattern_ == "" || pattern_ == "/" {
		return "@"
	}
	return pattern_
}

func trimSplit(pattern_ string) []string {
	_path := trimLeadingSlash(pattern_)
	return strings.Split(_path, "/")	
}

func subSplit(pattern_ string) []string {
	return strings.Split(pattern_, ":")	
}

func keyify(key_ string ) string {
	return "{" + key_ + "}"
}

func partKey(part_ string) string {
	_partChars := []rune(part_)
	_first := string(_partChars[0])
	_last := string(_partChars[len(_partChars)-1])
	if _first == "{" && _last == "}" {
		_key := strings.TrimPrefix(part_,"{")
		return strings.TrimSuffix(_key,"}")
	}
	return ""
}

func trimLeadingSlash(path_ string) string {
	_pathChars := []rune(path_)
	_first := string(_pathChars[0])
	if _first != "/" {
		return path_
	}

	for _i := range path_ {
		if _i > 0 {
			return path_[_i:]
		}
	}
	return path_[:0]
}

// TODO: implement?
// _matched, _subKeys, _subValues := subPathResolve(_rtPart, _reqParts[_i])					
// if _matched {
// 	for _j, _subKey := range _subKeys {
// 		_keys = append(_keys, _subKey)
// 		_values.Set(_subKey,_subValues[_j])
// 	}
// 	_matches++
// }

func subPathResolve(routePart_ string, pathPart_ string) (bool, []string, []string) {
	_matched := false 
	_keys := []string{}
	_values := []string{}

	return _matched, _keys, _values
	// TODO: update to support `thing:{id}` and `{id}:thing` style "part" resolution
					// _splitRoutePart := subSplit(_part)
					// _splitPathPart := subSplit(_pathParts[_i])
					// _splitLength := len(_splitRoutePart)
					// _splitPathLength := len(_splitPathPart)
					// if _splitLength == 1 || _splitLength != _splitPathLength {
					// 	continue
					// }
					// _subMatches := 0
					// for _j, _sub := range _splitRoutePart {
					// 	if _sub == _splitPathPart[_j] {
					// 		_subMatches++
					// 		_keys = append(_keys, _sub)
					// 		_values.Set(_sub,_sub)
					// 	} else {
					// 		_sk := partKey(_sub)
					// 		if _sk != "" {
					// 			_subMatches++ 
					// 			_keys = append(_keys, _sk)
					// 			_values.Set(_sk,_splitPathPart[_j])
					// 		}
					// 	}
					// }
					// if _splitLength == _subMatches {
					// 	_matches++
					// } else {
					// 	continue
					// }
}