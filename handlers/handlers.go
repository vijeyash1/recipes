package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/vijeyash1/recipes/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(collection *mongo.Collection, ctx context.Context, redis *redis.Client) *RecipesHandler {
	return &RecipesHandler{collection: collection, ctx: ctx, redisClient: redis}
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
func (app *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	recipe := &models.Recipe{}
	if err := c.ShouldBindJSON(recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := app.collection.InsertOne(app.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error while inserting recipe"})
		return
	}
	log.Println("removing data from redis")
	app.redisClient.Del(app.ctx, "recipes")
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation GET /recipes ListRecipes
// Returns a list of recipes
// ---
// produces:
// - application/json
// responses:
//   200:
//     description: Successful operation
func (app *RecipesHandler) ListRecipesHandler(c *gin.Context) {

	val, err := app.redisClient.Get(app.ctx, "recipes").Result()
	if err == redis.Nil {
		log.Println("Getting data from MongoDB")
		curr, err := app.collection.Find(app.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "error while fetching recipes from MongoDB"})
			return
		}
		defer curr.Close(app.ctx)
		recipes := make([]models.Recipe, 0)
		for curr.Next(app.ctx) {
			recipe := &models.Recipe{}
			err := curr.Decode(recipe)
			if err != nil {
				log.Fatal(err)
			}
			recipes = append(recipes, *recipe)
		}
		data, _ := json.Marshal(recipes)
		app.redisClient.Set(app.ctx, "recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching data from redis"})
		return
	} else {
		log.Println("Getting data from Redis")
		var recipes []models.Recipe
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}

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
func (app *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	ID := c.Param("id")
	recipe := &models.Recipe{}
	if err := c.ShouldBindJSON(recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	_, err = app.collection.UpdateOne(app.ctx, bson.M{"_id": objectId}, bson.D{{"$set", bson.D{{"name", recipe.Name}, {"tags", recipe.Tags}, {"ingredients", recipe.Ingredients}, {"instructions", recipe.Instructions}}}})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "error while updating recipe"})
		return
	}
	log.Println("removing data from redis")
	app.redisClient.Del(app.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"message": "recipe updated",
	})
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
func (app *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	ID := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return

	}
	if err, _ := app.collection.DeleteOne(app.ctx, bson.M{"_id": objectId}); err != nil {
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
func (app *RecipesHandler) GetOneHandler(c *gin.Context) {
	ID := c.Param("id")
	objectId, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	recipe := &models.Recipe{}
	if err := app.collection.FindOne(app.ctx, bson.M{"_id": objectId}).Decode(recipe); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}
