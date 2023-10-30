package middleware

import "github.com/gin-gonic/gin"

// WithTestingDharitriFacade middleware will set up an DharitriFacade object in the gin context
// should only be used in testing and not in conjunction with other middlewares as c.Next() instruction
// is not safe to be used multiple times in the same context.
func WithTestingDharitriFacade(dharitriFacade DharitriHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("dharitriFacade", dharitriFacade)
		c.Next()
	}
}
