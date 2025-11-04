package rot

import (
	"net/http"
	"net/url"
	"testing"
)

func BenchmarkPaginationExample(b *testing.B) {
	req := &http.Request{
		URL: &url.URL{RawQuery: "page=5"},
	}

	b.Loop()
	useRequest(req)
}

func TestHelperGuard(t *testing.T) {
	message := "hello"
	t.Helper()
	consume(message)
}

func useRequest(*http.Request) {}

func consume(...any) {}
