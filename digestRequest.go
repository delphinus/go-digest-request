package digestRequest

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/delphinus/random-string"
	"golang.org/x/net/context"
)

type httpClientKey struct{}

// HTTPClientKey will be used for a key of context
var HTTPClientKey httpClientKey

// ContextWithClient returns context with a specified *http.Client
func ContextWithClient(parent context.Context, client *http.Client) context.Context {
	return context.WithValue(parent, HTTPClientKey, client)
}

func clientFromContext(ctx context.Context) *http.Client {
	if client, ok := ctx.Value(HTTPClientKey).(*http.Client); ok {
		return client
	}
	return http.DefaultClient
}

// DigestRequest is a client for digest authentication requests
type DigestRequest struct {
	context.Context
	client             *http.Client
	username, password string
	nonceCount         nonceCount
}

type nonceCount int

func (nc nonceCount) String() string {
	c := int(nc)
	return fmt.Sprintf("%08x", c)
}

const algorithm = "algorithm"
const authorization = "Authorization"
const contentType = "Content-Type"
const nonce = "nonce"
const opaque = "opaque"
const qop = "qop"
const realm = "realm"
const wwwAuthenticate = "Www-Authenticate"

var wanted = []string{algorithm, nonce, opaque, qop, realm}

// New makes a DigestRequest instance
func New(ctx context.Context, username, password string) *DigestRequest {
	return &DigestRequest{
		Context:  ctx,
		client:   clientFromContext(ctx),
		username: username,
		password: password,
	}
}

// Do does requests as http.Do does
func (r *DigestRequest) Do(req *http.Request) (*http.Response, error) {
	parts, err := r.makeParts(req)
	if err != nil {
		return nil, err
	}

	if parts != nil {
		req.Header.Set(authorization, r.makeAuthorization(req, parts))
	}

	return r.client.Do(req)
}

func (r *DigestRequest) makeParts(req *http.Request) (map[string]string, error) {
	authReq, err := http.NewRequest(req.Method, req.URL.String(), nil)
	resp, err := r.client.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}

	if len(resp.Header[wwwAuthenticate]) == 0 {
		return nil, fmt.Errorf("headers do not have %s", wwwAuthenticate)
	}

	headers := strings.Split(resp.Header[wwwAuthenticate][0], ",")
	parts := make(map[string]string, len(wanted))
	for _, r := range headers {
		for _, w := range wanted {
			if strings.Contains(r, w) && strings.Contains(r, `"`) {
				parts[w] = strings.Split(r, `"`)[1]
			} else if strings.Contains(r, w) && strings.Contains(r, "=") {
				parts[w] = strings.Split(r, `=`)[1]
			}
		}
	}

	if len(parts) != len(wanted) {
		return nil, fmt.Errorf("header is invalid: %+v", parts)
	}

	return parts, nil
}

func getMD5(texts []string) string {
	h := md5.New()
	_, _ = io.WriteString(h, strings.Join(texts, ":"))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *DigestRequest) getNonceCount() string {
	r.nonceCount++
	return r.nonceCount.String()
}

func (r *DigestRequest) makeAuthorization(req *http.Request, parts map[string]string) string {
	ha1 := getMD5([]string{r.username, parts[realm], r.password})
	ha2 := getMD5([]string{req.Method, req.URL.String()})
	cnonce := randomString.Generate(16)
	nc := r.getNonceCount()
	response := getMD5([]string{
		ha1,
		parts[nonce],
		nc,
		cnonce,
		parts[qop],
		ha2,
	})
	return fmt.Sprintf(
		`Digest username="%s", realm="%s", nonce="%s", uri="%s", algorithm="%s", qop=%s, nc=%s, cnonce="%s", response="%s", opaque="%s"`,
		r.username,
		parts[realm],
		parts[nonce],
		req.URL.String(),
		parts[algorithm],
		parts[qop],
		nc,
		cnonce,
		response,
		parts[opaque],
	)
}
