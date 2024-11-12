package main

import (
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"log"
	"net"
	"os/exec"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserManager struct {
	db *gorm.DB
}

func NewUserManager(dbPath string) (*UserManager, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	um := &UserManager{db: db}
	err = um.createTable()
	if err != nil {
		return nil, err
	}

	return um, nil
}

func (um *UserManager) createTable() error {
	return um.db.AutoMigrate(&User{})
}

func (um *UserManager) AddUser(user *User) error {
	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		log.Fatal(err)
	}
	ipPoolCIDR := serverConfig.IPPool

	ipPool, err := generateIPPool(ipPoolCIDR)
	if err != nil {
		return err
	}

	var usedIPs []string
	err = um.db.Model(&User{}).Pluck("ip", &usedIPs).Error
	if err != nil {
		return err
	}

	var newIP string
	for _, ip := range ipPool {
		used := false
		for _, usedIP := range usedIPs {
			if ip == usedIP {
				used = true
				break
			}
		}
		if !used {
			newIP = ip
			break
		}
	}
	if newIP == "" {
		return errors.New("no available IP addresses")
	}

	privateKey, publicKey, err := generateKeys()
	if err != nil {
		return err
	}
	user.PrivateKey = privateKey
	user.PublicKey = publicKey
	if user.AllowedIPs == "" {
		user.AllowedIPs = newIP + "/24"
	}

	if user.AdvertiseRoutes != "" {
		routes, _ := um.GetAllRoutes()
		for _, v := range routes {
			if v == user.AdvertiseRoutes {
				return errors.New("advertise route already exists")
			}
		}
	}

	user.IP = newIP
	err = um.db.Create(user).Error
	return err
}

func (um *UserManager) GetAllUsers() ([]User, error) {
	var users []User
	err := um.db.Find(&users).Error
	return users, err
}

func (um *UserManager) GetAllRoutes() ([]string, error) {
	var routes []string
	err := um.db.Model(&User{}).Where("advertise_routes != ''").Pluck("advertise_routes", &routes).Error
	return routes, err
}

func (um *UserManager) UpdateUser(user User) error {
	err := um.db.Model(&User{}).Where("user_id = ?", user.UserID).Updates(user).Error
	return err
}

func (um *UserManager) UpdateUserEndpoints(serverConfig ServerConfig) error {
	endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)

	var users []User
	err := um.db.Find(&users).Error
	if err != nil {
		return err
	}

	for _, user := range users {
		err := um.db.Model(&User{}).Where("user_id = ?", user.UserID).Update("endpoint", endpoint).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (um *UserManager) DeleteUser(userID string) error {
	err := um.db.Where("user_id = ?", userID).Delete(&User{}).Error
	return err
}

// GenerateServerConfig generate server config
func (um *UserManager) GenerateServerConfig(serverConfig ServerConfig) (string, error) {
	users, err := um.GetAllUsers()
	if err != nil {
		return "", err
	}

	var configBuilder strings.Builder
	configBuilder.WriteString(fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
`, serverConfig.PrivateKey, serverConfig.IP, serverConfig.Port))

	if serverConfig.DNS != "" {
		configBuilder.WriteString(fmt.Sprintf("DNS = %s\n", serverConfig.DNS))
	}
	if serverConfig.Table != "" {
		configBuilder.WriteString(fmt.Sprintf("Table = %s\n", serverConfig.Table))
	}
	if serverConfig.MTU != 0 {
		configBuilder.WriteString(fmt.Sprintf("MTU = %d\n", serverConfig.MTU))
	}
	if serverConfig.PreUp != "" {
		configBuilder.WriteString(fmt.Sprintf("PreUp = %s\n", serverConfig.PreUp))
	}
	if serverConfig.PostUp != "" {
		configBuilder.WriteString(fmt.Sprintf("PostUp = %s\n", serverConfig.PostUp))
	}
	if serverConfig.PreDown != "" {
		configBuilder.WriteString(fmt.Sprintf("PreDown = %s\n", serverConfig.PreDown))
	}
	if serverConfig.PostDown != "" {
		configBuilder.WriteString(fmt.Sprintf("PostDown = %s\n", serverConfig.PostDown))
	}

	for _, user := range users {
		base := fmt.Sprintf(`[Peer]
PublicKey = %s
`, user.PublicKey)
		if user.AdvertiseRoutes != "" {
			base += fmt.Sprintf(`AllowedIPs = %s, %s 
`, user.IP+"/32", user.AdvertiseRoutes)
		} else {
			base += fmt.Sprintf(`AllowedIPs = %s 
`, user.IP+"/32")
		}
		configBuilder.WriteString(base)
	}

	return configBuilder.String(), nil
}

// GetAllUserTraffic 获取所有用户的流量数据
func (um *UserManager) GetAllUserTraffic() (UserTrafficList, error) {
	// 创建 wgctrl 客户端
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("无法创建 WireGuard 控制器: %w", err)
	}
	defer client.Close()

	// 获取系统上所有 WireGuard 接口信息
	devices, err := client.Devices()
	if err != nil {
		return nil, fmt.Errorf("无法获取 WireGuard 设备信息: %w", err)
	}

	// 获取所有用户信息
	users, err := um.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户信息: %w", err)
	}

	// 将用户 IP 与流量数据关联
	var trafficData []UserTrafficData
	for _, user := range users {
		for _, device := range devices {
			for _, peer := range device.Peers {

				// 检查用户 IP 是否匹配 Peer 的 AllowedIPs
				for _, allowedIP := range peer.AllowedIPs {
					if allowedIP.String() == user.IP+"/32" {
						trafficData = append(trafficData, UserTrafficData{
							UserID:        user.UserID,
							IP:            user.IP,
							ReceiveBytes:  uint64(peer.ReceiveBytes),
							TransmitBytes: uint64(peer.TransmitBytes),
							LastHandShake: peer.LastHandshakeTime,
						})
					}
				}
			}
		}
	}

	return trafficData, nil
}

// 生成密钥对
func generateKeys() (string, string, error) {
	privateKeyCmd := exec.Command("wg", "genkey")
	privateKeyOut, err := privateKeyCmd.Output()
	if err != nil {
		return "", "", err
	}
	privateKey := strings.TrimSpace(string(privateKeyOut))

	publicKeyCmd := exec.Command("wg", "pubkey")
	publicKeyCmd.Stdin = strings.NewReader(privateKey)
	publicKeyOut, err := publicKeyCmd.Output()
	if err != nil {
		return "", "", err
	}
	publicKey := strings.TrimSpace(string(publicKeyOut))

	return privateKey, publicKey, nil
}

// generate user config
func generateUserConfig(serverConfig ServerConfig, user User) string {
	var configBuilder strings.Builder

	configBuilder.WriteString(fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
`, user.PrivateKey, user.IP))

	if user.PreUp != "" {
		configBuilder.WriteString(fmt.Sprintf(`PreUp = %s
`, user.PreUp))
	}
	if user.PostUp != "" {
		configBuilder.WriteString(fmt.Sprintf(`PostUp = %s
`, user.PostUp))
	}
	if user.PreDown != "" {
		configBuilder.WriteString(fmt.Sprintf(`PreDown = %s
`, user.PreDown))
	}
	if user.PostDown != "" {
		configBuilder.WriteString(fmt.Sprintf(`PostDown = %s
`, user.PostDown))
	}
	configBuilder.WriteString(fmt.Sprintf(`
[Peer]
PublicKey = %s
`, serverConfig.PublicKey))
	if user.AcceptRoutes != "" {
		configBuilder.WriteString(fmt.Sprintf(`AllowedIPs = %s, %s
`, user.AllowedIPs, user.AcceptRoutes))
	} else {
		configBuilder.WriteString(fmt.Sprintf(`AllowedIPs = %s
`, user.AllowedIPs))
	}

	if user.Endpoint != "" {
		configBuilder.WriteString(fmt.Sprintf("Endpoint = %s\n", user.Endpoint))
	}
	if user.PersistentKeepalive != 0 {
		configBuilder.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", user.PersistentKeepalive))
	}

	return configBuilder.String()
}

// generateIPPool 根据 CIDR 生成 IP 池
func generateIPPool(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// 去掉网络地址和广播地址
	return ips[3 : len(ips)-1], nil
}

// inc 递增 IP 地址
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
