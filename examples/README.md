# Go-Inject 使用示例

本目录包含了 go-inject 依赖注入框架的各种使用示例，从基础用法到高级特性，帮助您快速上手和深入理解。

## 示例列表

### 1. [基础示例 (basic)](./basic/)
演示 go-inject 的基本用法：
- 定义服务接口和实现
- 使用 `inject:""` 标签进行依赖注入
- 使用 `inject.Populate()` 便捷函数
- 简单的依赖关系管理

**运行方式：**
```bash
cd examples/basic
go run main.go
```

### 2. [深度注入示例 (deep-injection)](./deep-injection/)
展示 go-inject 的深度注入特性：
- 手动创建的对象自动进行依赖注入
- 多层嵌套的依赖关系
- 模拟真实应用场景（Controller -> Service -> Repository）
- 演示框架集成场景

**运行方式：**
```bash
cd examples/deep-injection
go run main.go
```

**特色功能：**
- 即使对象是手动创建的，其内部依赖也会被自动注入
- 支持任意深度的嵌套依赖关系
- 适合与现有框架集成

### 3. [Web 服务示例 (web-service)](./web-service/)
完整的 Web 应用依赖注入示例：
- HTTP 服务器和路由处理
- 分层架构（Handler -> Service -> Repository）
- RESTful API 实现
- 实际的业务逻辑演示

**运行方式：**
```bash
cd examples/web-service
go run main.go
```

**测试 API：**
```bash
# 获取用户
curl http://localhost:8080/users/1

# 创建用户
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

### 4. [高级示例 (advanced)](./advanced/)
展示 go-inject 的高级特性：
- 命名注入（Named Injection）
- 多实例管理
- 复杂的依赖配置
- 依赖图信息查看

**运行方式：**
```bash
cd examples/advanced
go run main.go
```

**高级特性：**
- 使用 `inject:"name"` 进行命名注入
- 同一接口的多个实现
- 灵活的依赖配置策略

### 5. [容器示例 (container)](./container/)
演示使用 NewContainer 创建 IoC 容器的方式：
- 使用 `NewContainer()` 创建容器实例
- 通过 `Provides()` 方法批量注册服务
- 使用 `ProvideWithName()` 进行命名注册
- 容器化的依赖管理方式

**运行方式：**
```bash
cd examples/container
go run main.go
```

**容器特性：**
- 面向对象的容器管理
- 批量服务注册
- 支持命名服务注册
- 更清晰的生命周期管理

## 核心概念说明

### 1. 依赖注入标签
```go
type Service struct {
    Logger   Logger   `inject:""`           // 自动注入
    Database Database `inject:"primary"`    // 命名注入
}
```

### 2. 三种使用方式

#### 便捷函数方式（推荐用于简单场景）
```go
err := inject.Populate(
    &Logger{},
    &Database{},
    &Service{},
    &app,
)
```

#### Graph 方式（推荐用于复杂场景）
```go
var g inject.Graph
g.Provide(&inject.Object{Value: &Logger{}})
g.Provide(&inject.Object{Value: &Database{}, Name: "primary"})
g.Populate()
```

#### Container 方式（推荐用于面向对象场景）
```go
container := inject.NewContainer()
container.Provides(&Logger{}, &Database{}, &Service{}, &app)
// 或使用命名注册
container.ProvideWithName("primary", &Database{})
container.Populate()
```

### 3. 深度注入
go-inject 会自动发现和注入手动创建对象的依赖：
```go
// 手动创建的对象
controller := &Controller{
    Service: &Service{}, // 这个 Service 的依赖会被自动注入
}

// 注入时会处理 controller 及其内部的 Service
inject.Populate(dependencies..., &app{Controller: controller})
```

## 最佳实践

1. **接口优先**：定义接口而不是直接依赖具体实现
2. **单一职责**：每个服务专注于单一功能
3. **分层架构**：清晰的分层有助于依赖管理
4. **命名注入**：当需要同一接口的多个实现时使用命名注入
5. **错误处理**：始终检查 `Populate()` 的返回错误

## 与 Facebook inject 的区别

go-inject 基于 Facebook 的 inject 库，但增加了以下增强功能：

1. **深度注入**：自动处理手动创建对象的依赖注入
2. **更好的错误信息**：提供更详细的错误诊断
3. **性能优化**：改进了依赖解析算法
4. **Go 语言特性**：更好地利用 Go 的类型系统和反射机制

## 常见问题

### Q: 为什么我的依赖没有被注入？
A: 检查以下几点：
- 字段是否为导出字段（首字母大写）
- 是否添加了 `inject:""` 标签
- 是否在 `Populate()` 中提供了对应的依赖

### Q: 如何处理循环依赖？
A: go-inject 会检测并报告循环依赖错误。重新设计您的架构以避免循环依赖。

### Q: 可以注入值类型吗？
A: 不可以，go-inject 只支持指针类型的注入。

### Q: 如何在测试中使用？
A: 可以轻松替换依赖实现：
```go
// 生产环境
inject.Populate(&RealDatabase{}, &app)

// 测试环境
inject.Populate(&MockDatabase{}, &app)
```

## 更多信息

- [项目主页](https://github.com/ComingCL/go-inject)
- [API 文档](https://pkg.go.dev/github.com/ComingCL/go-inject)
- [Facebook inject 原项目](https://github.com/facebookarchive/inject)