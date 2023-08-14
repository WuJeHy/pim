package pim_server

import (
	"github.com/spf13/viper"
	"log"
)

type ImServerConfig struct {
	LoggerPath    string `json:"logger_path"    yaml:"logger_path"    mapstructure:"logger_path"`
	LoggerLevel   string `json:"logger_level"   yaml:"logger_level"   mapstructure:"logger_level"`
	RedisIP       string `json:"redis_ip"       yaml:"redis_ip"       mapstructure:"redis_ip"`
	RedisDB       int    `json:"redis_db"       yaml:"redis_db"       mapstructure:"redis_db"`
	RedisPassword string `json:"redis_password" yaml:"redis_password" mapstructure:"redis_password"`
	DBUri         string `json:"db_uri"         yaml:"db_uri"         mapstructure:"db_uri"`
	HttpPort      int    `json:"port"           yaml:"port"           mapstructure:"port"`
	RpcPort       int    `json:"rpc_port"           yaml:"rpc_port"           mapstructure:"rpc_port"`
	Workspace     string `json:"workspace" yaml:"workspace" mapstructure:"workspace"`
}

func NewConfig(path string) (*ImServerConfig, error) {
	v := viper.New()
	v.AddConfigPath(".")
	//v.SetConfigType("yaml")
	v.SetConfigFile(path)
	//v.SetConfigName("config")
	err := v.ReadInConfig()
	if err != nil {
		return nil, err

	}
	config := new(ImServerConfig)

	err = v.UnmarshalKey("pim", config)

	if err != nil {
		//log.Fatalln("配置文件错误 ", err)
		return nil, err
	}
	// 默认值设置
	if config.LoggerPath == "" {
		config.LoggerPath = "./logs"
	}

	if config.LoggerLevel == "" {
		config.LoggerLevel = "info"
	}
	if config.HttpPort == 0 {
		config.HttpPort = 10004
	}

	if config.RedisIP == "" {
		config.RedisIP = "127.0.0.1:6379"
	}

	if config.DBUri == "" {
		log.Fatalln("db_uri 未配置")
	}

	return config, nil
}
