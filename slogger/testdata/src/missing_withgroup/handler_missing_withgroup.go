package missing_withgroup

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*TraceHandler)(nil)

type TraceHandler struct { // want "TraceHandler implements slog.Handler but does not implement WithGroup method"
	slog.Handler
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID, ok := ctx.Value("traceID").(string)
	if ok && traceID != "" {
		r.AddAttrs(slog.String("traceID", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroupメソッドは意図的に実装しない
// これにより解析器がエラーを検出することをテストする