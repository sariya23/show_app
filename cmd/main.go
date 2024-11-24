package main

import (
	"log/slog"
	"net/http"
	"os"
	"show/internal/handler"
	"show/internal/middleware"

	"github.com/gin-gonic/gin"
	ssov1 "github.com/sariya23/sso_proto/gen/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	conn, err := grpc.NewClient("localhost:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := ssov1.NewAuthClient(conn)
	router := gin.Default()
	h := handler.Handler{GrpcClient: client, Log: log}
	router.POST("/register", h.Register)
	router.GET("/login", h.Login)
	protected := router.Group("/profile")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/:uid", func(c *gin.Context) {
		userID, _ := c.Get("uid")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{
			"message": "This is your profile",
			"userID":  userID,
			"email":   email,
		})
	})
	if err := router.Run("0.0.0.0:8082"); err != nil {
		panic(err)
	}
}
