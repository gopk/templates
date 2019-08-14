//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015
//

package templates

import (
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	templatesRegex = regexp.MustCompile("\\{\\{\\s*template\\s*['\"]([^'\"]+)['\"][^\\}]*\\}\\}")
)

type RseponseHandler func(*HttpResponse) error

type TemplateRender struct {
	path         string
	postfix      string
	funcs        template.FuncMap
	handlers     map[int]RseponseHandler
	cache        map[string]*template.Template
	Params       map[string]interface{}
	CacheEnabled bool
}

// MakeRender creates new template render with some option params
// @param path - to the directory of templates
// @param postfix - after file name. You can render template just with name "index", "search"
//                  and etc and set the extension of file in the postfix
// @param enabledCache - option
func MakeRender(path, postfix string, enabledCache bool) *TemplateRender {
	if len(postfix) > 1 {
		postfix = "." + postfix
	}
	return &TemplateRender{
		path:         path,
		postfix:      postfix,
		CacheEnabled: enabledCache,
		cache:        make(map[string]*template.Template),
		funcs:        make(template.FuncMap),
	}
}

// Func register function in template render
func (r *TemplateRender) Func(key string, value interface{}) *TemplateRender {
	r.funcs[key] = value
	return r
}

// RegisterHandler for reaction for some response code
func (r *TemplateRender) RegisterHandler(code int, handler RseponseHandler) *TemplateRender {
	if r.handlers == nil {
		r.handlers = make(map[int]RseponseHandler)
	}
	r.handlers[code] = handler
	return r
}

// Template returns inited template object
func (r *TemplateRender) Template(templates ...string) (*template.Template, error) {
	key := strings.Join(templates, ":")
	if t, ok := r.cache[key]; ok {
		return t, nil
	}

	t := template.New("").Funcs(r.funcs)

	exclude := []string{}
	if err := r.initTemplates(t, templates, &exclude); nil != err {
		return nil, err
	}

	if r.CacheEnabled {
		r.cache[key] = t
	}

	return t, nil
}

// Render template to the writer interface
func (r *TemplateRender) Render(w io.Writer, params map[string]interface{}, temps ...string) (err error) {
	if params == nil {
		params = map[string]interface{}{}
	}

	for key, val := range r.Params {
		params[key] = val
	}

	var tpl *template.Template
	if tpl, err = r.Template(temps...); err == nil {
		err = tpl.ExecuteTemplate(w, temps[len(temps)-1], params)
	}
	return
}

// RenderResponse prepared in response object
func (r *TemplateRender) RenderResponse(resp *HttpResponse) error {
	if r.handlers != nil {
		if handler, ok := r.handlers[resp.Code]; handler != nil && ok {
			return handler(resp)
		}
	}
	return r.Render(resp.Writer, resp.Context, resp.Template)
}

///////////////////////////////////////////////////////////////////////////////
// Internal
///////////////////////////////////////////////////////////////////////////////

func (r *TemplateRender) initTemplates(t *template.Template, tmps []string, exclude *[]string) error {
	firstLevel := 0 == len(*exclude)
	for tkey, tpl := range r.prepareTemplates(tmps...) {
		if t.Lookup(tkey) == nil {
			if data, err := ioutil.ReadFile(tpl); err == nil {
				tmps := templatesRegex.FindAllStringSubmatch(string(data), -1)

				ntemplates := []string{}
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

func (r *TemplateRender) prepareTemplates(templates ...string) map[string]string {
	ntpls := make(map[string]string)
	for _, t := range templates {
		fpath := filepath.Join(r.path, t+r.postfix)
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
