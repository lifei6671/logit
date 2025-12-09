package logit_test

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/lifei6671/logit"
)

func ExampleBuildDispatchCore() {
	rules := []logit.ZapDispatch{
		{FileSuffix: "", Levels: []zapcore.Level{zapcore.InfoLevel, zapcore.DebugLevel}},
		{FileSuffix: "wf", Levels: []zapcore.Level{zapcore.WarnLevel, zapcore.ErrorLevel}},
	}

	core, closeFn, err := logit.BuildDispatchCore(
		"service",
		"service.log",
		rules,
		logit.DefaultWriterBuild,
		logit.DefaultEncoder,
		logit.WithMaxFileNum(48),
		logit.WithFlushDuration(time.Second),
	)

	logger := zap.New(core)
	defer closeFn()

	logger.Info("INFO MESSAGE")
	logger.Warn("WARNING MESSAGE", logit.Any("key", err))

}

func ExampleNewSlogLogger() {
	rules := []logit.ZapDispatch{
		{FileSuffix: "", Levels: []zapcore.Level{zapcore.InfoLevel, zapcore.DebugLevel}},
		{FileSuffix: "wf", Levels: []zapcore.Level{zapcore.WarnLevel, zapcore.ErrorLevel}},
	}

	core, closeFn, err := logit.BuildDefaultZapCore(
		"1hour",
		"service.log",
		rules,
		logit.WithMaxFileNum(48),
		logit.WithFlushDuration(time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer closeFn()

	logger := logit.NewSlogLogger(core)

	// 埋入日志容器
	ctx := logit.WithContext(context.Background())

	// 写入日志字段
	logit.AddInfo(ctx, logit.Any("key", "value"))

	// 内部自动从日志容器内汇总所有字段并合并到日志中
	logger.InfoContext(ctx, "INFO MESSAGE")
}
