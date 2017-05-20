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

// startDigestServer is written with referring to
// https://github.com/abbot/go-http-auth/blob/master/examples/digest.go
func startDigestServer(ctx context.Context) *httptest.Server {
	a := auth.NewDigestAuthenticator("example.com", func(user, realm string) string {
		if user == "john" {
			return "b98e16cbc3d01734b264adba7baa3bf9" // password is "hello"
		}
		return ""
	})
	return startServer(ctx, a.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		fmt.Fprintf(w, "OK")
	}))
}

func startNormalServer(ctx context.Context) *httptest.Server {
	return startServer(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	}))
}

func startServer(ctx context.Context, h http.Handler) *httptest.Server {
	ts := httptest.NewServer(h)
	go func() {
		<-ctx.Done()
		ts.Close()
	}()
	return ts
}

type contexter func(context.Context) context.Context
type serverer func(context.Context) *httptest.Server

func testRequest(t *testing.T, server serverer, setClient contexter) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if setClient != nil {
		ctx = setClient(ctx)
	}

	ts := server(ctx)

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

func TestDigestRequestWithClient(t *testing.T) {
	testRequest(t, startDigestServer, func(ctx context.Context) context.Context {
		return ContextWithClient(ctx, http.DefaultClient)
	})
}

func TestDigestRequestWithoutClient(t *testing.T) {
	testRequest(t, startDigestServer, nil)
}

func TestNormalRequest(t *testing.T) {
	testRequest(t, startNormalServer, nil)
}
