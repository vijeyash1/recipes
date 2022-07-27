package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vijeyash1/recipes/models"
)

type AuthHandler struct {
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (app *AuthHandler) SignInHandler(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user.Username == "vijeyash" && user.Password == "password" {
		claims := Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, JWTOutput{Token: tokenString, Expires: time.Now().Add(time.Minute * 10)})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	}
}
func (app *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := c.GetHeader("Authorization")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if token == nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Next()
	}
}

func (app *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	if token == nil || !token.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token not expired"})
		return
	}

	claims.ExpiresAt = time.Now().Add(time.Minute * 10).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, JWTOutput{Token: tokenString, Expires: time.Now().Add(time.Minute * 10)})
}
