package lib

type BackendServerConfig struct {
	Host string `json:"host"`
	Url  string `json:"url"`
}

type Config struct {
	BackendConfig []BackendServerConfig `json:"backend"`
}
