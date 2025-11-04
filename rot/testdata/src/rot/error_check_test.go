package rot

import "context"

// Simulating require.NoError pattern
type testingT struct{}

func (t *testingT) NoError(err error) {
	if err != nil {
		panic(err)
	}
}

func serviceWithRequirePattern() {
	var t testingT
	client := &Client{}
	svc, err := NewService(client)
	t.NoError(err)
	_, err = svc.Validate(context.Background(), "maybe@example.com")
	t.NoError(err)
}

