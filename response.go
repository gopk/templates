//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015
//

package templates

import (
  "net/http"
)

type HttpResponseHandler func(w http.ResponseWriter, r *http.Request) *HttpResponse

type HttpResponse struct {
  Request  *http.Request
  Writer   http.ResponseWriter
  Template string
  Context  map[string]interface{}
  Error    error
  Code     int
}

func Response(code int, template string, ctx map[string]interface{}) *HttpResponse {
  return &HttpResponse{Code: code, Template: template, Context: ctx}
}

func (resp *HttpResponse) SetBase(w http.ResponseWriter, r *http.Request) *HttpResponse {
  resp.Writer = w
  resp.Request = r
  return resp
}

func (resp *HttpResponse) UpdateContext(ctx map[string]interface{}) *HttpResponse {
  if nil == resp.Context {
    resp.Context = ctx
  } else if nil != ctx {
    for k, v := range resp.Context {
      resp.Context[k] = v
    }
  }
  return resp
}

func (resp *HttpResponse) SetContext(ctx map[string]interface{}) *HttpResponse {
  resp.Context = ctx
  return resp
}

func HttpHandler(render *TemplateRender, f HttpResponseHandler) http.HandlerFunc {
  if nil == render {
    render = GlobalRender
  }
  return func(w http.ResponseWriter, r *http.Request) {
    resp := f(w, r)
    if nil != resp {
      resp.Writer = w
      resp.Request = r
      render.RenderResponse(resp)
    } else {
      http.Error(w, "Invalid http response", http.StatusInternalServerError)
    }
  }
}
