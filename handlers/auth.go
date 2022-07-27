package handlers

import (
	"context"
	"crypto/sha256"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/vijeyash1/recipes/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(collection *mongo.Collection, ctx context.Context) *AuthHandler {
	return &AuthHandler{collection: collection, ctx: ctx}
}

// swagger:operation POST /signin Signin
// used for signin
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: username
// - name: password
//   in: body
//   required: true
// responses:
//   '200':
//     description: sends token
func (app *AuthHandler) SignInHandler(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h := sha256.New()
	curr := app.collection.FindOne(app.ctx, bson.M{"username": user.Username,
		"password": string(h.Sum([]byte(user.Password)))})
	if curr.Err() != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username or password"})
		return
	}

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
}

// swagger:operation POST /signup Signup
// used for signup
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: username
// - name: password
//   in: body
//   required: true
// responses:
//   '200':
//     description: user created
func (app *AuthHandler) SignupHandler(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h := sha256.New()
	_, err := app.collection.InsertOne(app.ctx, bson.M{
		"username": user.Username,
		"password": string(h.Sum([]byte(user.Password))),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while inserting user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user created"})
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

// swagger:operation POST /refresh refreshToken
// used for refresh token
// ---
// produces:
// - application/json
// parameters:
// - name: Authorization
//   in: header
//   required: true
// responses:
//   '200':
//     description: sends refreshed token with extra 10 minutes
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
