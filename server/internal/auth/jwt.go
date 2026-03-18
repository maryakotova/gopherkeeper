package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte("GopherKeeper_Secret_Key")

func GenerateJWT(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(secret)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неподдерживаемый метод подписи: %v", token.Header["alg"])
			}
			return secret, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", int(claims["user_id"].(float64)))
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
		}
	}
}
