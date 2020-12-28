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

func Test_isPrivateHost(t *testing.T) {
	testcases := map[string]bool{
		"localhost":       true,
		"127.0.0.1":       true,  // IPv4 loopback
		"10.0.0.1":        true,  // RFC1918
		"172.16.0.1":      true,  // RFC1918
		"192.168.0.1":     true,  // RFC1918
		"169.254.0.1":     true,  // RFC3927 link-local
		"::1":             true,  // IPv6 loopback
		"fe80::":          true,  // IPv6 link-local
		"fc00::":          true,  // IPv6 unique local addr
		"129.129.129.129": false, // public
		"0.0.0.0":         false,
	}
	for input, expected := range testcases {
		actual := isPrivateHost(input)
		if actual != expected {
			t.Fatalf("expected %q to yield %t, got %t", input, expected, actual)
		}
	}
}

func Test_titleForWebpage(t *testing.T) {
	testcases := map[string]string{
		"http://localhost/foo/bar":        "A private page on localhost",
		"http://server.local/foo/bar":     "A private page on server.local",
		"https://myapp.localhost/foo/bar": "A private page on myapp.localhost",
		"http://127.0.0.1:1234/foo/bar":   "A private page on 127.0.0.1",
		"https://example.com":             "Example Domain",
	}
	for input, expected := range testcases {
		actual := titleForWebpage(input)
		if actual != expected {
			t.Fatalf("expected %q to yield %q, got %q", input, expected, actual)
		}
	}
}
