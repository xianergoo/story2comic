package v1

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SetupRouter 设置API路由基础框架
func SetupRouter(sessionSecret string, registerRoutes func(api *gin.RouterGroup)) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(APIVersionMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(RateLimitMiddleware(rate.Limit(10), 20))

	// Session 中间件
	store := cookie.NewStore([]byte(sessionSecret))
	r.Use(sessions.Sessions("novelforge_api", store))

	api := r.Group("/api/v1")
	registerRoutes(api)

	return r
}
