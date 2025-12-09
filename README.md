# Logit —— 高性能结构化日志组件（支持上下文聚合日志）

Logit 是一个基于 `zap` 封装的高性能日志组件，支持 **结构化日志、离散字段聚合输出、有序字段维护、日志等级字段隔离、元数据字段保留、日志切分** 等特性。
适用于微服务、企业级后端系统、API Gateway、任务系统、RPC 服务等。

---

## ✨ 特性亮点

### 高性能日志内核

* 采用 `zapcore.Core` 实现
* 零字符串拼接
* 字段化输出
* 可视化查询友好

---

### 上下文日志聚合能力（核心能力）

支持以下模式：

```go
ctx := logger.NewContext(ctx)

logger.AddField(ctx, zap.String("uid", "10001"))
logger.AddField(ctx, zap.String("action", "pay"))
logger.AddLevelField(ctx, zap.ErrorLevel, zap.String("errCode", "50001"))
logger.AddMetaField(ctx, zap.String("trace_id", "xxxx"))
```

最终输出为：

```json
{
  "trace_id": "xxxx",
  "uid": "10001",
  "action": "pay",
  "errCode": "50001",
  ...
}
```

- ✔ 仅在一次函数执行结束时输出
- ✔ 避免业务层多点日志污染
- ✔ 聚合信息更完整

---

### 字段有序、可覆盖、可删除

写入顺序严格保持：

```
logit.AddField(uid=Tom)
logit.AddField(time=300ms)
logit.AddField(uid=Jack) → 会覆盖但位置不变
```

最终结构：

```
uid=Jack → time=300ms
```

删除：

```
logit.RemoveField(ctx,"uid")
```

查找：

```
logit.FindFiedl(ctx,"uid")
```

级别隔离：

```
logit.AddLevelField(ctx,zap.ErrorLevel, logit.String("errCode", "E500"))
```

只有 Error 才输出。

---

### 元数据字段（Metadata）

* Request ID
* Trace ID
* Span ID
* Host 信息
* 用户身份
* 部署版本等

使用：

```
logit.AddMetaField(ctx, zap.String("trace", "xyz"))
```

日志等级无关均输出。

用途：

* 请求级 tracing
* 业务侧全链路记录
* 统一字段集化输出

---

### 日志切分支持（Rolling）

基于 lumberjack 实现：

支持功能：

* 按大小切割
* 按日期限制存活周期
* 压缩 `.gz`
* 保留最近 N 份日志

---

## 📦 安装

```shell
go get github.com/lifei6671/logit
```

---

## 🔧 使用示例

### 初始化（建议在 main.go 中执行）

```go
logger := logit.New(logger.Config{
	Filename:   "./app.log",
	MaxSize:    100,
	MaxBackups: 7,
	MaxAge:     10,
	Compress:   true,
	Level:      "debug",
	ToStdout:   true,
})
defer logger.Sync()
```

---

## 🧠 上下文日志聚合示例

### 推荐使用方式

```go
func BizHandler(ctx context.Context) error {
	ctx = logger.WithContext(ctx)
	defer logger.Sync(ctx)

	logger.AddMetadata(ctx, zap.String("trace_id", "abc123"))
	logger.AddField(ctx, zap.String("step", "input_processed"))

	result, err := queryDB(ctx)
	if err != nil {
		logger.AddLevelField(ctx, zap.ErrorLevel, zap.String("dbError", err.Error()))
		logger.Error(ctx, "DB failed")
		return err
	}

	logger.AddField(ctx, zap.Any("dbResult", result))
	logger.Info(ctx, "BizHandler success")
	return nil
}
```

最终输出字段为：

* trace_id
* step
* dbResult or dbError（自动级别控制）
* 执行时间（如自行加入）

---

## 🔄 与标准库Slog集成

Logit支持与Go标准库中的`slog`集成，可将Zap日志组件包装为`slog`日志组件，示例如下：

```go
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
```

## 🔍 调试日志输出示例

```go
logger.Info(ctx,"service started",
    logit.String("version", "1.0"),
    logit.Int("pid", os.Getpid()))
```

---

## 📁 日志输出格式示例

**Info 日志示例**

```json
{
  "time": "2025-01-08 10:02:33",
  "level": "info",
  "msg": "BizHandler success",
  "trace_id": "abc123",
  "step": "input_processed",
  "dbResult": {"count": 10, "status": "ok"}
}
```

**Error 日志示例**

```json
{
  "time": "2025-01-08 10:02:33",
  "level": "error",
  "msg": "DB failed",
  "trace_id": "abc123",
  "dbError": "connection refused"
}
```

---

## 🧬 推荐与框架集成方式

### Gin 框架

中间件可注入：

```go
func WithRequestLog(ctx context.Context) context.Context {
	traceID := generateTraceID()
	newCtx := logit.NewContext(ctx)
	logit.AddMetaField(ctx, logit.String("trace_id", traceID))
	return ctx
}
```

### gRPC

拦截器中：

```go
ctx := logit.NewContext(ctx)
defer logit.Flush(ctx)
```

---

## 🏗 设计原则

* 高性能（基于 zap 的零拷贝/低分配）
* 字段稳定排序（更利于分析日志）
* 按级别隔离字段
* 可覆盖逻辑更利于数据更新
* 元数据不丢失
* `final flush` 统一输出策略

适用于：

* 高 QPS 请求日志聚合
* 大型业务逻辑链路记录
* 支付业务关键链路追踪
* APM tracing 替代存储方式

---


## 📝 常用API说明

### 日志字段相关
- `AddField(ctx context.Context, field zap.Field)`：向上下文添加普通字段
- `AddMetaField(ctx context.Context, field zap.Field)`：添加元数据字段，所有日志级别都会输出
- `AddLevelField(ctx context.Context, lvl zapcore.Level, field zap.Field)`：添加指定级别字段，仅对应级别日志输出
- `AddDebug(ctx context.Context, fields ...zap.Field)`：添加Debug级别字段
- `AddInfo(ctx context.Context, fields ...zap.Field)`：添加Info级别字段
- `AddWarn(ctx context.Context, fields ...zap.Field)`：添加Warn级别字段
- `AddError(ctx context.Context, fields ...zap.Field)`：添加Error级别字段
- `AddFatal(ctx context.Context, fields ...zap.Field)`：添加Fatal级别字段
- `RemoveField(ctx context.Context, key string)`：删除指定字段
- `FindField(ctx context.Context, key string) (zap.Field, bool)`：查找指定字段
- `FindMetaField(ctx context.Context, key string) (zap.Field, bool)`：查找元数据字段

### 日志写入相关
- `Debug(ctx context.Context, msg string, fields ...zap.Field)`：输出Debug级别日志
- `Info(ctx context.Context, msg string, fields ...zap.Field)`：输出Info级别日志
- `Warn(ctx context.Context, msg string, fields ...zap.Field)`：输出Warn级别日志
- `Error(ctx context.Context, msg string, fields ...zap.Field)`：输出Error级别日志
- `Fatal(ctx context.Context, msg string, fields ...zap.Field)`：输出Fatal级别日志
- `Panic(ctx context.Context, msg string, fields ...zap.Field)`：输出Panic级别日志
- `Sync() error`：同步日志到磁盘

### 上下文相关
- `WithContext(ctx context.Context) context.Context`：将日志字段容器嵌入上下文
- `NewContext(ctx context.Context) context.Context`：初始化新的日志容器并嵌入上下文
- `Flush(ctx context.Context)`：将各级别日志统一写入磁盘

## 🚀 性能考量

1. **基于Zap内核**：Logit使用Zap作为底层日志内核，继承了其高性能特性，包括零字符串拼接和低内存分配
2. **缓冲机制**：通过上下文聚合日志字段，减少IO操作次数，提高性能
3. **异步写入**：支持异步写入日志，避免阻塞业务流程
4. **字段管理**：高效的字段管理机制，支持字段的添加、覆盖、删除和查找，操作复杂度低

## ❓ 常见问题

### 如何确保日志字段的顺序？
Logit会严格按照字段添加的顺序维护字段，后续添加的同名字段会覆盖之前的字段，但位置保持不变。

### 元数据字段和普通字段有什么区别？
元数据字段会在所有级别的日志中输出，而普通字段和级别字段则根据日志级别决定是否输出。

### 如何处理日志文件过大的问题？
Logit基于lumberjack实现了日志切分功能，可配置按大小切割、按日期限制存活周期、压缩旧日志等。

### 如何在分布式系统中追踪请求？
可以通过`AddMetaField`添加`trace_id`等追踪标识，这些标识会在所有相关日志中出现，便于追踪整个请求链路。

## 📄 License

MIT
