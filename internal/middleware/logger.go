package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
)

// Logger logs each request with method, path, status and latency.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		gin.DefaultWriter.Write([]byte(
			"[http] " + method + " " + path + " | " +
				itoa(status) + " | " + latency.String() + "\n",
		))
	}
}

// Recovery catches panics and returns a clean 500 envelope.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response.Internal(c, nil)
	})
}

func itoa(i int) string {
	// avoid strconv import noise for a tiny helper
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
