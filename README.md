# Templates GO wrapper

[![Build Status](https://github.com/gopk/templates/workflows/run%20tests/badge.svg)](https://github.com/gopk/templates/actions?workflow=run%20tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/gopk/templates)](https://goreportcard.com/report/github.com/gopk/templates)
[![GoDoc](https://godoc.org/github.com/gopk/templates?status.svg)](https://godoc.org/github.com/gopk/templates)
[![Coverage Status](https://coveralls.io/repos/github/gopk/templates/badge.svg?branch=master)](https://coveralls.io/github/gopk/templates?branch=master)

> MIT License

Project represents simple wrapper of standart GO templates to make it simpler.

## Example

```go
import (
  temp "github.com/gopk/templates/v2"
)

func main() {
  render := temp.NewHTMLFS(fs.Context, "", ".html", isNotDebug())

  mux := http.NewServeMux()
  mux.HandleFunc("/", render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *temp.HTTPResponse {
    return temp.Response(http.StatusOK, "index", params)
  })
}
```