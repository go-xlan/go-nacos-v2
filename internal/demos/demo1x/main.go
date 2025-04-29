package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/go-xlan/go-nacos-v2/nacosv2"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
)

func main() {
	// 配置 Nacos 客户端
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

	// 注册服务
	must.Done(client.RegisterService())

	// 上线服务
	client.Online(context.Background())

	// 获取服务实例
	instance := rese.P1(client.GetServiceInstance(context.Background(), "demo1x"))
	zaplog.SUG.Debugln(neatjsons.S(instance))

	// 创建带取消功能的 context
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute)
	defer cancelFunc()

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt) // 捕获 Ctrl+C (SIGINT)

	// 等待信号或上下文取消
	select {
	case <-sigCh:
		zaplog.SUG.Debugln("Received Ctrl+C, shutting down...")
	case <-ctx.Done():
		zaplog.SUG.Debugln("Context timeout, shutting down...")
	}

	// 清理逻辑
	client.Offline(context.Background())
	must.Done(client.DeregisterService())

	zaplog.SUG.Debugln("Returning from main(), exiting...")
}
