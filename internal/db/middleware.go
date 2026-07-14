package db

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/realestate-trust/monorepo/internal/core"
)

var JWTSecret = []byte("super-secret-key-for-local-demo-only")

// GenerateJWT creates a dummy JWT for a given user ID and role (for demo purposes)
func GenerateJWT(userID string, role core.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": string(role),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// RBACMiddleware checks if the user's role from the JWT token is among the allowed roles.
func RBACMiddleware(allowedRoles ...core.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userToken, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: missing or invalid token"})
			}
			claims, ok := userToken.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: invalid claims"})
			}

			roleRaw, ok := claims["role"].(string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: missing role in token"})
			}
			role := core.UserRole(roleRaw)

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: insufficient permissions"})
		}
	}
}
