package templates

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"

	"github.com/gopk/templates/v2/testtemplates"
)

func TestTemplate(t *testing.T) {
	a404 := false
	var render *Renderer

	t.Run("init", func(t *testing.T) {
		render = New("", "tpl.html", true)
		assert.NotNil(t, render)
		render = NewFS(testtemplates.Content, "", "tpl.html", true)
		assert.NotNil(t, render)
	})

	t.Run("register/funcs", func(t *testing.T) {
		assert.NotNil(t, render.Func("f1", func() string { return "f1" }))
		assert.NotNil(t, render.Funcs(template.FuncMap{"f2": func() string { return "f2" }}))
	})

	t.Run("register/handlers", func(t *testing.T) {
		assert.NotNil(t, render.RegisterHandler(http.StatusNotFound, func(r *HTTPResponse) error { a404 = true; return nil }))
	})

	t.Run("template", func(t *testing.T) {
		tpl, err := render.Template("index")
		assert.NoError(t, err)
		assert.NotNil(t, tpl)
	})

	t.Run("render", func(t *testing.T) {
		buff := &bytes.Buffer{}
		assert.NoError(t, render.Render(buff, Params{"p1": 1}, "index"))
		assert.NotEmpty(t, buff.String())
	})

	t.Run("response", func(t *testing.T) {
		resp := Response(http.StatusNotFound, "index", Params{})
		resp.Writer = &httptest.ResponseRecorder{}
		resp.Request = httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, render.RenderResponse(resp))
		assert.Equal(t, true, a404)
	})

	t.Run("httpHandler", func(t *testing.T) {
		a404 = false
		fnk := render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *HTTPResponse {
			return Response(http.StatusNotFound, "index", Params{})
		})
		fnk(&httptest.ResponseRecorder{}, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, true, a404)
	})
}
