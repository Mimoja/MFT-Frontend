package main

import "github.com/gin-gonic/gin"

func libraryHandler(c *gin.Context) {
	display(c, "library", &Page{Title: "Library"})
}
