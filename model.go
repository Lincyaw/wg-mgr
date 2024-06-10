package main

type ServerConfig struct {
	ServerIP   string `yaml:"server_ip"`
	Port       int    `yaml:"port"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
	IP         string `yaml:"ip"`
	DNS        string `yaml:"dns"`
	Table      string `yaml:"table"`
	MTU        int    `yaml:"mtu"`
	PreUp      string `yaml:"pre_up"`
	PostUp     string `yaml:"post_up"`
	PreDown    string `yaml:"pre_down"`
	PostDown   string `yaml:"post_down"`
	IPPool     string `yaml:"ip_pool"`
}

type UserConfig struct {
	UserID              string
	PublicKey           string
	PrivateKey          string
	IP                  string
	AllowedIPs          string
	Endpoint            string
	PersistentKeepalive int
}
