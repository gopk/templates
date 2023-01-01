package templates

import (
	"context"
	htmltemplate "html/template"
	texttemplate "text/template"
)

var (
	ctxRender = struct{ s string }{"render"}
)

// WithContext returns context with render value
func WithContext[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any](ctx context.Context, render *render[T, TT, FM]) context.Context {
	return context.WithValue(ctx, ctxRender, render)
}

// FromContext returns render from context
func FromContext[T templateIfaceTypes[TT, FM], TT templateTypes, FM ~map[string]any](ctx context.Context) *render[T, TT, FM] {
	return ctx.Value(ctxRender).(*render[T, TT, FM])
}

// FromContextHTML returns render from context
func FromContextHTML(ctx context.Context) *HTMLRender {
	return FromContext[*htmltemplate.Template, htmltemplate.Template, htmltemplate.FuncMap](ctx)
}

// FromContextPlain returns render from context
func FromContextPlain(ctx context.Context) *PlainRender {
	return FromContext[*texttemplate.Template, texttemplate.Template, texttemplate.FuncMap](ctx)
}
