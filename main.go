// Recipes API
//
// This is a sample recipes API. You can find out more about the API at https://github.com/vijeyash1/recipes
//
//	Schemes: http
//  Host: localhost:9000
//	BasePath: /
//	Version: 1.0.0
//	Contact: Vijeyash <avijeyash@gmail.com> http://vijeyash.com +918489635967
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
// swagger:meta
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/vijeyash1/recipes/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var authHandler *handlers.AuthHandler
var recipesHandler *handlers.RecipesHandler

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongoUrl := os.Getenv("MONGO_URI")
	dbName := os.Getenv("Db_Name")
	collName := os.Getenv("Collection_Name")
	ctx := context.Background()
	redisOptions := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisOptions)
	status := redisClient.Ping(ctx)
	fmt.Println(status)

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongoUrl).
		SetServerAPIOptions(serverAPIOptions)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}
	collection := client.Database(dbName).Collection(collName)
	recipesHandler = handlers.NewRecipesHandler(collection, ctx, redisClient)
	collectionUsers := client.Database(dbName).Collection("users")
	authHandler = handlers.NewAuthHandler(collectionUsers, ctx)

	log.Println("Connected to MongoDB")
}

func main() {

	router := gin.Default()
	router.Use(cors.Default())
	store, _ := redisStore.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	router.Use(sessions.Sessions("recipes_api", store))
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/signup", authHandler.SignupHandler)
	router.POST("/signout", authHandler.SignOutHandler)

	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	authorized.POST("/recipes", recipesHandler.NewRecipeHandler)

	authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	authorized.GET("/recipes/search", recipesHandler.GetOneHandler)

	println("server starts at port 9000")
	router.Run(":9000") // listen and serve on
}
