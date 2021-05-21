package lc

import (
	"fmt"
	"github.com/spf13/viper"
	"math"
)

type YamlConfig struct {
	Viper *viper.Viper
}

func NewYamlConfig() *YamlConfig {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading configuration file: %v", err)
		panic("Cannot read configuration file.")
	} else {
		fmt.Println("Config file read.")
		myviper := &YamlConfig{
			Viper: viper.GetViper(),
		}
		return myviper
	}
}

func (y *YamlConfig) SetupEdgeConfig(id *string) *Config {
	port := y.Viper.GetString("edge_nodes." + *id + ".port")
	host := y.Viper.GetString("edge_nodes." + *id + ".host")
	tehost := y.Viper.GetString("te.host")
	teport := y.Viper.GetString("te.port")
	epochDuration := y.Viper.GetInt("epoch.duration")
	epochMaxSize := y.Viper.GetInt("epoch.maxsize")
	nodeCount := len(y.Viper.GetStringMap("edge_nodes"))
	config := NewConfig("E_" + *id)
	config.TEAddr = fmt.Sprintf("%v:%v", tehost, teport)
	config.Node.Port = port
	config.Node.Ip = host
	config.Node.Uuid = *id
	config.F = int(math.Ceil(float64(nodeCount) / 2))
	config.Epoch.Duration = epochDuration
	config.Epoch.MaxSize = epochMaxSize
	config.ClusterNodes = y.GetAllClusterNodes(id)
	return config
}

func (y *YamlConfig) GetAllClusterNodes(exid *string) []NodeInfo {
	edgeNodes := y.Viper.GetStringMap("edge_nodes")
	nodes, i := make([]NodeInfo, len(edgeNodes)-1), 0
	fmt.Println(edgeNodes)

	for k := range edgeNodes {
		if k != *exid {
			n := NodeInfo{
				Ip:   y.Viper.GetString("edge_nodes." + k + ".host"),
				Port: y.Viper.GetString("edge_nodes." + k + ".port"),
				Uuid: k,
			}
			nodes[i] = n
			i += 1
		}
	}
	fmt.Printf("Cluster nodes (excluding self): %v", nodes)
	return nodes

}
