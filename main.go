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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/vijeyash1/recipes/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var recipesHandler *handlers.RecipesHandler

func init() {
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
		ApplyURI("mongodb+srv://vijeyash:vijeyash1V@cluster0.refiu.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}
	collection := client.Database("recipes").Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(collection, ctx, redisClient)
	log.Println("Connected to MongoDB")
}
func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.GetOneHandler)
	println("server starts at port 9000")
	router.Run(":9000") // listen and serve on
}
