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
	"github.com/pkg/errors"
	"github.com/yyle88/erero"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint  string //NACOS_ADDR (example: 127.0.0.1:8848)
	AppName   string //service name
	Address   string //service address (example: 0.0.0.0:8080)
	Group     string //group name (default: DEFAULT_GROUP)
	Namespace string //namespace-ID
}

type NacosClient struct {
	config            *Config
	NacosConfigClient config_client.IConfigClient
	NacosNamingClient naming_client.INamingClient
	serviceIp         string
	portNum           int
	zapLog            *zaplog.Zap
}

func NewNacosClient(config *Config, nacosOptions []constant.ClientOption, zapLog *zaplog.Zap) (*NacosClient, error) {
	serviceIp := utils.MustGetIpV4(config.Address)
	if serviceIp == "0.0.0.0" {
		serviceIp = rese.C1(utils.GetIpv4())
	}
	port := utils.MustGetPort(config.Address)
	portNum := rese.C1(strconv.Atoi(port))

	nacosParam := MustNewNacosClientParam(config, nacosOptions, zapLog)

	nacosNamingClient, err := clients.NewNamingClient(nacosParam)
	if err != nil {
		return nil, errors.WithMessage(err, "wrong to create nacos naming client")
	}
	nacosConfigClient, err := clients.NewConfigClient(nacosParam)
	if err != nil {
		return nil, errors.WithMessage(err, "wrong to create nacos config client")
	}

	return &NacosClient{
		config:            config,
		NacosNamingClient: nacosNamingClient,
		NacosConfigClient: nacosConfigClient,
		serviceIp:         serviceIp,
		portNum:           portNum,
		zapLog:            zapLog,
	}, nil
}

func (uc *NacosClient) RegisterService() error {
	if _, err := uc.NacosNamingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          uc.serviceIp,
		Port:        uint64(uc.portNum),
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

func (uc *NacosClient) DeregisterService() error {
	if _, err := uc.NacosNamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          uc.serviceIp,
		Port:        uint64(uc.portNum),
		ServiceName: uc.config.AppName,
		Ephemeral:   true,
	}); err != nil {
		return erero.Wro(err)
	}
	return nil
}

func (uc *NacosClient) Online(ctx context.Context) {
	allService, err := uc.NacosNamingClient.GetService(vo.GetServiceParam{
		ServiceName: uc.config.AppName,
	})
	if err != nil {
		uc.zapLog.LOG.Error("Nacos Service GetService wrong", zap.Error(err))
		return
	}
	for _, host := range allService.Hosts {
		if host.Ip == uc.serviceIp {
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
				return
			}
		}
	}
}

func (uc *NacosClient) Offline(ctx context.Context) {
	allService, err := uc.NacosNamingClient.GetService(vo.GetServiceParam{
		ServiceName: uc.config.AppName,
	})
	if err != nil {
		uc.zapLog.LOG.Error("Nacos Service GetService wrong", zap.Error(err))
		return
	}
	for _, host := range allService.Hosts {
		if host.Ip == uc.serviceIp {
			_, err := uc.NacosNamingClient.DeregisterInstance(vo.DeregisterInstanceParam{
				Ip:          host.Ip,
				Port:        host.Port,
				ServiceName: uc.config.AppName,
			})
			if err != nil {
				uc.zapLog.LOG.Error("Nacos Service Offline wrong", zap.Error(err))
				return
			}
		}
	}
}

func (uc *NacosClient) GetConfig(ctx context.Context, dataID string) (string, error) {
	return uc.NacosConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  uc.config.Group,
	})
}

// GetServiceInstance 获取指定服务的实例
func (uc *NacosClient) GetServiceInstance(ctx context.Context, serviceName string) (*model.Instance, error) {
	nacosNamingClient := uc.NacosNamingClient

	// 设置服务发现参数
	param := vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   uc.config.Group, // 使用默认组
		Clusters:    []string{},      // 如果有特定集群需求可以在这里指定
	}

	// 调用 Nacos 的服务发现接口获取健康的实例
	instance, err := nacosNamingClient.SelectOneHealthyInstance(param)
	if err != nil {
		uc.zapLog.LOG.Error("failed to discover service instance",
			zap.String("serviceName", serviceName),
			zap.Error(err))
		return nil, erero.Errorf("failed to get healthy instance for service %s: %v", serviceName, err)
	}

	// 记录发现的服务实例信息
	uc.zapLog.LOG.Info("service instance discovered",
		zap.String("serviceName", serviceName),
		zap.String("ip", instance.Ip),
		zap.Uint64("port", instance.Port),
		zap.String("instanceId", instance.InstanceId))

	return instance, nil
}
