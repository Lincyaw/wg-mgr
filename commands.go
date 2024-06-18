package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

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
			advertiseRoutes, _ := cmd.Flags().GetString("advertise-routes")
			acceptRoutes, _ := cmd.Flags().GetBool("accept-routes")
			preup, _ := cmd.Flags().GetString("preup")
			postup, _ := cmd.Flags().GetString("postup")
			predown, _ := cmd.Flags().GetString("predown")
			postdown, _ := cmd.Flags().GetString("postdown")
			endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)
			persistentKeepalive := 25

			var acceptedRoutes string
			if acceptRoutes {
				var routes []string

				availableRoutes, err := userManager.GetAllRoutes()
				if err != nil {
					log.Fatal(err)
				}

				// 使用 survey 进行交互式选择
				prompt := &survey.MultiSelect{
					Message: "Select routes to accept:",
					Options: availableRoutes,
				}
				err = survey.AskOne(prompt, &routes)
				if err != nil {
					log.Fatal(err)
				}

				acceptedRoutes = fmt.Sprintf("%s", strings.Join(routes, ","))
			}

			err = userManager.AddUser(&UserConfig{
				UserID:              userID,
				AllowedIPs:          allowedIPs,
				AdvertiseRoutes:     advertiseRoutes,
				Endpoint:            endpoint,
				AcceptRoutes:        acceptedRoutes,
				PersistentKeepalive: persistentKeepalive,
				PreUp:               preup,
				PostUp:              postup,
				PreDown:             predown,
				PostDown:            postdown,
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
					fmt.Printf("%s", generateUserConfig(*serverConfig, user))
				}
			}
		},
	}
	addUserCmd.Flags().String("id", "", "User ID")
	addUserCmd.Flags().String("allowedips", "", "For client side, which traffic can be passed to the server")
	addUserCmd.Flags().String("advertise-routes", "", "Advertise a route to the server, so that other client can connect to it")
	addUserCmd.Flags().Bool("accept-routes", false, "Accept a route to the server, so that other client can connect to it")
	// PostUp = sysctl -w net.ipv4.ip_forward=1; iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE
	// PostDown = sysctl -w net.ipv4.ip_forward=0; iptables -t nat -D POSTROUTING -o wg0 -j MASQUERADE
	addUserCmd.Flags().String("preup", "", "Pre up")
	addUserCmd.Flags().String("postup", "", "Post up")
	addUserCmd.Flags().String("predown", "", "Pre down")
	addUserCmd.Flags().String("postdown", "", "Post down")

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
			if userID == "" {
				log.Fatal("You must provide a user ID")
			}
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
					fmt.Printf("%s", generateUserConfig(*serverConfig, user))
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

func Server() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run server",
		Run: func(cmd *cobra.Command, args []string) {
			r := gin.Default()

			r.POST("/setup", setupHandler)
			r.POST("/adduser", addUserHandler)
			r.POST("/deluser", deleteUserHandler)
			r.POST("/getuser", getUserHandler)
			r.POST("/getall", getAllUsersHandler)

			addr, _ := cmd.Flags().GetString("addr")
			if addr == "" {
				addr = ":8080"
			}

			if err := r.Run(addr); err != nil {
				log.Fatal(err)
			}
		},
	}
	serverCmd.Flags().String("addr", "", "ip:port")
	return serverCmd
}
