package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"example/clog"
)

func main() {
	logger := clog.NewTraceHandler((slog.NewJSONHandler(os.Stdout, nil)))

	// traceIDをcontextに設定
	ctx := context.WithValue(context.Background(), "traceID", "trace-12345")

	fmt.Println("============================")
	fmt.Println("Default Logging:")
	fmt.Println("============================")
	logger.InfoContext(ctx, "info message", "user", "alice")
	logger.ErrorContext(ctx, "error message", "error", "something went wrong")
	// fmt.Printf("%T\n", logger.Handler())

	// groupを使用したロガーを作る
	groupedLogger := logger.With("slog-group", "my-group")

	fmt.Println("============================")
	fmt.Println("Logging with Group:")
	fmt.Println("============================")
	groupedLogger.InfoContext(ctx, "info message", "user", "bob")
	groupedLogger.ErrorContext(ctx, "error message", "error", "error occurred")
	// fmt.Printf("%T\n", groupedLogger.Handler())

}
