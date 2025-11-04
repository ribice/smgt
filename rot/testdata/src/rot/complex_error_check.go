package rot

type testCase struct {
	in      string
	want    string
	wantErr bool
}

type testT struct{}

func (t *testT) Fatalf(format string, args ...interface{}) {}

func testComplexErrorCheck() {
	var tt testCase
	var testT testT
	cleaned := "test"
	got, err := normalizePhone(cleaned)
	if (err != nil) != tt.wantErr {
		testT.Fatalf("NormalizePhone(%q) error = %v; wantErr=%v", tt.in, err, tt.wantErr)
	}
	if !tt.wantErr && got != tt.want {
		testT.Fatalf("NormalizePhone(%q) = %q; want %q", tt.in, got, tt.want)
	}
}

func testErrorCheckWithAnd() {
	var expectedErr error
	input := "test"
	result, err := processData(input)
	if err != nil && err != expectedErr {
		return
	}
	_ = result
}

func testErrorCheckInExpression() {
	value, err := getValue()
	if checkError(err) {
		return
	}
	_ = value
}

// Mock functions
func normalizePhone(cleaned string) (string, error) {
	return "", nil
}

func processData(input string) (string, error) {
	return "", nil
}

func getValue() (string, error) {
	return "", nil
}

func checkError(err error) bool {
	return err != nil
}

