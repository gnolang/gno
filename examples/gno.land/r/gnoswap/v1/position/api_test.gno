package position

import (
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"
)

func setupPositions(t *testing.T) {
	t.Helper()
	MakeMintPositionWithoutFee(t)
}

func TestApiGetPositions_ResponseStructure(t *testing.T) {
	tests := []struct {
		name         string
		apiCall      func() string
		expectedSize int
		validator    func(*ApiTestHelper, *json.Node)
		setRealm     bool
	}{
		{
			name: "Basic Response",
			apiCall: func() string {
				return ApiGetPositions()
			},
			expectedSize: 1,
			validator: func(helper *ApiTestHelper, position *json.Node) {
				helper.AssertPositionFields(position)
			},
			setRealm: true,
		},
		{
			name: "By Pool Path",
			apiCall: func() string {
				return ApiGetPositionsByPoolPath("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")
			},
			expectedSize: 1,
			validator: func(helper *ApiTestHelper, position *json.Node) {
				helper.AssertPositionFields(position)
			},
			setRealm: false,
		},
		{
			name: "By Address",
			apiCall: func() string {
				return ApiGetPositionsByAddress(adminAddr)
			},
			expectedSize: 1,
			validator: func(helper *ApiTestHelper, position *json.Node) {
				helper.AssertPositionFields(position)
			},
			setRealm: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPositionTest(t)
			setupPositions(t)

			if tt.setRealm {
				testing.SetRealm(adminRealm)
			}

			runApiTest(t, tt.apiCall, tt.expectedSize, tt.validator)
		})
	}
}

func TestApiGetPositionsFee(t *testing.T) {
	tests := []struct {
		name         string
		apiCall      func() string
		expectedSize int
		validator    func(*ApiTestHelper, *json.Node)
	}{
		{
			name: "Unclaimed Fee",
			apiCall: func() string {
				return ApiGetPositionsUnclaimedFee()
			},
			expectedSize: 1,
			validator: func(helper *ApiTestHelper, feeData *json.Node) {
				helper.AssertFeeFields(feeData)
			},
		},
		{
			name: "Unclaimed Fee By LP Position ID",
			apiCall: func() string {
				return ApiGetPositionUnclaimedFeeByLpPositionId(1)
			},
			expectedSize: 1,
			validator: func(helper *ApiTestHelper, feeData *json.Node) {
				helper.AssertFeeFields(feeData)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPositionTest(t)
			setupPositions(t)

			runApiTest(t, tt.apiCall, tt.expectedSize, tt.validator)
		})
	}
}

// runApiTest is a helper function that runs the common logic for API tests.
func runApiTest(
	t *testing.T,
	apiCall func() string,
	expectedSize int,
	validator func(*ApiTestHelper, *json.Node),
) {
	t.Helper()

	result := apiCall()
	helper := NewApiTestHelper(t, result)

	// Test basic API response structure
	helper.TestApiResponse()

	// Verify response array size
	response := helper.GetResponse()
	helper.AssertArraySize(response, expectedSize)

	// Validate first data
	data := helper.GetResponseArray(0)
	validator(helper, data)
}

//////// Helper Functions ////////

// ApiTestHelper is a helper struct for API response testing.
type ApiTestHelper struct {
	t      *testing.T
	result string
	root   *json.Node
}

// NewApiTestHelper creates a new API test helper.
func NewApiTestHelper(t *testing.T, result string) *ApiTestHelper {
	t.Helper()

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		t.Fatalf("JSON parsing failed: %v", err)
	}

	return &ApiTestHelper{
		t:      t,
		result: result,
		root:   root,
	}
}

// GetResponse returns the value of the response key.
func (h *ApiTestHelper) GetResponse() *json.Node {
	h.t.Helper()

	response, err := h.root.GetKey("response")
	if err != nil {
		h.t.Fatalf("response key not found: %v", err)
	}

	return response
}

// GetResponseArray returns the element at the specified index of the response array.
func (h *ApiTestHelper) GetResponseArray(index int) *json.Node {
	h.t.Helper()

	response := h.GetResponse()
	uassert.Equal(h.t, index+1, response.Size())

	element, err := response.GetIndex(index)
	if err != nil {
		h.t.Fatalf("array index %d not found: %v", index, err)
	}

	return element
}

// AssertField validates the value of a specific field.
func (h *ApiTestHelper) AssertField(node *json.Node, fieldName, expectedValue string) {
	h.t.Helper()

	field, err := node.GetKey(fieldName)
	if err != nil {
		h.t.Fatalf("field %s not found: %v", fieldName, err)
	}

	uassert.Equal(h.t, expectedValue, field.String())
}

// AssertFieldExists checks if a specific field exists.
func (h *ApiTestHelper) AssertFieldExists(node *json.Node, fieldName string) {
	h.t.Helper()

	_, err := node.GetKey(fieldName)
	if err != nil {
		h.t.Fatalf("field %s does not exist: %v", fieldName, err)
	}
}

// AssertFieldNotExists checks if a specific field does not exist.
func (h *ApiTestHelper) AssertFieldNotExists(node *json.Node, fieldName string) {
	h.t.Helper()

	_, err := node.GetKey(fieldName)
	if err == nil {
		h.t.Fatalf("field %s exists (should not exist)", fieldName)
	}
}

// AssertArraySize validates the size of an array.
func (h *ApiTestHelper) AssertArraySize(node *json.Node, expectedSize int) {
	h.t.Helper()

	uassert.Equal(h.t, expectedSize, node.Size())
}

// AssertStringField validates a string field.
func (h *ApiTestHelper) AssertStringField(node *json.Node, fieldName, expectedValue string) {
	h.t.Helper()

	field, err := node.GetKey(fieldName)
	if err != nil {
		h.t.Fatalf("field %s not found: %v", fieldName, err)
	}

	value, err := field.GetString()
	if err != nil {
		h.t.Fatalf("field %s is not a string: %v", fieldName, err)
	}

	uassert.Equal(h.t, expectedValue, value)
}

// AssertNumberField validates a numeric field.
func (h *ApiTestHelper) AssertNumberField(node *json.Node, fieldName string, expectedValue float64) {
	h.t.Helper()

	field, err := node.GetKey(fieldName)
	if err != nil {
		h.t.Fatalf("field %s not found: %v", fieldName, err)
	}

	value, err := field.GetNumeric()
	if err != nil {
		h.t.Fatalf("field %s is not a number: %v", fieldName, err)
	}

	uassert.Equal(h.t, expectedValue, value)
}

// AssertBoolField validates a boolean field.
func (h *ApiTestHelper) AssertBoolField(node *json.Node, fieldName string, expectedValue bool) {
	h.t.Helper()

	field, err := node.GetKey(fieldName)
	if err != nil {
		h.t.Fatalf("field %s not found: %v", fieldName, err)
	}

	value, err := field.GetBool()
	if err != nil {
		h.t.Fatalf("field %s is not a boolean: %v", fieldName, err)
	}

	uassert.Equal(h.t, expectedValue, value)
}

var mockPosition = map[string]string{
	"lpPositionId":             "1",
	"burned":                   "false",
	"owner":                    `"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"`,
	"operator":                 `"g100000000000000000000000000000000dnmcnx"`,
	"poolKey":                  `"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500"`,
	"tickLower":                "-887270",
	"tickUpper":                "887270",
	"liquidity":                `"50000"`,
	"feeGrowthInside0LastX128": `"0"`,
	"feeGrowthInside1LastX128": `"0"`,
	"token0Owed":               `"0"`,
	"token1Owed":               `"0"`,
	"token0Balance":            `"50000"`,
	"token1Balance":            `"50000"`,
	"fee0Unclaimed":            `"0"`,
	"fee1Unclaimed":            `"0"`,
}

// AssertPositionFields validates common position-related fields.
func (h *ApiTestHelper) AssertPositionFields(position *json.Node) {
	h.t.Helper()

	for fieldName, expectedValue := range mockPosition {
		h.AssertField(position, fieldName, expectedValue)
	}
}

// AssertFeeFields validates fee-related fields.
func (h *ApiTestHelper) AssertFeeFields(feeNode *json.Node) {
	h.t.Helper()

	expectedFields := map[string]string{
		"lpPositionId": "1",
		"fee0":         `"0"`,
		"fee1":         `"0"`,
	}

	for fieldName, expectedValue := range expectedFields {
		h.AssertField(feeNode, fieldName, expectedValue)
	}
}

// TestApiResponse tests the basic structure of API response.
func (h *ApiTestHelper) TestApiResponse() {
	h.t.Helper()

	// Check if response key exists
	h.AssertFieldExists(h.root, "response")

	// Check if response is an array
	response := h.GetResponse()
	if !response.IsArray() {
		h.t.Fatalf("response is not an array")
	}
}

// CreateExpectedPositionMap creates a map of expected position data.
func CreateExpectedPositionMap() map[string]string {
	return mockPosition
}

// CreateExpectedFeeMap creates a map of expected fee data.
func CreateExpectedFeeMap() map[string]string {
	return map[string]string{
		"lpPositionId": "1",
		"fee0":         "0",
		"fee1":         "0",
	}
}
