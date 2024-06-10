package main

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/gin-gonic/gin"
)

type AddUserRequest struct {
	ID         string `json:"id"`
	AllowedIPs string `json:"allowedips"`
}

type DeleteUserRequest struct {
	ID string `json:"id"`
}

type GetUserRequest struct {
	ID string `json:"id"`
}

func setupHandler(c *gin.Context) {
	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}
	defer userManager.db.Close()

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	config, err := userManager.GenerateServerConfig(*serverConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	configPath := "./wg.conf"
	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "VPN server configuration setup successfully", "data": gin.H{"config": config}})
}

func addUserHandler(c *gin.Context) {
	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}
	defer userManager.db.Close()

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)
	persistentKeepalive := 25

	err = userManager.AddUser(&UserConfig{
		UserID:              req.ID,
		AllowedIPs:          req.AllowedIPs,
		Endpoint:            endpoint,
		PersistentKeepalive: persistentKeepalive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	for _, user := range users {
		if user.UserID == req.ID {
			c.JSON(http.StatusOK, gin.H{"message": "User added successfully", "data": gin.H{"user_config": generateConfig(*serverConfig, user)}})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "User not found", "data": gin.H{"error": "User not found"}})
}

func deleteUserHandler(c *gin.Context) {
	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}
	defer userManager.db.Close()

	err = userManager.DeleteUser(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "data": gin.H{"user_id": req.ID}})
}

func getUserHandler(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "data": gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}
	defer userManager.db.Close()

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	for _, user := range users {
		if user.UserID == req.ID {
			c.JSON(http.StatusOK, gin.H{"message": "User found", "data": gin.H{"user_config": generateConfig(*serverConfig, user)}})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "User not found", "data": gin.H{"error": "User not found"}})
}

func getAllUsersHandler(c *gin.Context) {
	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}
	defer userManager.db.Close()

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error", "data": gin.H{"error": err.Error()}})
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 15, 20, 0, ' ', tabwriter.TabIndent)
	fmt.Fprintf(w, "ID\tIP\n")

	var userList []map[string]string
	for _, user := range users {
		userList = append(userList, map[string]string{
			"id": user.UserID,
			"ip": user.IP,
		})
		fmt.Fprintf(w, "%s\t%s\t\n", user.UserID, user.IP)
	}

	w.Flush()

	c.JSON(http.StatusOK, gin.H{"message": "Users retrieved successfully", "data": gin.H{"users": userList}})
}
