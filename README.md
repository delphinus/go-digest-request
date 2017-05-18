# go-digest-request - request with digest authentication in golang

[![CircleCI](https://circleci.com/gh/delphinus/go-digest-request.svg?style=svg)](https://circleci.com/gh/delphinus/go-digest-request)
[![Coverage Status](https://coveralls.io/repos/github/delphinus/go-digest-request/badge.svg?branch=master)](https://coveralls.io/github/delphinus/go-digest-request?branch=master)

## Usage

* When creating context, use `digestRequest.ContextWithClient()` for `appengine.urlfetch` in Google App Engine.

```go
import (
  "fmt"
  "io/ioutil"
  "net/http"

  "github.com/delphinus/go-digest-request"
  "golang.org/x/net/context"
)

func main() {
  ctx := context.Background()
  r := digestRequest.New(ctx, "john", "hello") // username & password

  req, _ := http.NewRequest("GET", "http://example.com", nil)
  resp, _ := r.Do(req)
  defer resp.Body.Close()

  b, _ := ioutil.ReadAll

  fmt.Println(string(b))
}
```
