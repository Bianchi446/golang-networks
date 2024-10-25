package middleware 

import(
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"
)

func TestTimeoutMiddleware(t *testing.T) {
	handler := http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			w.WriteHeader(http.StatusNoContent)
			time.Sleep(time.Minute)
		}), 
		time.Second(),
		"Timed out while reading response", 
	)
	r := httptest.NewRequest(http.MethodGet, "http://test/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("Unexpected status code: %q", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if actual := string(b); actual != "Timed out while reading response" {
		t.Logf("Unexpected body: %q", actual)
	}
}