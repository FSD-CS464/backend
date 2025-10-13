package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginStub(c *gin.Context) {
	// TODO: real OAuth/JWT minting
	c.JSON(http.StatusOK, gin.H{"access_token":"dev-token", "refresh_token":"dev-refresh"})
}
func RefreshStub(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"access_token":"dev-token"})
}
