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

## 📄 License

MIT
