package clienthandler

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/pkg/snowflake"
	"gin-scaffold/pkg/notify"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	_ = snowflake.Init(1)
	notify.SetDefault(notify.Noop{})
	code := m.Run()
	notify.SetDefault(nil)
	os.Exit(code)
}
