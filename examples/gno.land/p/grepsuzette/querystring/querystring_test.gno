package querystring

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestToMap(t *testing.T) {
	{
		m, e := ToMap("city=london&=&&&place=bigben&nuggets")
		{
			city, has := m["city"]
			uassert.True(t, has)
			uassert.Equal(t, city, "london")
		}
		{
			place, has := m["place"]
			uassert.True(t, has)
			uassert.Equal(t, place, "bigben")
		}
		{
			nuggets, has := m["nuggets"]
			uassert.True(t, has)
			uassert.Equal(t, nuggets, "")
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
	{
		m, e := ToMap("?abc")
		{
			abc, has := m["abc"]
			uassert.True(t, has)
			uassert.Equal(t, abc, "")
		}
	}
}
