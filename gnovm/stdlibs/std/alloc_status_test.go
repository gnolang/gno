package std

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

func TestGetUsageMem(t *testing.T) {
	m := gno.NewMachine("Alloc", nil)
	testSet := []struct {
		name              string
		maxAlloc          int64
		toBeAlloc         interface{}
		shouldHaveErr     bool
		expectedAllocated int64 // lenCost := _allocBaseSize + DataSize
	}{
		{
			name:              "Test Allocate String",
			maxAlloc:          500 * 1000 * 1000,
			toBeAlloc:         "TestString",
			shouldHaveErr:     false,
			expectedAllocated: 34,
		},
		{
			name:              "Test Allocate Integer",
			maxAlloc:          10 * 1000 * 1000,
			toBeAlloc:         123,
			shouldHaveErr:     false,
			expectedAllocated: 1,
		},
		{
			name:              "Test Allocate DataArray",
			maxAlloc:          500 * 1000 * 1000,
			toBeAlloc:         []int{1, 2, 3, 4, 5, 6, 7},
			shouldHaveErr:     false,
			expectedAllocated: 215,
		},
		//Check if we allocate more than maxSize
		{
			name:              "Test Max Size Allocate Panic",
			maxAlloc:          10,
			toBeAlloc:         "shouldHavePanic",
			shouldHaveErr:     true,
			expectedAllocated: 26,
		},
	}
	for _, tc := range testSet {
		t.Run(tc.name, func(t *testing.T) {
			newAllocator := gno.NewAllocator(tc.maxAlloc)
			m.Alloc = newAllocator
			if tc.shouldHaveErr {
				assert.Panics(t, func() {
					m.Alloc.Allocate(tc.maxAlloc + 1)
				})
			} else {
				switch tc.toBeAlloc.(type) {
				case string:
					m.Alloc.AllocateString(int64(len(tc.toBeAlloc.(string))))
				case int:
					m.Alloc.Allocate(1)
				case []int:
					m.Alloc.AllocateDataArray(int64(len(tc.toBeAlloc.([]int))))
				default:
					//Default ?
				}
				assert.Equal(t, tc.expectedAllocated, GetCurrAllocatedMem(m))
				assert.Equal(t, tc.maxAlloc, GetAllocMaxSize(m))
			}
		})
	}
}

func TestAllocTracer(t *testing.T) {
	
}