package traces

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Span struct {
	goroutineID        int
	parentNamespaceCtx namespaceContext
	span               trace.Span
}

func (s *Span) End() {
	if s == nil {
		return
	}

	namespaces[s.goroutineID] = s.parentNamespaceCtx
	s.span.End()
}

func (s *Span) SetAttributes(attributes ...attribute.KeyValue) {
	s.span.SetAttributes(attributes...)
}

func StartSpan(
	name string,
	attributes ...attribute.KeyValue,
) *Span {
	id := goroutineID()
	parentNamespaceCtx, ok := namespaces[id]
	if !ok {
		panic("should call InitNamespace() before start spans.")
	}
	spanCtx, span := otel.GetTracerProvider().Tracer("gno.land").Start(
		parentNamespaceCtx.ctx,
		name,
		trace.WithAttributes(attribute.String("component", string(parentNamespaceCtx.namespace))),
		trace.WithAttributes(attributes...),
	)

	namespaces[id] = namespaceContext{namespace: parentNamespaceCtx.namespace, ctx: spanCtx}

	s := &Span{
		goroutineID:        id,
		parentNamespaceCtx: parentNamespaceCtx,
		span:               span,
	}
	return s
}
