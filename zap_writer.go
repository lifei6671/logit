package logit

import (
	"context"
	"fmt"
	"io"

	"github.com/lifei6671/rotatefiles"
	"go.uber.org/zap/zapcore"
)

// WriterBuilder 日志写构建函数
type WriterBuilder func(ruleName, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, rotatefiles.RotateGenerator, error)

// DefaultWriterBuild 默认构建器
func DefaultWriterBuild(ruleName, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, rotatefiles.RotateGenerator, error) {
	return BuildZapWriteSyncer(ruleName, filename, opts...)
}

// BuildZapWriteSyncer 实现自动按日期分隔的日志写入器
func BuildZapWriteSyncer(ruleName string, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, rotatefiles.RotateGenerator, error) {
	o := &BuildZapWriterOption{}
	for _, f := range opts {
		f(o)
	}
	generator, err := rotatefiles.NewSimpleRotateGenerator(ruleName, filename, o.OnError)
	if err != nil {
		return nil, nil, err
	}

	if sErr := generator.Start(context.Background()); sErr != nil {
		return nil, nil, fmt.Errorf("start generator err:%s", sErr)
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

	var rOpts []rotatefiles.RotateFileOption
	if o.OnError != nil {
		rOpts = append(rOpts, rotatefiles.WithOnErr(o.OnError))
	}

	w, err := rotatefiles.NewRotateFile(opt, rOpts...)
	if err != nil {
		return nil, generator, err
	}

	return zapcore.AddSync(w), generator, nil
}
