//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015
//

package templates

import (
  "html/template"
  "net/http"
)

var (
  GlobalRender *TemplateRender
)

func GetGlobalRender() *TemplateRender {
  if nil == GlobalRender {
    GlobalRender = MakeRender("", "", true)
  }
  return GlobalRender
}

func InitGlobalRender(path, postfix string, enabledCache bool) *TemplateRender {
  if nil == GlobalRender {
    GlobalRender = MakeRender(path, postfix, enabledCache)
  } else {
    GlobalRender.path = path
    GlobalRender.postfix = postfix
    GlobalRender.CacheEnabled = enabledCache
  }
  return GlobalRender
}

func Func(key string, value interface{}) {
  GetGlobalRender().Func(key, value)
}

func RegisterHandler(code int, handler RseponseHandler) *TemplateRender {
  return GetGlobalRender().RegisterHandler(code, handler)
}

func Template(templates ...string) (*template.Template, error) {
  return GetGlobalRender().Template(templates...)
}

func Render(w http.ResponseWriter, params map[string]interface{}, temps ...string) error {
  return GetGlobalRender().Render(w, params, temps...)
}

func RenderResponse(resp *HttpResponse) error {
  return GetGlobalRender().RenderResponse(resp)
}
