package logit

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

type orderedField struct {
	Key   string
	Field zap.Field
}

type LogBuffer struct {
	mu sync.RWMutex

	// FIFO ordered
	metaOrder   []string
	normalOrder []string
	levelOrder  map[zapcore.Level][]string

	// key → field
	metaFields   map[string]zap.Field
	normalFields map[string]zap.Field
	levelFields  map[zapcore.Level]map[string]zap.Field
}

func newLogBuffer() *LogBuffer {
	return &LogBuffer{
		metaOrder:    []string{},
		normalOrder:  []string{},
		levelOrder:   map[zapcore.Level][]string{},
		metaFields:   map[string]zap.Field{},
		normalFields: map[string]zap.Field{},
		levelFields:  map[zapcore.Level]map[string]zap.Field{},
	}
}

// WithContext 将日志字段容器埋入 ctx 中，后续增加的字段不会立即写入磁盘
func WithContext(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}
	if b := findKeyCtx(ctx); b != nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, newLogBuffer())
}

// NewContext 初始化一个新的日志埋点容器
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, newLogBuffer())
}

func findKeyCtx(ctx context.Context) *LogBuffer {
	if ctx == nil {
		return nil
	}
	if v, ok := ctx.Value(ctxKey{}).(*LogBuffer); ok && v != nil {
		return v
	}
	return nil
}

func getBuf(ctx context.Context) *LogBuffer {
	if v := ctx.Value(ctxKey{}); v != nil {
		return v.(*LogBuffer)
	}
	return nil
}

// ---------------- 排序写入逻辑 ----------------

func ensureOrderedUpdate(order []string, key string) []string {
	for _, k := range order {
		if k == key {
			return order // key 已存在，不再重复
		}
	}
	return append(order, key)
}

// ---------------- 写入字段 --------------------

// AddField 增加单个字段
func AddField(ctx context.Context, field zap.Field) {
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.normalFields[field.Key] = field
	buf.normalOrder = ensureOrderedUpdate(buf.normalOrder, field.Key)
}

// AddMetaField 增加全局字段
func AddMetaField(ctx context.Context, field zap.Field) {
	AddMetaFields(ctx, field)
}

func AddMetaFields(ctx context.Context, fields ...zap.Field) {
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	for _, field := range fields {
		buf.metaFields[field.Key] = field
		buf.metaOrder = ensureOrderedUpdate(buf.metaOrder, field.Key)
	}
}

// AddLevelFields 增加自定义级别字段
func AddLevelFields(ctx context.Context, lvl zapcore.Level, fields ...zap.Field) {
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	if _, ok := buf.levelFields[lvl]; !ok {
		buf.levelFields[lvl] = map[string]zap.Field{}
		buf.levelOrder[lvl] = []string{}
	}
	for i := range fields {
		field := fields[i]
		buf.levelFields[lvl][field.Key] = field
		buf.levelOrder[lvl] = ensureOrderedUpdate(buf.levelOrder[lvl], field.Key)
	}

}

// AddDebug 增加 Debug 级别字段
func AddDebug(ctx context.Context, fields ...zap.Field) {
	AddLevelFields(ctx, zapcore.DebugLevel, fields...)
}

// AddInfo 增加 Info 级别字段
func AddInfo(ctx context.Context, fields ...zap.Field) {
	AddLevelFields(ctx, zapcore.InfoLevel, fields...)
}

// AddWarn 增加 Warn 级别字段
func AddWarn(ctx context.Context, fields ...zap.Field) {
	AddLevelFields(ctx, zapcore.WarnLevel, fields...)
}

// AddError 增加 Error 级别字段
func AddError(ctx context.Context, fields ...zap.Field) {
	AddLevelFields(ctx, zapcore.ErrorLevel, fields...)
}

// AddFatal 增加 Fatal 级别字段
func AddFatal(ctx context.Context, fields ...zap.Field) {
	AddLevelFields(ctx, zapcore.FatalLevel, fields...)
}

// RemoveField 删除字段
func RemoveField(ctx context.Context, key string) {
	buf := getBuf(ctx)
	if buf == nil {
		return
	}

	buf.mu.Lock()
	defer buf.mu.Unlock()

	delete(buf.normalFields, key)
	delete(buf.metaFields, key)
	for lvl := range buf.levelFields {
		delete(buf.levelFields[lvl], key)
	}
	for i, k := range buf.metaOrder {
		if k == key {
			buf.metaOrder = append(buf.metaOrder[:i], buf.metaOrder[i+1:]...)
			break
		}
	}
	for i, k := range buf.normalOrder {
		if k == key {
			buf.normalOrder = append(buf.normalOrder[:i], buf.normalOrder[i+1:]...)
		}
	}
	for lvl, keys := range buf.levelOrder {
		for i, k := range keys {
			if k == key {
				buf.levelOrder[lvl] = append(keys[:i], keys[i+1:]...)
			}
		}
	}

}

// FindField 查找指定字段
func FindField(ctx context.Context, key string) (zap.Field, bool) {
	buf := getBuf(ctx)
	if buf == nil {
		return zap.Field{}, false
	}

	buf.mu.RLock()
	defer buf.mu.RUnlock()

	if field, ok := buf.normalFields[key]; ok {
		return field, true
	}

	for _, fmap := range buf.levelFields {
		if field, ok := fmap[key]; ok {
			return field, true
		}
	}

	return zap.Field{}, false
}

// FindMetaField 查找全局字段
func FindMetaField(ctx context.Context, key string) (zap.Field, bool) {
	buf := getBuf(ctx)
	if buf == nil {
		return zap.Field{}, false
	}

	buf.mu.RLock()
	defer buf.mu.RUnlock()
	if field, ok := buf.metaFields[key]; ok {
		return field, true
	}
	return zap.Field{}, false
}
