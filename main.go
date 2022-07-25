package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/", response)

	router.Run(":8080") // listen and serve on
}

func response(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "response from server",
	})
}
