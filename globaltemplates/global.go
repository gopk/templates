//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015, 2022
//

package globaltemplates

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gopk/templates/v2"
)

var (
	Global *templates.Renderer
)

// G returns glonal renderer
func G() *templates.Renderer { return Global }

// Set renderer base configuration, it creates new or reconfigure the setting
func Set(path, postfix string, enabledCache bool) *templates.Renderer {
	return SetFS(nil, path, postfix, enabledCache)
}

// SetFS renderer base configuration, it creates new or reconfigure the setting with FS dirtectory
func SetFS(fs fs.FS, path, postfix string, enabledCache bool) *templates.Renderer {
	if Global == nil {
		Global = templates.NewFS(fs, path, postfix, enabledCache)
	} else {
		if postfix != "" {
			postfix = "." + strings.TrimLeft(postfix, ".")
		}
		Global.SourceFS = fs
		Global.Path = path
		Global.Postfix = postfix
		Global.CacheEnabled = enabledCache
	}
	return Global
}

// Func register one function in global renderer
func Func(key string, value any) *templates.Renderer {
	return G().Func(key, value)
}

// Funcs register functions in template Renderer
func Funcs(funcs template.FuncMap) *templates.Renderer {
	return G().Funcs(funcs)
}

// FuncList returns the list of prepared template function
func FuncList() template.FuncMap {
	return G().FuncList()
}

// SetDelims of the template
func SetDelims(start, end string) *templates.Renderer {
	return G().SetDelims(start, end)
}

// RegisterHandler for global renderer
func RegisterHandler(code int, handler templates.ResponseHandler) *templates.Renderer {
	return G().RegisterHandler(code, handler)
}

// Template parse and return new template object with all related sub templates
func Template(templates ...string) (*template.Template, error) {
	return G().Template(templates...)
}

// Render template with global renderer
func Render(w io.Writer, params templates.Params, templates ...string) error {
	return G().Render(w, params, templates...)
}

// RenderResponse of HTTP with global renderer
func RenderResponse(resp *templates.HTTPResponse) error {
	return G().RenderResponse(resp)
}

// HTTPHandler wraps http handler as render function
//
// Example:
// mux := http.NewServeMux()
// mux.HandleFunc("/", render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *HTTPResponse { return Response(http.StatusOK, "index", params) })
// mux.HandleFunc("/hello", render.HTTPHandler(getHello))
func HTTPHandler(f templates.HTTPResponseHandler) http.HandlerFunc {
	return G().HTTPHandler(f)
}

// ResetCache of templates
func ResetCache() {
	G().ResetCache()
}
