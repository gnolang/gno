package fp

import (
	"testing"
)

func TestMap(t *testing.T) {
	values := []interface{}{1, 2, 3}
	result := Map(values, func(v interface{}) interface{} { return v.(int) * 2 })
	expected := []interface{}{2, 4, 6}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Map failed, expected %v, got %v", expected, result)
		}
	}
}

func TestFilter(t *testing.T) {
	values := []interface{}{1, 2, 3, 4}
	result := Filter(values, func(v interface{}) bool { return v.(int)%2 == 0 })
	expected := []interface{}{2, 4}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Filter failed, expected %v, got %v", expected, result)
		}
	}
}

func TestReduce(t *testing.T) {
	values := []interface{}{1, 2, 3}
	result := Reduce(values, func(a, b interface{}) interface{} { return a.(int) + b.(int) }, 0)
	expected := 6

	if result != expected {
		t.Errorf("Reduce failed, expected %v, got %v", expected, result)
	}
}

func TestFlatMap(t *testing.T) {
	words := []interface{}{"go", "fn"}
	toChars := func(word interface{}) interface{} {
		chars := []interface{}{}
		for _, c := range word.(string) {
			chars = append(chars, string(c))
		}
		return chars
	}
	result := FlatMap(words, toChars)
	expected := []interface{}{"g", "o", "f", "n"}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("FlatMap failed, expected %v, got %v", expected, result)
		}
	}
}

func TestAllAnyNone(t *testing.T) {
	values := []interface{}{2, 4, 6, 8}
	isEven := func(x interface{}) bool { return x.(int)%2 == 0 }

	if !All(values, isEven) {
		t.Errorf("All failed, expected true")
	}
	if Any(values, func(x interface{}) bool { return x.(int)%2 != 0 }) {
		t.Errorf("Any failed, expected false")
	}
	if !None(values, func(x interface{}) bool { return x.(int)%2 != 0 }) {
		t.Errorf("None failed, expected true")
	}
}

func TestChunk(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	result := Chunk(values, 2)
	expected := [][]interface{}{{1, 2}, {3, 4}, {5}}

	if len(result) != len(expected) {
		t.Errorf("Chunk failed, expected length %v, got %v", len(expected), len(result))
	}
	for i, chunk := range result {
		for j, v := range chunk {
			if v != expected[i][j] {
				t.Errorf("Chunk failed, expected %v, got %v", expected, result)
			}
		}
	}
}

func TestFind(t *testing.T) {
	values := []interface{}{1, 2, 3, 4}
	result, found := Find(values, func(v interface{}) bool { return v.(int) > 2 })

	if !found {
		t.Error("Find failed, expected to find a value")
	}
	if result != 3 {
		t.Errorf("Find failed, expected 3, got %v", result)
	}

	_, found = Find(values, func(v interface{}) bool { return v.(int) > 10 })
	if found {
		t.Error("Find failed, expected not to find a value")
	}
}

func TestReverse(t *testing.T) {
	values := []interface{}{1, 2, 3, 4}
	result := Reverse(values)
	expected := []interface{}{4, 3, 2, 1}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Reverse failed, expected %v, got %v", expected, result)
		}
	}
}

func TestZipUnzip(t *testing.T) {
	a := []interface{}{1, 2, 3}
	b := []interface{}{"a", "b", "c"}

	zipped := Zip(a, b)
	expected := [][2]interface{}{
		{1, "a"},
		{2, "b"},
		{3, "c"},
	}

	for i, pair := range zipped {
		if pair[0] != expected[i][0] || pair[1] != expected[i][1] {
			t.Errorf("Zip failed, expected %v, got %v", expected, zipped)
		}
	}

	unzippedA, unzippedB := Unzip(zipped)
	for i := range a {
		if unzippedA[i] != a[i] || unzippedB[i] != b[i] {
			t.Errorf("Unzip failed, expected %v and %v, got %v and %v", a, b, unzippedA, unzippedB)
		}
	}
}

func TestGroupBy(t *testing.T) {
	values := []interface{}{1, 2, 3, 4}
	result := GroupBy(values, func(v interface{}) interface{} { return v.(int) % 2 })

	if len(result[0]) != 2 || len(result[1]) != 2 {
		t.Errorf("GroupBy failed, expected 2 elements in each group, got %v", result)
	}

	if !contains(result[0], 2) || !contains(result[0], 4) {
		t.Errorf("GroupBy failed, even group should contain 2 and 4")
	}
	if !contains(result[1], 1) || !contains(result[1], 3) {
		t.Errorf("GroupBy failed, odd group should contain 1 and 3")
	}
}

func TestFlatten(t *testing.T) {
	values := [][]interface{}{{1, 2}, {3, 4}, {5}}
	result := Flatten(values)
	expected := []interface{}{1, 2, 3, 4, 5}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Flatten failed, expected %v, got %v", expected, result)
		}
	}
}

// Helper function for testing
func contains(slice []interface{}, item interface{}) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
