package handlers

import (
	"context"
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
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

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("token", sessionToken)
	session.Save()
	c.JSON(http.StatusOK, gin.H{
		"message": "logged in",
	})
}

// swagger:operation POST /refresh auth refresh
// Refresh token
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Successful operation
//     '401':
//         description: Invalid credentials
func (app *AuthHandler) RefreshHandler(c *gin.Context) {
	session := sessions.Default(c)
	sessionToken := session.Get("token")
	sessionUser := session.Get("username")
	if sessionToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
		return
	}

	sessionToken = xid.New().String()
	session.Set("username", sessionUser.(string))
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "New session issued"})
}

// swagger:operation POST /signout auth signOut
// Signing out
// ---
// responses:
//     '200':
//         description: Successful operation
func (app *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
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
		session := sessions.Default(c)
		token := session.Get("token")
		if token == nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Not logged in yet"})
			c.Abort()
			return
		}
		c.Next()
	}
}
