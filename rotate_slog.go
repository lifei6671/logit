package logit

import (
	"time"
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
