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
	UserID              string `json:"user_id"`
	PublicKey           string `json:"public_key"`
	PrivateKey          string `json:"private_key"`
	IP                  string `json:"ip"`
	AllowedIPs          string `json:"allowed_ips"`
	AdvertiseRoutes     string `json:"advertise_routes"`
	AcceptRoutes        string `json:"accept_routes"`
	Endpoint            string `json:"endpoint"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	PreUp               string `json:"pre_up"`
	PostUp              string `json:"post_up"`
	PreDown             string `json:"pre_down"`
	PostDown            string `json:"post_down"`
}

type UserTrafficData struct {
	UserID        string
	IP            string
	ReceiveBytes  uint64
	TransmitBytes uint64
}
