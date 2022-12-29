package globaltemplates

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gopk/templates/v2"
	"github.com/gopk/templates/v2/testtemplates"
)

func TestTemplate(t *testing.T) {
	a404 := false

	t.Run("init", func(t *testing.T) {
		assert.NotNil(t, Set("", "tpl.html", true))
		assert.NotNil(t, SetFS(testtemplates.Content, "", "tpl.html", true))
	})

	t.Run("register/funcs", func(t *testing.T) {
		assert.NotNil(t, Func("f1", func() string { return "f1" }))
		assert.NotNil(t, Funcs(template.FuncMap{"f2": func() string { return "f2" }}))
		assert.Equal(t, 2, len(FuncList()))
	})

	t.Run("register/handlers", func(t *testing.T) {
		assert.NotNil(t, RegisterHandler(http.StatusNotFound, func(r *templates.HTTPResponse) error { a404 = true; return nil }))
	})

	t.Run("template", func(t *testing.T) {
		tpl, err := Template("index")
		assert.NoError(t, err)
		assert.NotNil(t, tpl)
	})

	t.Run("render", func(t *testing.T) {
		buff := &bytes.Buffer{}
		assert.NoError(t, Render(buff, templates.Params{"p1": 1}, "index"))
		assert.NotEmpty(t, buff.String())
	})

	t.Run("response", func(t *testing.T) {
		resp := templates.Response(http.StatusNotFound, "index", templates.Params{})
		resp.Writer = &httptest.ResponseRecorder{}
		resp.Request = httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, RenderResponse(resp))
		assert.Equal(t, true, a404)
	})

	t.Run("httpHandler", func(t *testing.T) {
		a404 = false
		fnk := HTTPHandler(func(w http.ResponseWriter, r *http.Request) *templates.HTTPResponse {
			return templates.Response(http.StatusNotFound, "index", templates.Params{})
		})
		fnk(&httptest.ResponseRecorder{}, httptest.NewRequest("GET", "/", nil))
		assert.Equal(t, true, a404)
	})
}
