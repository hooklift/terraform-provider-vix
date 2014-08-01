package helper_test

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/c4milo/terraform_vix/helper"
)

func TestExpBackoff(t *testing.T) {
	duration := time.Millisecond
	max := time.Hour
	for i := 0; i < math.MaxUint16; i++ {
		duration = helper.ExpBackoff(duration, max)
		assert(t, duration > 0, fmt.Sprintf("duration too small: %v %v", duration, i))
		assert(t, duration <= max, fmt.Sprintf("duration too large: %v %v", duration, i))
	}
}

// Test exponential backoff and that it continues retrying if a 5xx response is
// received
func TestGetURLExpBackOff(t *testing.T) {
	var expBackoffTests = []struct {
		count int
		body  string
	}{
		{0, "number of attempts: 0"},
		{1, "number of attempts: 1"},
		{2, "number of attempts: 2"},
	}
	client := helper.NewHttpClient()

	for i, tt := range expBackoffTests {
		mux := http.NewServeMux()
		count := 0
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if count == tt.count {
				io.WriteString(w, fmt.Sprintf("number of attempts: %d", count))
				return
			}
			count++
			http.Error(w, "", 500)
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()

		data, err := client.GetRetry(ts.URL)
		ok(t, err)
		assert(t, count == tt.count, fmt.Sprintf("Test case %d failed: %d != %d", i, count, tt.count))
		assert(t, string(data) == tt.body, fmt.Sprintf("Test case %d failed: %s != %s", i, tt.body, data))
	}
}

// Test that it stops retrying if a 4xx response comes back
func TestGetURL4xx(t *testing.T) {
	client := helper.NewHttpClient()
	retries := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retries++
		http.Error(w, "", 404)
	}))
	defer ts.Close()

	_, err := client.GetRetry(ts.URL)
	assert(t, err != nil, fmt.Sprintf("Error was expected not to be nil. Got %v", err))
	equals(t, err.Error(), "Not found. HTTP status code: 404")
	assert(t, retries <= 1, fmt.Sprintf("Number of retries:\n%d\nExpected number of retries:\n%s", retries, 1))
}

// Test that it fetches and returns user-data just fine
func TestGetURL2xx(t *testing.T) {
	var cloudcfg = `
#cloud-config
coreos: 
	oem:
	    id: test
	    name: CoreOS.box for Test
	    version-id: %VERSION_ID%+%BUILD_ID%
	    home-url: https://github.com/coreos/coreos-cloudinit
	    bug-report-url: https://github.com/coreos/coreos-cloudinit
	update:
		reboot-strategy: best-effort
`

	client := helper.NewHttpClient()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, cloudcfg)
	}))
	defer ts.Close()

	data, err := client.GetRetry(ts.URL)
	ok(t, err)
	assert(t, string(data) == cloudcfg, fmt.Sprintf("%s != %s", string(data), cloudcfg))
}

// Test attempt to fetching using malformed URL
func TestGetMalformedURL(t *testing.T) {
	client := helper.NewHttpClient()

	var tests = []struct {
		url  string
		want string
	}{
		{"boo", "URL boo does not have a valid HTTP scheme. Skipping."},
		{"mailto://boo", "URL mailto://boo does not have a valid HTTP scheme. Skipping."},
		{"ftp://boo", "URL ftp://boo does not have a valid HTTP scheme. Skipping."},
		{"", "URL is empty. Skipping."},
	}

	for _, test := range tests {
		_, err := client.GetRetry(test.url)
		assert(t, err != nil, fmt.Sprintf("Error was expected not to be nil. Got %v", err))
		assert(t, err.Error() == test.want, fmt.Sprintf("Incorrect result\ngot:  %v\nwant: %v", err, test.want))
	}
}
