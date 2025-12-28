package logit_test

import (
	"context"
	"errors"
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

func ExampleNew() {

	ctx := logit.NewContext(context.Background())

	logger := logit.New(logit.Config{
		Filename:   "service.log",
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   true,
	})

	logger.Info(ctx, "INFO MESSAGE")
	logger.Warn(ctx, "WARNING MESSAGE", logit.Any("key", errors.New("error")))
}

func ExampleNewWithZap() {
	// 初始日志分发规则
	rules := []logit.ZapDispatch{
		{FileSuffix: "", Levels: []zapcore.Level{zapcore.InfoLevel, zapcore.DebugLevel}},
		{FileSuffix: "wf", Levels: []zapcore.Level{zapcore.WarnLevel, zapcore.ErrorLevel}},
	}

	// 初始化日志切分和清理规则
	core, closeFn, err := logit.BuildDispatchCore(
		"service",
		"service.log",
		rules,
		logit.DefaultWriterBuild,
		logit.DefaultEncoder,
		logit.WithMaxFileNum(48),
		logit.WithFlushDuration(time.Second),
	)

	// 初始化 zap 核心
	zapLogger := zap.New(core)
	defer closeFn()

	// 初始化日志包装规则
	logger := logit.NewWithZap(zapLogger)

	// 初始化日志埋点
	ctx := logit.WithContext(context.Background())

	// 增加日志字段
	logit.AddInfo(ctx, logit.Any("key", "value"))
	logit.AddError(ctx, logit.Any("errmsg", err))

	// 写入日志
	logger.Info(ctx, "INFO MESSAGE", logit.Any("key", "value"))
}
