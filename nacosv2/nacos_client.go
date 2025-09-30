// Package nacosv2: Nacos v2 SDK integration client with service registration and config management
// Provides unified Nacos operations including service discovery, registration, and configuration access
// Supports auto IP detection, health check management, and seamless online/offline operations
//
// nacosv2: Nacos v2 SDK 集成客户端，支持服务注册和配置管理
// 提供统一的 Nacos 操作，包括服务发现、注册和配置访问
// 支持自动 IP 检测、健康检查管理和无缝上线/下线操作
package nacosv2

import (
	"context"
	"strconv"

	"github.com/go-xlan/go-nacos-v2/internal/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/yyle88/erero"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

// Config represents Nacos client configuration settings
// Contains essential connection and service identification info
//
// Config 代表 Nacos 客户端配置设置
// 包含基本的连接和服务识别信息
type Config struct {
	Endpoint  string // Nacos server address (example: 127.0.0.1:8848) // Nacos 服务器地址
	AppName   string // Service application name // 服务应用名称
	Address   string // Service bind address (example: 0.0.0.0:8080) // 服务绑定地址
	Group     string // Service group name (default: DEFAULT_GROUP) // 服务组名称
	Namespace string // Nacos namespace identifier // Nacos 命名空间标识符
}

// NacosClient wraps Nacos SDK clients with additional service management
// Manages both config and naming operations with automatic service info tracking
//
// NacosClient 封装 Nacos SDK 客户端，提供额外的服务管理功能
// 管理配置和命名操作，自动跟踪服务信息
type NacosClient struct {
	config            *Config                     // Client configuration settings // 客户端配置设置
	NacosConfigClient config_client.IConfigClient // Nacos config operations client // Nacos 配置操作客户端
	NacosNamingClient naming_client.INamingClient // Nacos naming operations client // Nacos 命名操作客户端
	serviceHost       string                      // Service host IP address // 服务主机 IP 地址
	servicePort       int                         // Service port number // 服务端口号
	zapLog            *zaplog.Zap                 // Zap logging instance // Zap 日志实例
}

// NewNacosClient creates new Nacos client with auto service info detection
// Initializes both config and naming clients with given configuration
// Returns client instance and error if initialization fails
//
// NewNacosClient 创建新的 Nacos 客户端，自动检测服务信息
// 使用给定配置初始化配置和命名客户端
// 返回客户端实例，初始化失败时返回错误
func NewNacosClient(config *Config, nacosOptions []constant.ClientOption, zapLog *zaplog.Zap) (*NacosClient, error) {
	serviceHost := utils.MustGetIPv4(config.Address)
	if serviceHost == "0.0.0.0" {
		serviceHost = rese.C1(utils.GetIPv4())
	}
	port := utils.MustGetPort(config.Address)
	servicePort := rese.C1(strconv.Atoi(port))

	clientParam := MustNewNacosClientParam(config, nacosOptions, zapLog)

	namingClient, err := clients.NewNamingClient(clientParam)
	if err != nil {
		return nil, erero.WithMessage(err, "wrong to create nacos naming client")
	}
	configClient, err := clients.NewConfigClient(clientParam)
	if err != nil {
		return nil, erero.WithMessage(err, "wrong to create nacos config client")
	}

	return &NacosClient{
		config:            config,
		NacosNamingClient: namingClient,
		NacosConfigClient: configClient,
		serviceHost:       serviceHost,
		servicePort:       servicePort,
		zapLog:            zapLog,
	}, nil
}

// RegisterService registers current service instance to Nacos naming server
// Creates ephemeral instance registration with health check enabled
// Returns error if registration fails
//
// RegisterService 将当前服务实例注册到 Nacos 命名服务器
// 创建带健康检查的临时实例注册
// 注册失败时返回错误
func (uc *NacosClient) RegisterService() error {
	if _, err := uc.NacosNamingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          uc.serviceHost,
		Port:        uint64(uc.servicePort),
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		ServiceName: uc.config.AppName,
		Metadata:    map[string]string{"preserved.register.source": "golang-gin"},
		Ephemeral:   true,
	}); err != nil {
		return erero.Wro(err)
	}
	return nil
}

// DeregisterService removes current service instance from Nacos naming server
// Cleans up service registration when shutting down
// Returns error if deregistration fails
//
// DeregisterService 从 Nacos 命名服务器移除当前服务实例
// 关闭时清理服务注册
// 注销失败时返回错误
func (uc *NacosClient) DeregisterService() error {
	if _, err := uc.NacosNamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          uc.serviceHost,
		Port:        uint64(uc.servicePort),
		ServiceName: uc.config.AppName,
		Ephemeral:   true,
	}); err != nil {
		return erero.Wro(err)
	}
	return nil
}

// Online brings service instance online by re-registering to Nacos
// Finds matching instance and updates registration status to healthy
// Returns error if operation fails
//
// Online 通过重新注册到 Nacos 使服务实例上线
// 查找匹配实例并更新注册状态为健康
// 操作失败时返回错误
func (uc *NacosClient) Online(ctx context.Context) error {
	service, err := uc.NacosNamingClient.GetService(vo.GetServiceParam{
		ServiceName: uc.config.AppName,
	})
	if err != nil {
		uc.zapLog.LOG.Error("Nacos Service GetService wrong", zap.Error(err))
		return erero.Wro(err)
	}
	for _, host := range service.Hosts {
		if host.Ip == uc.serviceHost {
			_, err := uc.NacosNamingClient.RegisterInstance(vo.RegisterInstanceParam{
				Ip:          host.Ip,
				Port:        host.Port,
				Weight:      1,
				Enable:      true,
				Healthy:     true,
				ServiceName: uc.config.AppName,
				Metadata:    map[string]string{"preserved.register.source": "golang-gin"},
				Ephemeral:   true,
			})
			if err != nil {
				uc.zapLog.LOG.Error("Nacos Service Online wrong", zap.Error(err))
				return erero.Wro(err)
			}
		}
	}
	return nil
}

// Offline takes service instance offline by deregistering from Nacos
// Finds matching instance and removes registration without shutdown
// Returns error if operation fails
//
// Offline 通过从 Nacos 注销使服务实例下线
// 查找匹配实例并移除注册而不关闭服务
// 操作失败时返回错误
func (uc *NacosClient) Offline(ctx context.Context) error {
	service, err := uc.NacosNamingClient.GetService(vo.GetServiceParam{
		ServiceName: uc.config.AppName,
	})
	if err != nil {
		uc.zapLog.LOG.Error("Nacos Service GetService wrong", zap.Error(err))
		return erero.Wro(err)
	}
	for _, host := range service.Hosts {
		if host.Ip == uc.serviceHost {
			_, err := uc.NacosNamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
				Ip:          host.Ip,
				Port:        host.Port,
				ServiceName: uc.config.AppName,
			})
			if err != nil {
				uc.zapLog.LOG.Error("Nacos Service Offline wrong", zap.Error(err))
				return erero.Wro(err)
			}
		}
	}
	return nil
}

// GetConfig retrieves configuration content from Nacos config center
// Fetches config data using specified data ID and configured group
// Returns config content string and error if fetch fails
//
// GetConfig 从 Nacos 配置中心检索配置内容
// 使用指定的数据 ID 和配置的组获取配置数据
// 返回配置内容字符串，获取失败时返回错误
func (uc *NacosClient) GetConfig(ctx context.Context, dataID string) (string, error) {
	return uc.NacosConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  uc.config.Group,
	})
}

// GetServiceInstance discovers and returns one healthy instance of specified service
// Selects instance from configured group with health check validation
// Returns instance info and error if discovery fails
//
// GetServiceInstance 发现并返回指定服务的一个健康实例
// 从配置的组中选择实例并进行健康检查验证
// 返回实例信息，发现失败时返回错误
func (uc *NacosClient) GetServiceInstance(ctx context.Context, serviceName string) (*model.Instance, error) {
	namingClient := uc.NacosNamingClient

	// Setup service discovery parameters // 设置服务发现参数
	param := vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   uc.config.Group, // Use configured group // 使用配置的组
		Clusters:    []string{},      // Specify clusters if needed // 需要时指定集群
	}

	// Call Nacos service discovery to get healthy instance // 调用 Nacos 服务发现获取健康实例
	instance, err := namingClient.SelectOneHealthyInstance(param)
	if err != nil {
		uc.zapLog.LOG.Error("unable to discover service instance",
			zap.String("serviceName", serviceName),
			zap.Error(err))
		return nil, erero.Errorf("cannot get healthy instance of service %s: %v", serviceName, err)
	}

	// Log discovered service instance info // 记录发现的服务实例信息
	uc.zapLog.LOG.Info("service instance discovered",
		zap.String("serviceName", serviceName),
		zap.String("ip", instance.Ip),
		zap.Uint64("port", instance.Port),
		zap.String("instanceId", instance.InstanceId))

	return instance, nil
}
