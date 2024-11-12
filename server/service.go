package main

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"log"
	"net"
	"os/exec"
	"strings"
)

type UserManager struct {
	db *sql.DB
}

func NewUserManager(dbPath string) (*UserManager, error) {
	db, err := sql.Open("sqlite", dbPath)
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
	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL UNIQUE,
        public_key TEXT NOT NULL,
        private_key TEXT NOT NULL,
        ip TEXT NOT NULL UNIQUE,
        allowed_ips TEXT NOT NULL,
        endpoint TEXT NOT NULL,
        persistent_keepalive INTEGER,
        pre_up TEXT,
        post_up TEXT,
        pre_down TEXT,
        post_down TEXT,
        advertise_routes TEXT,
        accept_routes TEXT
    );`
	_, err := um.db.Exec(createTable)
	return err
}

func (um *UserManager) AddUser(user *UserConfig) error {
	// 从 YAML 文件加载服务器配置
	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		log.Fatal(err)
	}
	ipPoolCIDR := serverConfig.IPPool
	// 生成 IP 池
	ipPool, err := generateIPPool(ipPoolCIDR)
	if err != nil {
		return err
	}
	// 查询表中已存在的 IP 地址
	rows, err := um.db.Query("SELECT ip FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	var usedIPs []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return err
		}
		usedIPs = append(usedIPs, ip)
	}

	// 查找未使用的 IP
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

	// 生成密钥对
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
	stmt, err := um.db.Prepare("INSERT INTO users(user_id, public_key, private_key, ip, allowed_ips, endpoint, persistent_keepalive, pre_up, post_up, pre_down, post_down, advertise_routes, accept_routes) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.UserID, user.PublicKey, user.PrivateKey, newIP, user.AllowedIPs, user.Endpoint, user.PersistentKeepalive, user.PreUp, user.PostUp, user.PreDown, user.PostDown, user.AdvertiseRoutes, user.AcceptRoutes)
	return err
}

func (um *UserManager) GetAllUsers() ([]UserConfig, error) {
	rows, err := um.db.Query("SELECT user_id, public_key, private_key, ip, allowed_ips, endpoint, persistent_keepalive, pre_up, post_up, pre_down, post_down, advertise_routes, accept_routes FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserConfig
	for rows.Next() {
		var user UserConfig
		err = rows.Scan(&user.UserID, &user.PublicKey, &user.PrivateKey, &user.IP, &user.AllowedIPs, &user.Endpoint, &user.PersistentKeepalive, &user.PreUp, &user.PostUp, &user.PreDown, &user.PostDown, &user.AdvertiseRoutes, &user.AcceptRoutes)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (um *UserManager) GetAllRoutes() ([]string, error) {
	rows, err := um.db.Query("SELECT advertise_routes FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]string, 0)
	for rows.Next() {
		var route string
		err = rows.Scan(&route)
		if err != nil {
			return nil, err
		}
		if route != "" {
			results = append(results, route)
		}
	}
	return results, nil
}

func (um *UserManager) UpdateUser(user UserConfig) error {
	stmt, err := um.db.Prepare("UPDATE users SET public_key = ?, private_key = ?, ip = ?, allowed_ips = ?, endpoint = ?, persistent_keepalive = ?, advertise_routes = ?, accept_routes = ? WHERE user_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.PublicKey, user.PrivateKey, user.IP, user.AllowedIPs, user.Endpoint, user.PersistentKeepalive, user.UserID, user.AdvertiseRoutes, user.AcceptRoutes)
	return err
}

func (um *UserManager) UpdateUserEndpoints(serverConfig ServerConfig) error {
	endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)
	stmt, err := um.db.Prepare("UPDATE users SET endpoint = ? WHERE user_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	users, err := um.GetAllUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		_, err := stmt.Exec(endpoint, user.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (um *UserManager) DeleteUser(id string) error {
	stmt, err := um.db.Prepare("DELETE FROM users WHERE user_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
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
func generateUserConfig(serverConfig ServerConfig, user UserConfig) string {
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
