//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015, 2022-2023
//

package templates

import (
	htmltemplate "html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	texttemplate "text/template"
)

var (
	templatesRegex = regexp.MustCompile("\\{\\{\\s*template\\s*['\"]([^'\"]+)['\"][^\\}]*\\}\\}")
)

type templateTypes interface {
	htmltemplate.Template | texttemplate.Template
}

type templateIface[T templateTypes, FM ~map[string]any] interface {
	Funcs(funcMap FM) *T
	Delims(left, right string) *T
	Lookup(name string) *T
	Parse(text string) (*T, error)
	New(name string) *T
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type templateIfaceTypes[T templateTypes, FM ~map[string]any] interface {
	*htmltemplate.Template | *texttemplate.Template
	Funcs(funcMap FM) *T
	Delims(left, right string) *T
	Lookup(name string) *T
	Parse(text string) (*T, error)
	New(name string) *T
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

// Params of the render
type Params map[string]any

// ResponseHandler for the particular HTTP code
type ResponseHandler func(*HTTPResponse) error

// render defines new render object
type render[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any] struct {
	mx sync.Mutex

	Path         string
	Postfix      string
	SourceFS     fs.FS
	Params       Params
	DelimStart   string
	DelimEnd     string
	CacheEnabled bool

	funcs    FM
	handlers map[int]ResponseHandler
	cache    map[string]T
}

type (
	HTMLRender  = render[*htmltemplate.Template, htmltemplate.Template, htmltemplate.FuncMap]
	PlainRender = render[*texttemplate.Template, texttemplate.Template, texttemplate.FuncMap]
)

// New creates new template render with some option params
//
// @param path - to the directory of templates
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func New[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any](path, postfix string, enabledCache bool) *render[T, TT, FM] {
	return NewFS[T, TT, FM](nil, path, postfix, enabledCache)
}

// NewHTML creates new template render with some option params
//
// @param path - to the directory of templates
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewHTML(path, postfix string, enabledCache bool) *HTMLRender {
	return New[*htmltemplate.Template, htmltemplate.Template, htmltemplate.FuncMap](path, postfix, enabledCache)
}

// NewPlain creates new template render with some option params
//
// @param path - to the directory of templates
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewPlain(path, postfix string, enabledCache bool) *PlainRender {
	return New[*texttemplate.Template, texttemplate.Template, texttemplate.FuncMap](path, postfix, enabledCache)
}

// NewFS creates new template render with some option params for FS object
//
// @param fs - preinited directory in memory
// @param path - to the directory of templates inside FS
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewFS[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any](fs fs.FS, path, postfix string, enabledCache bool) *render[T, TT, FM] {
	if len(postfix) > 1 {
		postfix = "." + strings.TrimLeft(postfix, ".")
	}
	return &render[T, TT, FM]{
		Path:         path,
		Postfix:      postfix,
		SourceFS:     fs,
		CacheEnabled: enabledCache,
		cache:        make(map[string]T),
		funcs:        make(FM),
	}
}

// NewHTMLFS creates new template render with some option params for FS object
//
// @param fs - preinited directory in memory
// @param path - to the directory of templates inside FS
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewHTMLFS(fs fs.FS, path, postfix string, enabledCache bool) *HTMLRender {
	return NewFS[*htmltemplate.Template, htmltemplate.Template, htmltemplate.FuncMap](fs, path, postfix, enabledCache)
}

// NewPlainFS creates new template render with some option params for FS object
//
// @param fs - preinited directory in memory
// @param path - to the directory of templates inside FS
// @param postfix - after file name. You can render template just with name "index", "search"
// and etc and set the extension of file in the postfix
// @param enabledCache - option
func NewPlainFS(fs fs.FS, path, postfix string, enabledCache bool) *PlainRender {
	return NewFS[*texttemplate.Template, texttemplate.Template, texttemplate.FuncMap](fs, path, postfix, enabledCache)
}

// Func register function in template render
func (r *render[T, TT, FM]) Func(key string, value any) *render[T, TT, FM] {
	r.funcs[key] = value
	return r
}

// Funcs register functions in template render
func (r *render[T, TT, FM]) Funcs(funcs htmltemplate.FuncMap) *render[T, TT, FM] {
	for fkey, fk := range funcs {
		r.funcs[fkey] = fk
	}
	return r
}

// FuncList returns the list of prepared template function
func (r *render[T, TT, FM]) FuncList() FM {
	return r.funcs
}

// SetDelims of the template
func (r *render[T, TT, FM]) SetDelims(start, end string) *render[T, TT, FM] {
	r.DelimStart = start
	r.DelimEnd = end
	r.ResetCache()
	return r
}

// SetDefaults params for the render as default params
func (r *render[T, TT, FM]) SetDefaults(params Params) *render[T, TT, FM] {
	r.Params = params
	return r
}

// RegisterHandler for reaction for some response code
func (r *render[T, TT, FM]) RegisterHandler(code int, handler ResponseHandler) *render[T, TT, FM] {
	if r.handlers == nil {
		r.handlers = make(map[int]ResponseHandler)
	}
	r.handlers[code] = handler
	return r
}

// Template parse and return new template object with all related sub templates
func (r *render[T, TT, FM]) Template(templates ...string) (T, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	key := strings.Join(templates, ":")
	if t, ok := r.cache[key]; ok {
		return t, nil
	}
	var (
		templ   any = r.newTemplate().Funcs(r.funcs)
		exclude     = []string{}
	)
	if r.DelimStart != "" {
		templ = templ.((templateIface[TT, FM])).Delims(r.DelimStart, r.DelimEnd)
	}
	if err := r.initTemplates(templ.(T), templates, &exclude); err != nil {
		return nil, err
	}
	if r.CacheEnabled {
		r.cache[key] = templ.(T)
	}
	return templ.(T), nil
}

// Render template to the writer interface
// The last template in the list will be main rendering template
//
// Example:
// render.Render(out, nil, "layouts/main", "index") // "index" as a target template
func (r *render[T, TT, FM]) Render(w io.Writer, params Params, templates ...string) (err error) {
	if params == nil {
		params = make(Params, len(r.Params))
	}
	for key, val := range r.Params {
		params[key] = val
	}
	var tpl T
	if tpl, err = r.Template(templates...); err == nil {
		err = tpl.ExecuteTemplate(w, templates[len(templates)-1], params)
	}
	return err
}

// RenderResponse prepared in response object
func (r *render[T, TT, FM]) RenderResponse(resp *HTTPResponse) error {
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
func (r *render[T, TT, FM]) HTTPHandler(f HTTPResponseHandler) http.HandlerFunc {
	return HTTPHandler(r, f)
}

// ResetCache of templates
func (r *render[T, TT, FM]) ResetCache() {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.cache = make(map[string]T, len(r.cache))
}

///////////////////////////////////////////////////////////////////////////////
// Internal
///////////////////////////////////////////////////////////////////////////////

func (r *render[T, TT, FM]) newTemplate() T {
	var t any = *new(T)
	switch t.(type) {
	case *htmltemplate.Template:
		var tmp any = htmltemplate.New("")
		return tmp.(T)
	case *texttemplate.Template:
		var tmp any = texttemplate.New("")
		return tmp.(T)
	}
	return nil
}

func (r *render[T, TT, FM]) initTemplates(t T, tmps []string, exclude *[]string) error {
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
					if err = r.initTemplates(t, ntemplates, exclude); err != nil {
						return err
					}
				}

				var newTpl any = t.New(tkey)
				if _, err = newTpl.(templateIface[TT, FM]).Parse(string(data)); err != nil {
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

func (r *render[T, TT, FM]) readFile(filename string) ([]byte, error) {
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

func (r *render[T, TT, FM]) prepareTemplates(templates ...string) map[string]string {
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
