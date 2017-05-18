package digestRequest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abbot/go-http-auth"
	"golang.org/x/net/context"
)

// startServer is written with referring to
// https://github.com/abbot/go-http-auth/blob/master/examples/digest.go
func startServer(ctx context.Context) *httptest.Server {
	a := auth.NewDigestAuthenticator("example.com", func(user, realm string) string {
		if user == "john" {
			return "b98e16cbc3d01734b264adba7baa3bf9" // password is "hello"
		}
		return ""
	})
	ts := httptest.NewServer(a.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		fmt.Fprintf(w, "OK")
	}))

	go func() {
		<-ctx.Done()
		ts.Close()
	}()

	return ts
}

func TestDigestRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = ContextWithClient(ctx, http.DefaultClient)
	ts := startServer(ctx)

	r := New(ctx, "john", "hello")

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Errorf("error in NewRequest: %v", err)
	}

	resp, err := r.Do(req)
	if err != nil {
		t.Errorf("error in Do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("error status code: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("error in ReadAll: %v", err)
	}

	if string(b) != "OK" {
		t.Errorf("invalid body: %s", string(b))
	}
}
