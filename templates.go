//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015, 2022
//

package templates

import (
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	templatesRegex = regexp.MustCompile("\\{\\{\\s*template\\s*['\"]([^'\"]+)['\"][^\\}]*\\}\\}")
)

// Params of the renderer
type Params map[string]any

// ResponseHandler for the particular HTTP code
type ResponseHandler func(*HTTPResponse) error

// Renderer defines new renderer object
type Renderer struct {
	mx sync.Mutex

	Path         string
	Postfix      string
	SourceFS     fs.FS
	Params       Params
	DelimStart   string
	DelimEnd     string
	CacheEnabled bool

	funcs    template.FuncMap
	handlers map[int]ResponseHandler
	cache    map[string]*template.Template
}

// New creates new template Renderer with some option params
// @param path - to the directory of templates
// @param postfix - after file name. You can Renderer template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func New(path, postfix string, enabledCache bool) *Renderer {
	return NewFS(nil, path, postfix, enabledCache)
}

// NewFS creates new template Renderer with some option params for FS object
// @param fs - preinited directory in memory
// @param path - to the directory of templates inside FS
// @param postfix - after file name. You can Renderer template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewFS(fs fs.FS, path, postfix string, enabledCache bool) *Renderer {
	if len(postfix) > 1 {
		postfix = "." + strings.TrimLeft(postfix, ".")
	}
	return &Renderer{
		Path:         path,
		Postfix:      postfix,
		SourceFS:     fs,
		CacheEnabled: enabledCache,
		cache:        make(map[string]*template.Template),
		funcs:        make(template.FuncMap),
	}
}

// Func register function in template Renderer
func (r *Renderer) Func(key string, value any) *Renderer {
	r.funcs[key] = value
	return r
}

// Funcs register functions in template Renderer
func (r *Renderer) Funcs(funcs template.FuncMap) *Renderer {
	for fkey, fk := range funcs {
		r.funcs[fkey] = fk
	}
	return r
}

// FuncList returns the list of prepared template function
func (r *Renderer) FuncList() template.FuncMap {
	return r.funcs
}

// SetDelims of the template
func (r *Renderer) SetDelims(start, end string) *Renderer {
	r.DelimStart = start
	r.DelimEnd = end
	r.ResetCache()
	return r
}

// RegisterHandler for reaction for some response code
func (r *Renderer) RegisterHandler(code int, handler ResponseHandler) *Renderer {
	if r.handlers == nil {
		r.handlers = make(map[int]ResponseHandler)
	}
	r.handlers[code] = handler
	return r
}

// Template parse and return new template object with all related sub templates
func (r *Renderer) Template(templates ...string) (*template.Template, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	key := strings.Join(templates, ":")
	if t, ok := r.cache[key]; ok {
		return t, nil
	}
	var (
		templ   = template.New("").Funcs(r.funcs)
		exclude = []string{}
	)
	if r.DelimStart != "" {
		templ = templ.Delims(r.DelimStart, r.DelimEnd)
	}
	if err := r.initTemplates(templ, templates, &exclude); nil != err {
		return nil, err
	}
	if r.CacheEnabled {
		r.cache[key] = templ
	}
	return templ, nil
}

// Render template to the writer interface
// The last template in the list will be main rendering template
//
// Example:
// render.Render(out, nil, "layouts/main", "index") // "index" as a target template
func (r *Renderer) Render(w io.Writer, params Params, templates ...string) (err error) {
	if params == nil {
		params = make(Params, len(r.Params))
	}
	for key, val := range r.Params {
		params[key] = val
	}
	var tpl *template.Template
	if tpl, err = r.Template(templates...); err == nil {
		err = tpl.ExecuteTemplate(w, templates[len(templates)-1], params)
	}
	return err
}

// RenderResponse prepared in response object
func (r *Renderer) RenderResponse(resp *HTTPResponse) error {
	if r.handlers != nil {
		if handler, ok := r.handlers[resp.Code]; handler != nil && ok {
			return handler(resp)
		}
	}
	return r.Render(resp.Writer, resp.Context, resp.Template)
}

// HTTPHandler wraps http handler as render function
//
// Example:
// mux := http.NewServeMux()
// mux.HandleFunc("/", render.HTTPHandler(func(w http.ResponseWriter, r *http.Request) *HTTPResponse { return Response(http.StatusOK, "index", params) })
// mux.HandleFunc("/hello", render.HTTPHandler(getHello))
func (r *Renderer) HTTPHandler(f HTTPResponseHandler) http.HandlerFunc {
	return HTTPHandler(r, f)
}

// ResetCache of templates
func (r *Renderer) ResetCache() {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.cache = make(map[string]*template.Template, len(r.cache))
}

///////////////////////////////////////////////////////////////////////////////
// Internal
///////////////////////////////////////////////////////////////////////////////

func (r *Renderer) initTemplates(t *template.Template, tmps []string, exclude *[]string) error {
	firstLevel := len(*exclude) == 0
	for tkey, tpl := range r.prepareTemplates(tmps...) {
		if t.Lookup(tkey) == nil {
			if data, err := r.readFile(tpl); err == nil {
				var (
					ntemplates []string
					tmps       = templatesRegex.FindAllStringSubmatch(string(data), -1)
				)
				if len(tmps) > 0 {
					for _, it := range tmps {
						if sIndexOf(it[1], *exclude) < 0 {
							*exclude = append(*exclude, it[1])
							ntemplates = append(ntemplates, it[1])
						}
					}
				}

				// Prepare new templates
				if len(ntemplates) > 0 {
					if err = r.initTemplates(t, ntemplates, exclude); nil != err {
						return err
					}
				}

				if _, err = t.New(tkey).Parse(string(data)); nil != err {
					return err
				}
			} else if firstLevel {
				return err
			}
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// Helpers
///////////////////////////////////////////////////////////////////////////////

func (r *Renderer) readFile(filename string) ([]byte, error) {
	if r.SourceFS != nil {
		file, err := r.SourceFS.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)
	}
	return os.ReadFile(filename)
}

func (r *Renderer) prepareTemplates(templates ...string) map[string]string {
	ntpls := make(map[string]string, len(templates))
	for _, t := range templates {
		fpath := filepath.Join(r.Path, t+r.Postfix)
		ntpls[t] = fpath
	}
	return ntpls
}

func sIndexOf(v string, arr []string) int {
	for i, s := range arr {
		if s == v {
			return i
		}
	}
	return -1
}
