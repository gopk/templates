package templates

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gopk/templates/v2/testtemplates"
)

func TestTemplate(t *testing.T) {
	a404 := false
	var render *HTMLRender

	t.Run("init", func(t *testing.T) {
		renderPlain := NewPlain("", ".tpl.html", true)
		assert.NotNil(t, renderPlain)
		renderPlain = NewPlainFS(testtemplates.Content, "", ".tpl.html", true)
		assert.NotNil(t, renderPlain)
		render = NewHTML("", ".tpl.html", true)
		assert.NotNil(t, render)
		render = NewHTMLFS(testtemplates.Content, "", ".tpl.html", true)
		assert.NotNil(t, render)
		assert.NotNil(t, render.SetDelims("{{", "}}"))
		render.Params = Params{"global": "param"}
	})

	t.Run("register/funcs", func(t *testing.T) {
		assert.NotNil(t, render.Func("f1", func() string { return "f1" }))
		assert.NotNil(t, render.Funcs(template.FuncMap{"f2": func() string { return "f2" }}))
		assert.Equal(t, 2, len(render.FuncList()))
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
		assert.NoError(t, render.Render(buff, nil, "index"))
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

		var resp1 = httptest.NewRecorder()
		errFnk1 := render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *HTTPResponse {
			return Response(http.StatusInternalServerError, "", nil)
		})
		errFnk1(resp1, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, "Invalid response render", strings.TrimSpace(resp1.Body.String()))

		var resp2 = httptest.NewRecorder()
		errFnk2 := render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *HTTPResponse { return nil })
		errFnk2(resp2, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, "Invalid http response", strings.TrimSpace(resp2.Body.String()))
	})
}
