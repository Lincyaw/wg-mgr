package main

import (
	"fmt"
	"strings"
	"time"
)

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
	LastHandShake time.Time
}
type UserTrafficList []UserTrafficData

// formatBytes 将字节数转换为合适的单位
func formatBytes(bytes uint64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// colorize 为字符串添加颜色
func colorize(text, colorCode string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
}

// String 格式化输出 UserTrafficList 为表格形式
func (data UserTrafficList) String() string {
	// 表头
	header := fmt.Sprintf("%-15s | %-15s | %-15s | %-15s | %-20s", "UserID", "IP", "ReceiveBytes", "TransmitBytes", "LastHandshake")
	divider := strings.Repeat("-", len(header))
	var rows []string
	rows = append(rows, header)
	rows = append(rows, divider)

	// 数据行
	for _, d := range data {
		receiveBytes := colorize(formatBytes(d.ReceiveBytes), "32")   // 绿色
		transmitBytes := colorize(formatBytes(d.TransmitBytes), "34") // 蓝色
		lastHandshake := d.LastHandShake.Format("2006-01-02 15:04:05")
		row := fmt.Sprintf("%-15s | %-15s | %-15s | %-15s | %-20s",
			d.UserID, d.IP, receiveBytes, transmitBytes, lastHandshake)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}
