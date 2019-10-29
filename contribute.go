package main

import "github.com/gin-gonic/gin"

func contributeHandler(c *gin.Context) {
	display(c, "contribute", &Page{Title: "Contibute"})
}
