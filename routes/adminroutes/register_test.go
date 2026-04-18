package adminroutes

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegister_NilJWT_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r, nil, nil, nil, nil, nil, nil, nil, nil)
}
