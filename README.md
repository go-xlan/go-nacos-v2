[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/go-xlan/go-nacos-v2/release.yml?branch=main&label=BUILD)](https://github.com/go-xlan/go-nacos-v2/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/go-xlan/go-nacos-v2)](https://pkg.go.dev/github.com/go-xlan/go-nacos-v2)
[![Coverage Status](https://img.shields.io/coveralls/github/go-xlan/go-nacos-v2/main.svg)](https://coveralls.io/github/go-xlan/go-nacos-v2?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.23+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/go-xlan/go-nacos-v2.svg)](https://github.com/go-xlan/go-nacos-v2/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-xlan/go-nacos-v2)](https://goreportcard.com/report/github.com/go-xlan/go-nacos-v2)

# go-nacos-v2

Nacos v2 SDK integration client with service registration and configuration management.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->
## CHINESE README

[中文说明](README.zh.md)
<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Core Features

🎯 **Service Registration**: Auto IP detection with ephemeral instance registration
⚡ **Service Discovery**: Health-check based instance selection with group support
🔄 **Config Management**: Dynamic configuration retrieval from Nacos config center
🌍 **Online/Offline Control**: Graceful service state management without restart
📋 **Multi-Environment**: Namespace and group-based service isolation

## Installation

```bash
go get github.com/go-xlan/go-nacos-v2
```

## Quick Start

### Complete Service Lifecycle Example

This example demonstrates complete service registration, discovery, and graceful shutdown handling.

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

⬆️ **Source:** [Source](internal/demos/demo1x/main.go)

## Configuration

### Nacos Config

```go
config := &nacosv2.Config{
	Endpoint:  "127.0.0.1:8848",  // Nacos server address
	AppName:   "my-service",       // Service name
	Address:   "0.0.0.0:8080",     // Service bind address
	Group:     "DEFAULT_GROUP",    // Service group
	Namespace: "public",           // Namespace ID
}
```

### Client Options

```go
clientOptions := []constant.ClientOption{
	constant.WithCacheDir("/tmp/nacos/cache"),
	constant.WithLogDir("/tmp/nacos/log"),
	constant.WithLogLevel("info"),
	constant.WithNotLoadCacheAtStart(true),
}
```

## API Reference

### Core Methods

- `NewNacosClient(config, options, zapLog)` - Create Nacos client with auto service info detection
- `RegisterService()` - Register service instance to Nacos
- `DeregisterService()` - Remove service instance from Nacos
- `Online(ctx)` - Bring service online by re-registering
- `Offline(ctx)` - Take service offline by deregistering
- `GetServiceInstance(ctx, serviceName)` - Discover healthy service instance
- `GetConfig(ctx, dataID)` - Retrieve configuration from config center

## Advanced Features

### Auto IP Detection

The client detects service IP from allowed network interfaces:
- macOS: `en0` (Ethernet/Wi-Fi)
- Linux: `eth0` (Ethernet)
- VMware: `ens224`
- AWS EC2: `ens5`

Falls back to `0.0.0.0` detection when bind address is configured.

### Health Check Management

All registered instances are ephemeral with health checks enabled, ensuring:
- Auto deregistration on service crash
- Real-time instance health status
- Load balancing based on healthy instances

### Namespace Isolation

Use namespaces to separate environments:

```go
// Production environment
prodConfig := &nacosv2.Config{
	Namespace: "prod-namespace-id",
	// ...
}

// Testing environment
testConfig := &nacosv2.Config{
	Namespace: "test-namespace-id",
	// ...
}
```

## Examples

### Service Registration

**Register service instance:**
```go
client := rese.P1(nacosv2.NewNacosClient(config, clientOptions, zapLog))
rese.V0(client.RegisterService())
```

**Deregister on shutdown:**
```go
defer rese.V0(client.DeregisterService())
```

### Service Discovery

**Get healthy service instance:**
```go
instance := rese.P1(client.GetServiceInstance(context.Background(), "service-name"))
fmt.Printf("Instance: %s:%d\n", instance.Ip, instance.Port)
```

### Configuration Management

**Retrieve configuration:**
```go
configData := rese.P1(client.GetConfig(context.Background(), "database-config"))
fmt.Println("Config:", configData)
```

### Online/Offline Control

**Bring service online:**
```go
rese.V0(client.Online(context.Background()))
```

**Take service offline (without stopping):**
```go
rese.V0(client.Offline(context.Background()))
```

### Multi-Environment Setup

**Production configuration:**
```go
prodConfig := &nacosv2.Config{
	Endpoint:  "nacos-prod.company.com:8848",
	AppName:   "payment-service",
	Namespace: "prod-namespace-id",
	Group:     "PROD_GROUP",
}
```

**Testing configuration:**
```go
testConfig := &nacosv2.Config{
	Endpoint:  "nacos-test.company.com:8848",
	AppName:   "payment-service",
	Namespace: "test-namespace-id",
	Group:     "TEST_GROUP",
}
```

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-09-26 07:39:27.188023 +0000 UTC -->

## 📄 License

MIT License. See [LICENSE](LICENSE).

---

## 🤝 Contributing

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Found a mistake?** Open an issue on GitHub with reproduction steps
- 💡 **Have a feature idea?** Create an issue to discuss the suggestion
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes and use significant commit messages
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/go-xlan/go-nacos-v2.svg?variant=adaptive)](https://starchart.cc/go-xlan/go-nacos-v2)