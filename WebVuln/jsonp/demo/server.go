package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/jsonp", func(c *gin.Context) {
		data := map[string]interface{}{
			"name": "bar",
			"desc": "desc..",
		}
		c.JSONP(http.StatusOK, data)
	})
	r.Run(":8080")
}
