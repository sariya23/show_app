package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ssov1 "github.com/sariya23/sso_proto/gen/sso"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Handler struct {
	GrpcClient ssov1.AuthClient
	Log        *slog.Logger
}

func (h *Handler) Login(c *gin.Context) {
	const op = "handler.Login"
	ctx, cancel := context.WithTimeout(c, time.Second*15)
	log := h.Log.With("op", slog.String("op", op))
	log.Info("login user")
	defer cancel()
	b := c.Request.Body
	data, err := io.ReadAll(b)
	if err != nil {
		log.Error("cannot read body", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error while read body"})
		return
	}
	defer b.Close()
	var r RegisterRequest
	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Error("cannor unmarshal body", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error while unmarshal body"})
		return
	}

	client := h.GrpcClient
	response, err := client.Login(ctx, &ssov1.LoginRequest{Email: r.Login, Password: r.Password, AppId: 1})
	if err != nil {
		log.Error("smth wrong in sso service", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "insernal error in sso service", "err": err.Error()})
		return
	}
	c.Header("Authorization", "Bearer "+response.GetToken())
	c.JSON(http.StatusOK, gin.H{"message": "login success"})
}

func (h *Handler) Register(c *gin.Context) {
	const op = "handler.Register"
	log := h.Log.With("op", slog.String("op", op))
	log.Info("register user")
	ctx, cancel := context.WithTimeout(c, time.Second*15)
	defer cancel()
	b := c.Request.Body
	data, err := io.ReadAll(b)
	if err != nil {
		log.Error("cannot read body", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error while read body"})
		return
	}
	defer b.Close()
	var r RegisterRequest
	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Error("cannor unmarshal body", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error while unmarshal body"})
		return
	}

	client := h.GrpcClient
	response, err := client.Register(ctx, &ssov1.RegisterRequest{Email: r.Login, Password: r.Password})
	if err != nil {
		log.Error("smth wrong in sso service", slog.String("err", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error in sso service", "err": err.Error()})
		return
	}
	log.Info("register success")
	c.JSON(http.StatusOK, gin.H{"message": "register success", "user_id": response.GetUserId()})
}
