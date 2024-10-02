package lib

import (
	"github.com/spf13/viper"
)

type BackendServerConfig struct {
	Host string `mapsturcture:"host"`
	Url  string `mapstructure:"url"`
}

type Config struct {
	BackendConfig []BackendServerConfig `mapstructure:"backend"`
}

// reads dir name and finds config file and reads the config according
// to the extension.
func ReadConfig(filename string) (Config, error) {
	viper.SetConfigFile(filename)
	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}
