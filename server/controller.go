package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type AddUserRequest struct {
	ID              string `json:"id"`
	AllowedIPs      string `json:"allowedips"`
	PreUp           string `json:"pre_up"`
	PostUp          string `json:"post_up"`
	PreDown         string `json:"pre_down"`
	PostDown        string `json:"post_down"`
	AdvertiseRoutes string `json:"advertise_routes"`
	AcceptRoutes    string `json:"accept_routes"`
}

type DeleteUserRequest struct {
	ID string `json:"id"`
}

type GetUserRequest struct {
	ID string `json:"id"`
}

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func setupHandler(c *gin.Context) {
	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	defer userManager.db.Close()

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	config, err := userManager.GenerateServerConfig(*serverConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	configPath := "./wg.conf"
	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "VPN server configuration setup successfully", Data: gin.H{"config": config}})
}

func addUserHandler(c *gin.Context) {
	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	defer userManager.db.Close()

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	endpoint := fmt.Sprintf("%s:%d", serverConfig.ServerIP, serverConfig.Port)
	persistentKeepalive := 25

	err = userManager.AddUser(&UserConfig{
		UserID:              req.ID,
		AllowedIPs:          req.AllowedIPs,
		Endpoint:            endpoint,
		AdvertiseRoutes:     req.AdvertiseRoutes,
		AcceptRoutes:        req.AcceptRoutes,
		PersistentKeepalive: persistentKeepalive,
		PreUp:               req.PreUp,
		PostUp:              req.PostUp,
		PreDown:             req.PreDown,
		PostDown:            req.PostDown,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	for _, user := range users {
		if user.UserID == req.ID {
			c.JSON(http.StatusOK, Response{Message: "User added successfully", Data: gin.H{"user_config": generateUserConfig(*serverConfig, user)}})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Message: "User not found", Data: gin.H{"error": "User not found"}})
}

func deleteUserHandler(c *gin.Context) {
	var req DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	defer userManager.db.Close()

	err = userManager.DeleteUser(req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "User deleted successfully", Data: gin.H{"user_id": req.ID}})
}

func getUserHandler(c *gin.Context) {
	var req GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "Invalid request body"}})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, Response{Message: "Bad Request", Data: gin.H{"error": "User ID is required"}})
		return
	}

	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	defer userManager.db.Close()

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	serverConfig, err := LoadServerConfig("server.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	for _, user := range users {
		if user.UserID == req.ID {
			c.JSON(http.StatusOK, Response{Message: "User found", Data: gin.H{"user_config": generateUserConfig(*serverConfig, user)}})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Message: "User not found", Data: gin.H{"error": "User not found"}})
}

func getAllUsersHandler(c *gin.Context) {
	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	defer userManager.db.Close()

	users, err := userManager.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "Users retrieved successfully", Data: users})
}

func getAllRoutesHandler(c *gin.Context) {
	userManager, err := NewUserManager("./users.db")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
	}
	defer userManager.db.Close()
	routes, err := userManager.GetAllRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
	}
	c.JSON(http.StatusOK, Response{Message: "Routes retrieved successfully", Data: routes})
}
