package std

import (
	testing "testing/base"
)
func TestDeleteSliceIndex(t *testing.T) {
	data := []string{"someKey","anotherKey"}
	if len(data)!=2{
		t.Errorf("unexpected number of arguments on slice expected %d got %d",2, len(data))
	}
	DeleteSliceIndex(data,0)
	if len(data)!=1{
		t.Errorf("unexpected number of arguments on slice expected %d got %d",1, len(data))
	}
}
