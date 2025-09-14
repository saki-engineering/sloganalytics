package complete_handler

import (
	"context"
	"log/slog"
)

// TraceHandlerが正しくslog.Handlerインターフェースを実装していることを確認
var _ slog.Handler = (*TraceHandler)(nil)

// TraceHandler は slog.Handler インターフェースを完全に実装するハンドラー
// WithAttrs と WithGroup の両方のメソッドを明示的に実装している
type TraceHandler struct {
	slog.Handler
}

// Handle implements the slog.Handler interface by adding trace information
// and delegating to the embedded handler.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID, ok := ctx.Value("traceID").(string)
	if ok && traceID != "" {
		r.AddAttrs(slog.String("traceID", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs implements the slog.Handler interface by creating a new handler
// with the specified attributes.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup implements the slog.Handler interface by creating a new handler
// with the specified group.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}

// この実装では両方のメソッド（WithAttrs、WithGroup）が正しく実装されているため、
// 解析器はエラーを報告しないはずです。