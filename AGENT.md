# AGENT.md - httpx 项目开发指南

本文件为 AI 助手和开发者提供 httpx 项目的开发规范和最佳实践。

## 项目概述

**httpx** 是一个下一代轻量级 Go Web 框架，基于 Go 1.22+ 标准库构建，提供类型安全的 API 开发体验。

- 核心特性：泛型端点、自动 OpenAPI 生成、标准库兼容、可插拔组件
- 设计理念：拥抱标准库、类型安全、简洁高效
- 技术栈：Go 1.22+, net/http, go-playground/validator, kin-openapi

## 项目结构

```
httpx/
├── binding/          # 参数绑定（Path/Query/Header/Body/File）
├── codec/            # 统一编解码接口
├── convert/          # 类型转换工具
├── endpoint/         # 泛型端点注册
├── errors/           # 错误处理
├── examples/         # 示例代码
│   ├── 01-basic/     # 基础 CRUD
│   ├── 02-file-upload/
│   ├── 03-middleware/
│   ├── 04-binding/
│   └── 05-openapi/
├── middleware/       # 内置中间件（Logger/Recovery/CORS）
├── openapi/          # OpenAPI 3.0 文档生成
├── render/           # 响应渲染
├── validate/         # 参数校验
├── context.go        # 请求上下文
├── engine.go         # 核心引擎
├── option.go         # 配置选项
├── router.go         # 路由器
└── go.mod
```

## 核心组件

### Engine (`engine.go`)
框架核心，管理所有可插拔组件。
- 配置：通过 `httpx.With*` 选项配置
- 中间件：使用 `engine.Use()` 添加
- 启动：`engine.Start(addr)` 或 `engine.Run(httpServer)`

### Context (`context.go`)
封装 HTTP 请求和响应，提供便捷方法：
- `c.Bind(&req)` - 绑定请求参数
- `c.Validate(&req)` - 校验参数
- `c.Render(code, resp)` - 渲染响应
- `c.Param/GetQuery/GetHeader` - 获取参数
- `c.Request()` / `c.Writer()` - 访问原生对象

### Router (`router.go`)
基于 `net/http.ServeMux` 封装，支持：
- 标准 HTTP 方法：Get/Post/Put/Patch/Delete/Head/Options
- 路由模式：`/users/{id}`, `/api/*`
- 中间件：路由级中间件支持

### Endpoint (`endpoint/builder.go`)
泛型端点构建器，核心 API：
```go
endpoint.GET(engine, path, handler).
    Summary("...").
    Tags("...").
    OpenAPISpec(spec).
    Register()
```

处理函数签名：`func(context.Context, *Req) (*Resp, error)`

## 开发规范

### 代码风格
- 遵循标准 Go 代码风格
- 使用 `gofmt` 格式化代码
- 包级文档注释必不可少
- 公开 API 必须有完整注释

### 测试规范
- 新功能必须包含单元测试
- 测试文件命名：`*_test.go`
- 测试覆盖率目标：> 80%
- 运行测试：`go test ./...`

### 提交规范
- 提交信息清晰描述变更内容
- 一个提交一个关注点
- 提交前确保测试通过

## 常见任务

### 添加新的绑定类型
在 `binding/binding.go` 中：
1. 实现 `Binding` 接口
2. 在 `DefaultBinding` 中注册

### 添加新的中间件
在 `middleware/` 目录：
1. 函数签名：`func(http.Handler) http.Handler`
2. 使用标准库 `http.Handler` 接口

### 添加新的渲染格式
在 `render/` 目录：
1. 实现 `Renderer` 接口
2. 通过 `httpx.WithRenderer()` 配置

### 扩展 OpenAPI 生成
在 `openapi/openapi.go` 中：
- 使用 `kin-openapi` 库
- 通过 endpoint 构建器的 `OpenAPISpec()` 集成

## 调试技巧

### 启用详细日志
```go
engine.Use(middleware.Logger)
```

### 访问原始 Context
```go
func Handler(ctx context.Context, req *Req) (*Resp, error) {
    c := httpx.MustGetContext(ctx)
    // c.Request(), c.Writer()
}
```

### OpenAPI 调试
访问 `/openapi.json` 查看生成的规范，访问 `/docs` 查看 Swagger UI。

## 依赖管理

主要依赖：
- `github.com/go-playground/validator/v10` - 参数校验
- `github.com/go-playground/form/v4` - 表单绑定
- `github.com/getkin/kin-openapi` - OpenAPI 生成
- `google.golang.org/protobuf` - Protobuf 支持

添加依赖：`go get <module>`
更新依赖：`go get -u ./...`

## 示例参考

查看 `examples/` 目录获取完整示例：
- `01-basic` - 基础 CRUD 操作
- `02-file-upload` - 文件上传处理
- `03-middleware` - 中间件使用
- `04-binding` - 各种参数绑定方式
- `05-openapi` - OpenAPI 文档生成

## 注意事项

1. **Go 版本要求**：Go 1.22+（使用了 `http.ServeMux` 新路由语法）
2. **泛型使用**：端点处理函数必须使用泛型包装
3. **错误处理**：推荐返回自定义错误类型，通过 `WithErrorHandler` 统一处理
4. **内存管理**：文件上传后记得调用 `MultipartForm.RemoveAll()` 清理临时文件
5. **中间件顺序**：先添加的中间件先执行（洋葱模型）

## 获取帮助

- 查看 README.md 了解基本使用
- 阅读源码中的注释和文档
- 参考 examples/ 目录下的完整示例
- 运行测试用例了解各组件行为
