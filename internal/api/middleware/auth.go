package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func AuthRequired(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// gets the bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" { // exclusively for ws because ws can't set up custom headers. So, I set auth so that it takes token from query
			if t := c.Query("token"); t != "" {
				authHeader = "Bearer " + t
			}
		}
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ") // remove leading spaces

		// Validate JWT signature and expiry
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Stateful check — token must exist in sessions table
		var count int
		db.Get(&count, `SELECT COUNT(*) FROM sessions WHERE token = $1`, tokenStr)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired or logged out"})
			return
		}

		// Inject claims into context for handlers to use
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("email", claims["email"].(string))
		c.Set("token", tokenStr)

		c.Next()
	}
}
