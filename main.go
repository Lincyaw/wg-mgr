package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
	_ "modernc.org/sqlite"

	"github.com/spf13/cobra"
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
        persistent_keepalive INTEGER
    );`
	_, err := um.db.Exec(createTable)
	return err
}

func (um *UserManager) AddUser(user *UserConfig) error {
	// 查询表中是否存在数据
	var count int
	err := um.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	var maxIP string

	if count == 0 {
		maxIP = "100.10.10.1"
	} else {
		// 查询表中最大的 IP
		err := um.db.QueryRow("SELECT MAX(ip) FROM users").Scan(&maxIP)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	// 解析最大的 IP，并加一
	ipParts := strings.Split(maxIP, ".")
	lastOctet, err := strconv.Atoi(ipParts[len(ipParts)-1])
	if err != nil {
		return err
	}
	lastOctet++

	// 构建新的 IP
	newIP := fmt.Sprintf("%s.%d", strings.Join(ipParts[:len(ipParts)-1], "."), lastOctet)

	// 生成密钥对
	privateKey, publicKey, err := generateKeys()
	if err != nil {
		return err
	}
	user.PrivateKey = privateKey
	user.PublicKey = publicKey
	if user.AllowedIPs == "" {
		user.AllowedIPs = newIP + "/32"
	}

	stmt, err := um.db.Prepare("INSERT INTO users(user_id, public_key, private_key, ip, allowed_ips, endpoint, persistent_keepalive) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.UserID, user.PublicKey, user.PrivateKey, newIP, user.AllowedIPs, user.Endpoint, user.PersistentKeepalive)
	return err
}

func (um *UserManager) GetAllUsers() ([]UserConfig, error) {
	rows, err := um.db.Query("SELECT user_id, public_key, private_key, ip, allowed_ips, endpoint, persistent_keepalive FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserConfig
	for rows.Next() {
		var user UserConfig
		err = rows.Scan(&user.UserID, &user.PublicKey, &user.PrivateKey, &user.IP, &user.AllowedIPs, &user.Endpoint, &user.PersistentKeepalive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (um *UserManager) UpdateUser(user UserConfig) error {
	stmt, err := um.db.Prepare("UPDATE users SET public_key = ?, private_key = ?, ip = ?, allowed_ips = ?, endpoint = ?, persistent_keepalive = ? WHERE user_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.PublicKey, user.PrivateKey, user.IP, user.AllowedIPs, user.Endpoint, user.PersistentKeepalive, user.UserID)
	return err
}

func (um *UserManager) DeleteUser(id string) error {
	stmt, err := um.db.Prepare("DELETE FROM users WHERE user_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id)
	return err
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

func generateConfig(serverConfig ServerConfig, user UserConfig) string {
	var configBuilder strings.Builder

	configBuilder.WriteString(fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
`, user.PrivateKey, user.IP))

	configBuilder.WriteString(fmt.Sprintf(`
[Peer]
PublicKey = %s
AllowedIPs = %s
`, user.PublicKey, user.AllowedIPs))

	if user.Endpoint != "" {
		configBuilder.WriteString(fmt.Sprintf("Endpoint = %s\n", user.Endpoint))
	}
	if user.PersistentKeepalive != 0 {
		configBuilder.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", user.PersistentKeepalive))
	}

	return configBuilder.String()
}

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
		configBuilder.WriteString(fmt.Sprintf(`[Peer]
PublicKey = %s
AllowedIPs = %s
`, user.PublicKey, user.AllowedIPs))
	}

	return configBuilder.String(), nil
}

func LoadServerConfig(filePath string) (*ServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func Setup() *cobra.Command {
	var setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup VPN server configuration",
		Run: func(cmd *cobra.Command, args []string) {
			userManager, err := NewUserManager("./users.db")
			if err != nil {
				log.Fatal(err)
			}
			defer userManager.db.Close()

			// 从 YAML 文件加载服务器配置
			serverConfig, err := LoadServerConfig("server.yaml")
			if err != nil {
				log.Fatal(err)
			}

			config, err := userManager.GenerateServerConfig(*serverConfig)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(config)

			// 将配置写入文件
			configPath := "./wg.conf"
			err = os.WriteFile(configPath, []byte(config), 0644)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return setupCmd
}
func Add() *cobra.Command {
	var addUserCmd = &cobra.Command{
		Use:   "adduser",
		Short: "Add a new user to VPN",
		Run: func(cmd *cobra.Command, args []string) {
			userManager, err := NewUserManager("./users.db")
			if err != nil {
				log.Fatal(err)
			}
			defer userManager.db.Close()

			serverConfig, err := LoadServerConfig("server.yaml")
			if err != nil {
				log.Fatal(err)
			}

			userID, _ := cmd.Flags().GetString("id")
			if userID == "" {
				log.Fatal("You must provide a user ID")
			}
			allowedIPs, _ := cmd.Flags().GetString("allowedips")
			endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)
			persistentKeepalive := 25

			err = userManager.AddUser(&UserConfig{
				UserID:              userID,
				AllowedIPs:          allowedIPs,
				Endpoint:            endpoint,
				PersistentKeepalive: persistentKeepalive,
			})
			if err != nil {
				log.Fatal(err)
			}

			users, err := userManager.GetAllUsers()
			if err != nil {
				log.Fatal(err)
			}
			for _, user := range users {
				if user.UserID == userID {
					fmt.Printf("%s", generateConfig(*serverConfig, user))
				}
			}
		},
	}
	addUserCmd.Flags().String("id", "", "User ID")
	addUserCmd.Flags().String("allowedips", "", "Allowed IPs")

	return addUserCmd
}
func Delete() *cobra.Command {
	var deleteUserCmd = &cobra.Command{
		Use:   "deluser",
		Short: "Delete a user from VPN",
		Run: func(cmd *cobra.Command, args []string) {
			userManager, err := NewUserManager("./users.db")
			if err != nil {
				log.Fatal(err)
			}
			defer userManager.db.Close()

			userID, _ := cmd.Flags().GetString("id")

			err = userManager.DeleteUser(userID)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("User %s deleted successfully\n", userID)
		},
	}
	deleteUserCmd.Flags().String("id", "", "User ID")
	return deleteUserCmd
}

func Get() *cobra.Command {
	var getUserCmd = &cobra.Command{
		Use:   "getuser",
		Short: "Get a user from VPN",
		Run: func(cmd *cobra.Command, args []string) {
			userManager, err := NewUserManager("./users.db")
			if err != nil {
				log.Fatal(err)
			}
			defer userManager.db.Close()
			userID, _ := cmd.Flags().GetString("id")
			if userID == "" {
				log.Fatal("User ID is required")
			}
			users, err := userManager.GetAllUsers()
			if err != nil {
				return
			}
			serverConfig, err := LoadServerConfig("server.yaml")
			if err != nil {
				log.Fatal(err)
			}
			for _, user := range users {
				if user.UserID == userID {
					fmt.Printf("%s", generateConfig(*serverConfig, user))
				}
			}
		},
	}
	getUserCmd.Flags().String("id", "", "User ID")
	return getUserCmd
}
func GetAllUsers() *cobra.Command {
	var getAllUsersCmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all users from VPN",
		Run: func(cmd *cobra.Command, args []string) {
			userManager, err := NewUserManager("./users.db")
			if err != nil {
				log.Fatal(err)
			}
			defer userManager.db.Close()
			users, err := userManager.GetAllUsers()
			if err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 15, 20, 0, ' ', tabwriter.TabIndent)
			fmt.Fprintf(w, "ID\tIP\n")

			for _, user := range users {
				fmt.Fprintf(w, "%s\t%s\t\n", user.UserID, user.IP)
			}

			w.Flush()
		},
	}
	return getAllUsersCmd
}
func main() {
	var rootCmd = &cobra.Command{Use: "vpn-tool"}

	rootCmd.AddCommand(Setup(), Add(), Delete(), Get(), GetAllUsers())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
