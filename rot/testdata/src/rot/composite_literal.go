package rot

import "time"

type TestStruct struct {
	StringField   string
	IntField      int
	Int64Field    int64
	UintField     uint
	Float64Field  float64
	BoolFieldTrue bool
	BoolFieldFalse bool
	TimeField     time.Time
}

func testFunction() {
	testString := "test_value"
	testInt := 42
	testInt64 := int64(123)
	testBool := true
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	zeroInt64 := int64(0)

	tests := []struct {
		name     string
		input    TestStruct
		expected map[string]any
	}{
		{
			name: "all non-zero simple types",
			input: TestStruct{
				StringField:    testString,
				IntField:       testInt,
				Int64Field:     testInt64,
				UintField:      456,
				Float64Field:   3.14,
				BoolFieldTrue:  testBool,
				BoolFieldFalse: false,
				TimeField:      testTime,
			},
			expected: map[string]any{
				"zero": zeroInt64,
			},
		},
	}

	_ = tests
}

func testSliceLiteral() {
	val1 := 1
	val2 := 2
	val3 := 3

	slice := []int{val1, val2, val3}
	_ = slice
}

func testMapLiteral() {
	key1 := "key1"
	key2 := "key2"
	val1 := "value1"
	val2 := "value2"

	m := map[string]string{
		key1: val1,
		key2: val2,
	}
	_ = m
}

func testArrayLiteral() {
	val1 := 1
	val2 := 2

	arr := [2]int{val1, val2}
	_ = arr
}

func testFunctionArgs() {
	arg1 := "arg1"
	arg2 := "arg2"

	result := processArgs(arg1, arg2)
	_ = result
}

func processArgs(a1, a2 string) string {
	return a1 + a2
}
