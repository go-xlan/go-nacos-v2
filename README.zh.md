[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/go-xlan/go-nacos-v2/release.yml?branch=main&label=BUILD)](https://github.com/go-xlan/go-nacos-v2/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/go-xlan/go-nacos-v2)](https://pkg.go.dev/github.com/go-xlan/go-nacos-v2)
[![Coverage Status](https://img.shields.io/coveralls/github/go-xlan/go-nacos-v2/main.svg)](https://coveralls.io/github/go-xlan/go-nacos-v2?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.23+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/go-xlan/go-nacos-v2.svg)](https://github.com/go-xlan/go-nacos-v2/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-xlan/go-nacos-v2)](https://goreportcard.com/report/github.com/go-xlan/go-nacos-v2)

# go-nacos-v2

Nacos v1 SDK 集成客户端，支持服务注册和配置管理。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->
## 英文文档

[ENGLISH README](README.md)
<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 核心特性

🎯 **服务注册**: 自动 IP 检测与临时实例注册
⚡ **服务发现**: 基于健康检查的实例选择，支持分组
🔄 **配置管理**: 从 Nacos 配置中心动态获取配置
🌍 **上下线控制**: 无需重启的优雅服务状态管理
📋 **多环境支持**: 基于命名空间和分组的服务隔离

## 安装

```bash
go get github.com/go-xlan/go-nacos-v2
```

## 快速开始

### 完整服务生命周期示例

此示例展示完整的服务注册、发现和优雅关闭处理。

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/go-xlan/go-nacos-v2/nacosv2"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
)

func main() {
	config := &nacosv2.Config{
		Endpoint:  "127.0.0.1:8848",
		AppName:   "demo1x",
		Address:   "0.0.0.0:8080",
		Group:     "DEFAULT_GROUP",
		Namespace: "public",
	}
	clientOptions := []constant.ClientOption{
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogDir("/tmp/nacos/log"),
	}
	client := rese.P1(nacosv2.NewNacosClient(config, clientOptions, zaplog.ZAP.NewZap("module", "demo1x")))

	must.Done(client.RegisterService())

	must.Done(client.Online(context.Background()))

	serviceInstance := rese.P1(client.GetServiceInstance(context.Background(), "demo1x"))
	zaplog.SUG.Debugln(neatjsons.S(serviceInstance))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()

	waitChan := make(chan os.Signal, 1)
	signal.Notify(waitChan, os.Interrupt)

	select {
	case <-waitChan:
		zaplog.SUG.Debugln("Received Ctrl+C, shutting down...")
	case <-ctx.Done():
		zaplog.SUG.Debugln("Context timeout, shutting down...")
	}

	must.Done(client.Offline(context.Background()))
	must.Done(client.DeregisterService())

	zaplog.SUG.Debugln("Returning from main(), exiting...")
}
```

⬆️ **源码:** [源码](internal/demos/demo1x/main.go)

## 配置

### Nacos 配置

```go
config := &nacosv2.Config{
	Endpoint:  "127.0.0.1:8848",  // Nacos 服务器地址
	AppName:   "my-service",       // 服务名称
	Address:   "0.0.0.0:8080",     // 服务绑定地址
	Group:     "DEFAULT_GROUP",    // 服务分组
	Namespace: "public",           // 命名空间 ID
}
```

### 客户端选项

```go
clientOptions := []constant.ClientOption{
	constant.WithCacheDir("/tmp/nacos/cache"),
	constant.WithLogDir("/tmp/nacos/log"),
	constant.WithLogLevel("info"),
	constant.WithNotLoadCacheAtStart(true),
}
```

## API 参考

### 核心方法

- `NewNacosClient(config, options, zapLog)` - 创建 Nacos 客户端，自动检测服务信息
- `RegisterService()` - 向 Nacos 注册服务实例
- `DeregisterService()` - 从 Nacos 移除服务实例
- `Online(ctx)` - 通过重新注册使服务上线
- `Offline(ctx)` - 通过注销使服务下线
- `GetServiceInstance(ctx, serviceName)` - 发现健康的服务实例
- `GetConfig(ctx, dataID)` - 从配置中心获取配置

## 高级特性

### 自动 IP 检测

客户端从允许的网络接口检测服务 IP：
- macOS: `en0` (以太网/Wi-Fi)
- Linux: `eth0` (以太网)
- VMware: `ens224`
- AWS EC2: `ens5`

当绑定地址配置为 `0.0.0.0` 时自动检测。

### 健康检查管理

所有注册的实例都是临时实例并启用健康检查，确保：
- 服务崩溃时自动注销
- 实时实例健康状态
- 基于健康实例的负载均衡

### 命名空间隔离

使用命名空间分隔不同环境：

```go
// 生产环境
prodConfig := &nacosv2.Config{
	Namespace: "prod-namespace-id",
	// ...
}

// 测试环境
testConfig := &nacosv2.Config{
	Namespace: "test-namespace-id",
	// ...
}
```

## 示例

### 服务注册

**注册服务实例：**
```go
client := rese.P1(nacosv2.NewNacosClient(config, clientOptions, zapLog))
rese.V0(client.RegisterService())
```

**关闭时注销：**
```go
defer rese.V0(client.DeregisterService())
```

### 服务发现

**获取健康的服务实例：**
```go
instance := rese.P1(client.GetServiceInstance(context.Background(), "service-name"))
fmt.Printf("实例: %s:%d\n", instance.Ip, instance.Port)
```

### 配置管理

**获取配置：**
```go
configData := rese.P1(client.GetConfig(context.Background(), "database-config"))
fmt.Println("配置:", configData)
```

### 上下线控制

**使服务上线：**
```go
rese.V0(client.Online(context.Background()))
```

**使服务下线（不停止服务）：**
```go
rese.V0(client.Offline(context.Background()))
```

### 多环境配置

**生产环境配置：**
```go
prodConfig := &nacosv2.Config{
	Endpoint:  "nacos-prod.company.com:8848",
	AppName:   "payment-service",
	Namespace: "prod-namespace-id",
	Group:     "PROD_GROUP",
}
```

**测试环境配置：**
```go
testConfig := &nacosv2.Config{
	Endpoint:  "nacos-test.company.com:8848",
	AppName:   "payment-service",
	Namespace: "test-namespace-id",
	Group:     "TEST_GROUP",
}
```

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 许可证类型

MIT 许可证。详见 [LICENSE](LICENSE)。

---

## 🤝 项目贡献

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **发现问题？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **功能建议？** 创建 issue 讨论您的想法
- 📖 **文档疑惑？** 报告问题，帮助我们改进文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，帮助我们优化性能
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：为面向用户的更改更新文档，并使用有意义的提交消息
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来为此项目做出贡献。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/go-xlan/go-nacos-v2.svg?variant=adaptive)](https://starchart.cc/go-xlan/go-nacos-v2)