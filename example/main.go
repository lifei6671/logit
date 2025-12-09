package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/lifei6671/logit"
)

func main() {
	logger := logit.New(logit.Config{
		Filename:   "./app.log",
		MaxSize:    100, // MB
		MaxBackups: 10,
		MaxAge:     7, // days
		Compress:   true,
		Level:      "debug",
		ToStdout:   false,
	})

	ctx := logit.WithContext(context.Background())

	logit.AddMetaFields(ctx, logit.Any("host", "127.0.0.1"))

	logger.Info(ctx, "service started", logit.String("service", "service"))

	logger.Debug(ctx, "debugging user info", zap.String("user", "Tom"), zap.Int("age", 30))

	logger.Error(ctx, "error log", logit.Int64("logid", time.Now().Unix()))

	for i := 0; i < 1000; i++ {
		logit.AddInfo(ctx, zap.Int("index", i))
		time.Sleep(1 * time.Millisecond)
	}
	logger.Info(ctx, "batch log", zap.String("log", "finish"))
	_ = logger.Sync()
}
