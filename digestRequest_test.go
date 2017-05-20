package digestRequest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abbot/go-http-auth"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func testRequest(h http.HandlerFunc, setClient func(context.Context) context.Context) error {
	ctx := context.Background()

	if setClient != nil {
		ctx = setClient(ctx)
	}

	ts := httptest.NewServer(h)
	defer ts.Close()

	r := New(ctx, "john", "hello")

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		return errors.Wrap(err, "error in NewRequest")
	}

	resp, err := r.Do(req)
	if err != nil {
		return errors.Wrap(err, "error in Do")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("error status code: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error in ReadAll")
	}

	if string(b) != "OK" {
		return errors.Errorf("invalid body: %s", string(b))
	}

	return nil
}

// digestHandler is written with referring to
// https://github.com/abbot/go-http-auth/blob/master/examples/digest.go
var digestHandler = func() http.HandlerFunc {
	a := auth.NewDigestAuthenticator("example.com", func(user, realm string) string {
		if user == "john" {
			return "b98e16cbc3d01734b264adba7baa3bf9" // password is "hello"
		}
		return ""
	})
	return a.Wrap(func(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
		fmt.Fprintf(w, "OK")
	})
}()

func TestDigestRequestWithClient(t *testing.T) {
	err := testRequest(digestHandler, func(ctx context.Context) context.Context {
		return ContextWithClient(ctx, http.DefaultClient)
	})
	if err != nil {
		t.Errorf("error in testRequest: %v", err)
	}
}

func TestDigestRequestWithoutClient(t *testing.T) {
	if err := testRequest(digestHandler, nil); err != nil {
		t.Errorf("error in testRequest: %v", err)
	}
}

func testNormalRequest(writer func(w http.ResponseWriter)) error {
	return testRequest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer(w)
		}),
		nil,
	)
}

func TestNormalRequest(t *testing.T) {
	err := testNormalRequest(func(w http.ResponseWriter) {
		fmt.Fprintf(w, "OK")
	})
	if err != nil {
		t.Errorf("error in testRequest: %v", err)
	}
}

func TestNormalRequestWithUnauthorizedError(t *testing.T) {
	err := testNormalRequest(func(w http.ResponseWriter) {
		http.Error(w, "OK", http.StatusUnauthorized)
	})
	if !strings.Contains(err.Error(), "headers do not have Www-Authenticate") {
		t.Errorf("different error: %v", err)
	}
}

func TestNormalRequestWithInvalidHeaders(t *testing.T) {
	err := testNormalRequest(func(w http.ResponseWriter) {
		w.Header().Set(wwwAuthenticate, "hoge")
		http.Error(w, "OK", http.StatusUnauthorized)
	})
	if !strings.Contains(err.Error(), "header is invalid") {
		t.Errorf("different error: %v", err)
	}
}

func TestInvalidRequests(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil) // invalid request
	if err != nil {
		t.Fatalf("error in NewRequest: %v", err)
	}

	_, err = New(context.Background(), "", "").Do(req)
	if err == nil {
		t.Fatalf("no error")
	}
}
