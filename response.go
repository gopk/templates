//
// @project Templates
// @author Dmitry Ponomarev <demdxx@gmail.com> 2015, 2022
//

package templates

import (
	"net/http"
)

// HTTPResponseHandler represents default HTTP response handlerer
type HTTPResponseHandler func(w http.ResponseWriter, r *http.Request) *HTTPResponse

// HTTPResponse represents HTTP response with all necessary information
type HTTPResponse struct {
	Request  *http.Request
	Writer   http.ResponseWriter
	Template string
	Context  Params
	Error    error
	Code     int
}

// Response returns new response object for HTTPHandler wrapper
func Response(code int, template string, ctx Params) *HTTPResponse {
	return &HTTPResponse{Code: code, Template: template, Context: ctx}
}

// ExtendParams which will be passed to the renderer
func (resp *HTTPResponse) ExtendParams(ctx Params) *HTTPResponse {
	if resp.Context == nil {
		resp.Context = ctx
	} else {
		for k, v := range ctx {
			resp.Context[k] = v
		}
	}
	return resp
}

// SetParams which will be passed to the renderer
func (resp *HTTPResponse) SetParams(ctx Params) *HTTPResponse {
	resp.Context = ctx
	return resp
}

// HTTPHandler reponse wrapper for default http handler
//
// Example:
// mux := http.NewServeMux()
// mux.HandleFunc("/", HTTPHandler(render, func(w http.ResponseWriter, r *http.Request) *HTTPResponse){ return Response(http.StatusOK, "index", params) })
// mux.HandleFunc("/hello", HTTPHandler(render, getHello))
func HTTPHandler[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any](render *render[T, TT, FM], f HTTPResponseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if resp := f(w, r.WithContext(WithContext(r.Context(), render))); resp != nil {
			resp.Writer = w
			resp.Request = r
			if err := render.RenderResponse(resp); err != nil {
				http.Error(w, "Invalid response render", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Invalid http response", http.StatusInternalServerError)
		}
	}
}
