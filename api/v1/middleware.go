package v1

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"golang.org/x/time/rate"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境应该限制域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limit rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(limit, burst)
	
	return func(c *gin.Context) {
		if !limiter.Allow() {
			Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return requestid.New()
}

// APIVersionMiddleware API版本中间件
func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("api_version", "v1")
		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		c.Next()
		
		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		requestID := c.GetString("X-Request-ID")
		
		// 记录API访问日志
		if query != "" {
			path = path + "?" + query
		}
		
		// 可以根据需要记录到文件或日志系统
		if status >= 400 {
			// 错误日志
			c.Error(fmt.Errorf("[%s] %s %s %d %v %s", 
				clientIP, method, path, status, latency, requestID))
		}
		
		// 简单的控制台输出
		fmt.Printf("[API] %s %s %d %v %s\n", 
			method, path, status, latency, requestID)
	}
}