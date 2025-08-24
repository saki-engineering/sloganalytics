package clog

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*TraceHandler)(nil)

type TraceHandler struct {
	slog.Handler
}

func NewTraceHandler(h slog.Handler) *slog.Logger {
	return slog.New(&TraceHandler{Handler: h})
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID, ok := ctx.Value("traceID").(string)
	if ok && traceID != "" {
		r.AddAttrs(slog.String("traceID", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

// func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
// 	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
// }

// func (h *TraceHandler) WithGroup(name string) slog.Handler {
// 	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
// }
