package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/:name", response)

	router.Run(":8080") // listen and serve on
}

func response(c *gin.Context) {
	name := c.Param("name")
	c.JSON(200, gin.H{
		"message": "Hello " + name,
	})
}
