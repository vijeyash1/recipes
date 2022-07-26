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
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var recipes []Recipe

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Title        string             `json:"title" bson:"title"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

type application struct {
	collection *mongo.Collection
}

var app = &application{}

func init() {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://vijeyash:vijeyash1V@cluster0.refiu.mongodb.net/?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}
	collection := client.Database("recipes").Collection("recipes")
	app = &application{collection: collection}
	log.Println("Connected to MongoDB")


func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/recipes", app.NewRecipeHandler)
	router.GET("/recipes", app.ListRecipesHandler)
	router.PUT("/recipes/:id", app.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", app.DeleteRecipeHandler)
	router.GET("/recipes/search", app.GetOneHandler)
	println("server starts at port 9000")
	router.Run(":9000") // listen and serve on
}

// swagger:operation POST /recipes NewRecipe
// Creates a new recipe
// ---
// produces:
// - application/json
// consumes:
// - application/json
// responses:
//   200:
//     description: returns the created recipe
//   400:
//     description: invalid request
func (app *application) NewRecipeHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	recipe := &Recipe{}
	if err := c.ShouldBindJSON(recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := app.collection.InsertOne(ctx, recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while inserting recipe"})
		return
	}
	recipes = append(recipes, *recipe)
	c.JSON(http.StatusOK, recipes)

}

// swagger:operation GET /recipes ListRecipes
// Returns a list of recipes
// ---
// produces:
// - application/json
// responses:
//   200:
//     description: Successful operation
func (app *application) ListRecipesHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	curr, err := app.collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer curr.Close(ctx)

	for curr.Next(ctx) {
		recipe := &Recipe{}
		err := curr.Decode(recipe)
		if err != nil {
			log.Fatal(err)
		}
		recipes = append(recipes, *recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation PUT /recipes/{id} UpdateRecipe
// Updates a recipe
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id of the recipe
//   required: true
//   type: string
// responses:
//   200:
//     description: returns the updated recipe
//   404:
//     description: Recipe not found
func (app *application) UpdateRecipeHandler(c *gin.Context) {
	ID := c.Param("id")
	recipe := &Recipe{}
	if err := c.ShouldBindJSON(recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err, _ := app.collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{"$set": recipe}); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation DELETE /recipes/{id} DeleteRecipe
// Deletes a recipe
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id of the recipe
//   required: true
//   type: string
// responses:
//   200:
//     description: returns the deleted recipe
//   404:
//     description: Recipe not found
func (app *application) DeleteRecipeHandler(c *gin.Context) {
	ID := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return

	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err, _ := app.collection.DeleteOne(ctx, bson.M{"_id": objectId}); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "recipe deleted"})
}

// swagger:operation GET /recipes/{id} SearchRecipes
// Searches recipes
// ---
// produces:
// - application/json
// parameters:
// - name: query
//   in: query
//   description: query to search
//   required: true
//   type: string
// responses:
//   200:
//     description: returns the searched recipes
//   404:
//     description: Recipe not found
func (app *application) GetOneHandler(c *gin.Context) {
	ID := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	recipe := &Recipe{}
	if err := app.collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(recipe); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}
