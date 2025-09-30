package nacosv2

import (
	"os"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/yyle88/must"
	"github.com/yyle88/rese"
	"github.com/yyle88/zaplog"
	"go.uber.org/zap"
)

// MustNewNacosClientParam creates Nacos client parameters with endpoint parsing
// Builds complete client config from user config and additional options
// Panics if endpoint is missing or invalid (use in initialization)
//
// MustNewNacosClientParam 创建带端点解析的 Nacos 客户端参数
// 从用户配置和附加选项构建完整的客户端配置
// 端点缺失或无效时会 panic（用于初始化阶段）
func MustNewNacosClientParam(config *Config, nacosOptions []constant.ClientOption, zapLog *zaplog.Zap) vo.NacosClientParam {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = must.Nice(os.Getenv("NACOS_ADDR"))
	}
	zapLog.LOG.Debug("nacos", zap.String("endpoint", endpoint))

	endpointParts := strings.Split(endpoint, ":")
	nacosHost := endpointParts[0]
	nacosPort := rese.C1(strconv.Atoi(endpointParts[1]))

	var opts = []constant.ClientOption{
		constant.WithEndpoint(endpoint),
		constant.WithAppName(config.AppName),
		constant.WithNamespaceId(config.Namespace),
	}

	opts = append(opts, nacosOptions...)

	clientConfig := constant.NewClientConfig(opts...)

	serverConfig := constant.NewServerConfig(nacosHost, uint64(nacosPort))

	param := vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: []constant.ServerConfig{*serverConfig},
	}
	return param
}
