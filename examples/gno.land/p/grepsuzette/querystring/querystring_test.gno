package querystring

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestToMap(t *testing.T) {
	{
		m, e := ToMap("a=b&=&&&c=d&e")
		{
			a, has := m["a"]
			uassert.True(t, has)
			uassert.Equal(t, a, "b")
		}
		{
			c, has := m["c"]
			uassert.True(t, has)
			uassert.Equal(t, c, "d")
		}
		{
			e, has := m["e"]
			uassert.True(t, has)
			uassert.Equal(t, e, "")
		}
	}
	{
		m, e := ToMap("a=b&=&&?c=d&e&a=f")
		{
			a, has := m["a"]
			uassert.True(t, has)
			uassert.Equal(t, a, "f")
		}
		{
			c, has := m["c"]
			uassert.True(t, has)
			uassert.Equal(t, c, "d")
		}
		{
			e, has := m["e"]
			uassert.True(t, has)
			uassert.Equal(t, e, "")
		}
	}
	{
		// only the first question mark is considered
		m, e := ToMap("a=b&=&&?c=d&e&a=f???")
		{
			a, has := m["a"]
			uassert.True(t, has)
			uassert.Equal(t, a, "f???")
		}
		{
			c, has := m["c"]
			uassert.True(t, has)
			uassert.Equal(t, c, "d")
		}
		{
			e, has := m["e"]
			uassert.True(t, has)
			uassert.Equal(t, e, "")
		}
	}
	{
		m, e := ToMap("?a=b&a=c=d")
		{
			a, has := m["a"]
			uassert.True(t, has)
			uassert.Equal(t, a, "c=d")
		}
	}
}
