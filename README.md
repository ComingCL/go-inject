# Go-Inject - Enhanced Dependency Injection Framework

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)](.)

Go-Inject 是一个基于 Facebook 的 [inject](https://github.com/facebookarchive/inject) 项目增强的 Go 语言依赖注入框架。它提供了基于反射的依赖注入功能，并在原有基础上增加了**深度注入**等重要特性。

## 🚀 主要特性

### 原有特性（继承自 Facebook inject）
- ✅ **基于反射的依赖注入**：自动解析和注入依赖关系
- ✅ **结构体标签支持**：使用 `inject:""` 标签标记需要注入的字段
- ✅ **命名注入**：支持通过名称注入特定实例
- ✅ **私有注入**：支持创建私有实例
- ✅ **接口注入**：自动匹配实现了接口的类型
- ✅ **内联结构体**：支持内联结构体的依赖注入
- ✅ **循环依赖检测**：自动检测并报告循环依赖
- ✅ **日志支持**：可配置的注入过程日志记录

### 🆕 增强特性
- 🎯 **深度注入（Deep Injection）**：自动注入手动创建对象的内部依赖
- 🔄 **递归依赖解析**：支持多层嵌套的依赖关系自动解析
- 🛡️ **增强的错误处理**：更详细的错误信息和调试支持
- 📊 **完善的测试覆盖**：包含深度注入的完整测试用例

## 📦 安装

```bash
go get github.com/ComingCL/go-inject
```

## 🎯 深度注入特性详解

### 什么是深度注入？

深度注入是本项目相对于原始 Facebook inject 的主要增强功能。它解决了以下场景：

**场景描述**：当你手动创建了一个对象，该对象内部包含其他需要依赖注入的字段时，传统的依赖注入框架无法处理这种情况。

```go
// 传统方式：无法自动注入 d.A 内部的依赖
d := &Service{
    A: &ComponentA{}, // 手动创建的对象，内部的 C 字段无法被自动注入
}
```

**深度注入解决方案**：自动检测并注入手动创建对象的内部依赖。

### 深度注入工作原理

1. **自动发现**：框架检测到字段中存在手动创建的对象
2. **自动注册**：将发现的对象自动注册到依赖图中
3. **递归注入**：递归地为该对象的所有依赖字段进行注入
4. **多层支持**：支持任意深度的嵌套依赖关系

## 📚 使用指南

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/ComingCL/go-inject"
)

// 定义服务接口
type Logger interface {
    Log(message string)
}

// 实现日志服务
type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(message string) {
    fmt.Println("LOG:", message)
}

// 定义数据库服务
type Database struct {
    Logger Logger `inject:""`
}

func (d *Database) Query(sql string) {
    d.Logger.Log("Executing: " + sql)
}

// 定义用户服务
type UserService struct {
    DB     *Database `inject:""`
    Logger Logger    `inject:""`
}

func (u *UserService) GetUser(id int) {
    u.Logger.Log(fmt.Sprintf("Getting user %d", id))
    u.DB.Query("SELECT * FROM users WHERE id = ?")
}

func main() {
    var g inject.Graph
    
    // 注册依赖
    logger := &ConsoleLogger{}
    db := &Database{}
    userService := &UserService{}
    
    g.Provide(&inject.Object{Value: logger})
    g.Provide(&inject.Object{Value: db})
    g.Provide(&inject.Object{Value: userService})
    
    // 执行依赖注入
    if err := g.Populate(); err != nil {
        panic(err)
    }
    
    // 使用服务
    userService.GetUser(123)
}
```

### 深度注入示例

```go
package main

import (
    "fmt"
    "github.com/ComingCL/go-inject"
)

type ComponentA struct {
    C *ComponentC `inject:""`
}

type ComponentB struct{}

type ComponentC struct {
    B *ComponentB `inject:""`
}

type Service struct {
    A *ComponentA `inject:""`
}

func main() {
    var g inject.Graph
    
    // 注册基础依赖
    b := &ComponentB{}
    c := &ComponentC{}
    
    // 创建包含手动实例的服务
    service := &Service{
        A: &ComponentA{}, // 手动创建的实例
    }
    
    g.Provide(&inject.Object{Value: b})
    g.Provide(&inject.Object{Value: c})
    g.Provide(&inject.Object{Value: service})
    
    // 执行依赖注入（包括深度注入）
    if err := g.Populate(); err != nil {
        panic(err)
    }
    
    // 验证深度注入结果
    fmt.Printf("service.A.C != nil: %v\n", service.A.C != nil)         // true
    fmt.Printf("service.A.C.B != nil: %v\n", service.A.C.B != nil)     // true
    fmt.Printf("service.A.C.B == b: %v\n", service.A.C.B == b)         // true
}
```

### 命名注入

```go
type Config struct {
    DatabaseURL string
    RedisURL    string
}

type Service struct {
    MainDB  *Database `inject:"main_db"`
    CacheDB *Database `inject:"cache_db"`
}

func main() {
    var g inject.Graph
    
    mainDB := &Database{URL: "postgres://main"}
    cacheDB := &Database{URL: "redis://cache"}
    service := &Service{}
    
    g.Provide(&inject.Object{Value: mainDB, Name: "main_db"})
    g.Provide(&inject.Object{Value: cacheDB, Name: "cache_db"})
    g.Provide(&inject.Object{Value: service})
    
    g.Populate()
}
```

### 私有注入

```go
type Service struct {
    Logger Logger `inject:"private"`  // 创建私有实例
}
```

## 🏗️ 项目结构

```
go-inject/
├── README.md              # 项目文档
├── LICENSE               # MIT 许可证
├── go.mod               # Go 模块文件
├── inject.go            # 核心注入逻辑
├── inject_test.go       # 测试用例
├── structtag.go         # 结构体标签解析
├── structtag_test.go    # 标签解析测试
├── ioc_container.go     # IoC 容器实现
└── examples/            # 使用示例
    ├── basic/           # 基础用法示例
    ├── deep-injection/  # 深度注入示例
    ├── web-service/     # Web 服务示例
    └── advanced/        # 高级用法示例
```

## 🧪 测试

运行所有测试：

```bash
go test -v
```

运行特定测试：

```bash
go test -run TestForDeepInject -v
```

查看测试覆盖率：

```bash
go test -cover
```

## 📈 性能对比

| 特性 | Facebook inject | Go-Inject |
|------|----------------|-----------|
| 基础注入 | ✅ | ✅ |
| 深度注入 | ❌ | ✅ |
| 递归依赖 | 部分支持 | ✅ 完全支持 |
| 错误处理 | 基础 | 增强 |
| 测试覆盖 | 基础 | 完整 |

## 🔄 从 Facebook inject 迁移

如果你正在使用 Facebook 的 inject 包，迁移到 go-inject 非常简单：

1. 更新导入路径：
```go
// 旧的
import "github.com/facebookgo/inject"

// 新的
import "github.com/ComingCL/go-inject"
```

2. 代码无需修改，所有原有功能保持兼容

3. 可选：利用新的深度注入特性优化你的代码

## 🤝 贡献指南

我们欢迎社区贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/ComingCL/go-inject.git
cd go-inject

# 运行测试
go test -v

# 检查代码格式
go fmt ./...

# 运行静态分析
go vet ./...
```

## 📄 许可证

本项目基于 MIT 许可证开源。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 感谢 Facebook 团队开源的原始 [inject](https://github.com/facebookarchive/inject) 项目
- 感谢所有为本项目做出贡献的开发者

## 📞 支持

如果你遇到问题或有建议，请：

1. 查看 [examples](examples/) 目录中的示例
2. 搜索现有的 [Issues](https://github.com/ComingCL/go-inject/issues)
3. 创建新的 Issue 描述你的问题

## 🔗 相关链接

- [Go 官方文档](https://golang.org/doc/)
- [依赖注入模式](https://en.wikipedia.org/wiki/Dependency_injection)
- [原始 Facebook inject 项目](https://github.com/facebookarchive/inject)

---

**Go-Inject** - 让依赖注入更简单、更强大！ 🚀