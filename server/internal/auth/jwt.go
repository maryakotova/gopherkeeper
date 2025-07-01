package auth

import (
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

		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", int(claims["user_id"].(float64)))
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
		}
	}
}

// type AuthMiddleware struct {
// 	secret []byte
// }

// func NewAuthMiddleware(secret string) *AuthMiddleware {
// 	return &AuthMiddleware{
// 		secret: []byte(secret),
// 	}
// }

// // // var secret = "GopherKeeper_Secret" //[]byte(os.Getenv("JWT_SECRET"))

// func (auth *AuthMiddleware) GenerateJWT(userID uint) (string, error) {
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": userID,
// 		"exp":     time.Now().Add(24 * time.Hour).Unix(),
// 	})

// 	return token.SignedString(auth.secret)
// }

// func (auth *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		auth := c.GetHeader("Authorization")
// 		tokenStr := strings.TrimPrefix(auth, "Bearer ")

// 		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
// 			return auth.secret, nil
// 		})

// 		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 			c.Set("user_id", uint(claims["user_id"].(float64)))
// 			c.Next()
// 		} else {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			c.Abort()
// 		}
// 	}
// }
