package radar

import (
	"net/http"
	"net/url"
	"testing"
)

func Test_isBinaryResource(t *testing.T) {
	testcases := []struct {
		expected bool
		resp     *http.Response
		u        *url.URL
	}{
		{true, &http.Response{}, &url.URL{Path: "/hi.pdf"}},
		{false, &http.Response{}, &url.URL{Path: "/hi.html"}},
		{false, &http.Response{Header: http.Header{"Content-Type": []string{"text/html"}}}, &url.URL{}},
		{true, &http.Response{Header: http.Header{"Content-Type": []string{"text/plain"}}}, &url.URL{}},
	}
	for _, testcase := range testcases {
		actual := isBinaryResource(testcase.resp, testcase.u)
		if actual != testcase.expected {
			t.Fatalf("expected isBinaryResource to return %t, got %t\nresp: %#v\nurl: %#v", testcase.expected, actual, testcase.resp, testcase.u)
		}
	}
}
