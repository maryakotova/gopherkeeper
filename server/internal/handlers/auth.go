package handlers

import (
	"GophKeeper/server/internal/models"
	"GophKeeper/server/internal/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: s,
	}
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = fmt.Errorf("[RegisterHandler]: ошибка при парсинге JSON: %w", err)
		log.Print(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if req.Username == "" || req.Password == "" {
		err := fmt.Errorf("[RegisterHandler]: имя пользователя и пароль должны быть заполнены")
		log.Print(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := h.service.Register(c.Request.Context(), req.Username, req.Password); err != nil {
		err = fmt.Errorf("ошибка при создании пользователя: %w", err)
		log.Print(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)
	log.Printf("[RegisterHandler]: пользователь создан")

}

func (h *AuthHandler) LoginHandler(c *gin.Context) {

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = fmt.Errorf("[LoginHandler]: ошибка при парсинге JSON: %w", err)
		log.Print(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if req.Username == "" || req.Password == "" {
		err := fmt.Errorf("[LoginHandler]: имя пользователя и пароль должны быть заполнены")
		log.Print(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	jwt, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		err = fmt.Errorf("ошибка при логине : %w", err)
		log.Print(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
	log.Printf("[LoginHandler]: успешный логин, токен отправлен")
}
