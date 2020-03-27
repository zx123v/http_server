package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/hello", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("Hello, It Home!"))
	})
	
	router.Run(":8888")
}