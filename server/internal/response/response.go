package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type body struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, body{OK: true, Data: data})
}

func Fail(c *gin.Context, status int, msg string) {
	if status == 0 {
		status = http.StatusBadRequest
	}
	c.JSON(status, body{OK: false, Error: msg})
}

func Abort(c *gin.Context, status int, msg string) {
	Fail(c, status, msg)
	c.Abort()
}
