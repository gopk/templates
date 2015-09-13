//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015
//

package templates

import (
  "html/template"
  "io/ioutil"
  "net/http"
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
  CacheEnabled bool
}

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

func (r *TemplateRender) Func(key string, value interface{}) *TemplateRender {
  r.funcs[key] = value
  return r
}

func (r *TemplateRender) RegisterHandler(code int, handler RseponseHandler) *TemplateRender {
  if nil == r.handlers {
    r.handlers = make(map[int]RseponseHandler)
  }
  r.handlers[code] = handler
  return r
}

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

func (r *TemplateRender) Render(w http.ResponseWriter, params map[string]interface{}, temps ...string) (err error) {
  var tpl *template.Template
  if tpl, err = r.Template(temps...); nil != err {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else if err = tpl.ExecuteTemplate(w, temps[len(temps)-1], params); nil != err {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  return
}

func (r *TemplateRender) RenderResponse(resp *HttpResponse) error {
  if nil != r.handlers {
    if handler, ok := r.handlers[resp.Code]; nil != handler && ok {
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
    if nil == t.Lookup(tkey) {
      if data, err := ioutil.ReadFile(tpl); nil == err {
        tmps := templatesRegex.FindAllStringSubmatch(string(data), -1)

        ntemplates := []string{}
        if nil != tmps && len(tmps) > 0 {
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
  if nil != arr {
    for i, s := range arr {
      if s == v {
        return i
      }
    }
  }
  return -1
}
