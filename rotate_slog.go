package logit

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/lifei6671/rotatefiles"
	"go.uber.org/zap/zapcore"
)

type BuildZapWriterOption struct {
	RuleName      string
	Filename      string
	OnError       func(error)
	FlushDuration time.Duration
	CheckDuration time.Duration
	MaxFileNum    int
	BufferSize    int
}

type ZapWriterOptions func(*BuildZapWriterOption)

// WithOnErr 错误回调，当内部发生错误时会调用该方法。业务方可以实现写日志。
func WithOnErr(errFn func(err error)) ZapWriterOptions {
	return func(o *BuildZapWriterOption) {
		o.OnError = errFn
	}
}

// WithFlushDuration 周期性的将缓冲数据写入磁盘
func WithFlushDuration(duration time.Duration) ZapWriterOptions {
	return func(o *BuildZapWriterOption) {
		o.FlushDuration = duration
	}
}

// WithCheckDuration 是否周期性检查文件是否被删除，如果被删除了会自动重建
func WithCheckDuration(duration time.Duration) ZapWriterOptions {
	return func(o *BuildZapWriterOption) {
		o.CheckDuration = duration
	}
}

// WithMaxFileNum 保留的文件数量
func WithMaxFileNum(num int) ZapWriterOptions {
	return func(o *BuildZapWriterOption) {
		o.MaxFileNum = num
	}
}

// WithBufferSize 缓冲大小
func WithBufferSize(size int) ZapWriterOptions {
	return func(option *BuildZapWriterOption) {
		option.BufferSize = size
	}
}

// BuildZapWriteSyncer 实现自动按日期分隔的日志写入器
func BuildZapWriteSyncer(ruleName string, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, error) {
	o := &BuildZapWriterOption{}
	for _, f := range opts {
		f(o)
	}
	generator, err := rotatefiles.NewSimpleRotateGenerator(ruleName, filename, o.OnError)
	if err != nil {
		return nil, err
	}
	opt := &rotatefiles.RotateOption{
		RotateGenerator: generator,
		NewWriter: func(ctx context.Context, wc io.WriteCloser) (rotatefiles.AsyncWriter, error) {
			return rotatefiles.NewAsyncWriter(wc, o.BufferSize, rotatefiles.WithErrCallback(func(n int, err error) {
				o.OnError(fmt.Errorf("rotate write err: n=%d err=%w", n, err))
			})), nil
		},
		FlushDuration: o.FlushDuration,
		CheckDuration: o.CheckDuration,
		MaxFileNum:    o.MaxFileNum,
	}

	w, err := rotatefiles.NewRotateFile(opt)
	if err != nil {
		return nil, err
	}

	return zapcore.AddSync(w), nil
}
