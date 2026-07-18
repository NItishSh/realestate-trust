package db

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

const CorrelationIDHeader = "X-Correlation-ID"
const CorrelationIDContextKey = "correlation_id"

// CorrelationIDMiddleware extracts or generates a Request/Correlation ID
func CorrelationIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			cid := req.Header.Get(CorrelationIDHeader)
			if cid == "" {
				cid = uuid.New().String()
			}

			// Store in Echo context
			c.Set(CorrelationIDContextKey, cid)

			// Propagate into Request Context
			ctx := context.WithValue(req.Context(), CorrelationIDContextKey, cid)
			c.SetRequest(req.WithContext(ctx))

			// Add to response header
			c.Response().Header().Set(CorrelationIDHeader, cid)

			return next(c)
		}
	}
}

// RequestLoggerMiddleware logs HTTP requests with correlation context
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			status := 0
			if res, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
				status = res.Status
			}
			if status == 0 && err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = http.StatusInternalServerError
				}
			}

			slog.InfoContext(req.Context(), "HTTP request",
				"method", req.Method,
				"uri", req.URL.Path,
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			return err
		}
	}
}

// SlogCorrelationHandler integrates context-based Correlation ID to structured slog logs
type SlogCorrelationHandler struct {
	slog.Handler
}

func (h *SlogCorrelationHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if cid, ok := ctx.Value(CorrelationIDContextKey).(string); ok {
			r.AddAttrs(slog.String("correlation_id", cid))
		}
	}
	return h.Handler.Handle(ctx, r)
}
