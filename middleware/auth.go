package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"gotask/entity"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string      `json:"user_id"`
	Email  string      `json:"email"`
	Role   entity.Role `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "gotask-super-secret-key-jwt-2026"
	}
	return []byte(secret)
}

func GenerateToken(userID string, email string, role entity.Role) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "header otorisasi diperlukan",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "format otorisasi harus Bearer <token>",
			})
			return
		}

		tokenString := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("metode signing token tidak valid")
			}
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "token tidak valid atau sudah kadaluarsa",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

func AuthorizeRoles(allowedRoles ...entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "role pengguna tidak ditemukan dalam sesi",
			})
			return
		}

		userRole, ok := roleVal.(entity.Role)
		if !ok {
			// fallback check if string
			if strRole, isStr := roleVal.(string); isStr {
				userRole = entity.Role(strRole)
			} else {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "format role pengguna tidak valid",
				})
				return
			}
		}

		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "akses ditolak: role Anda tidak memiliki izin untuk tindakan ini",
		})
	}
}

func GetCurrentUserID(c *gin.Context) (string, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user_id tidak ditemukan di context")
	}
	id, ok := val.(string)
	if !ok {
		return "", errors.New("user_id bukan string")
	}
	return id, nil
}

func GetCurrentUserRole(c *gin.Context) (entity.Role, error) {
	val, exists := c.Get("role")
	if !exists {
		return "", errors.New("role tidak ditemukan di context")
	}
	r, ok := val.(entity.Role)
	if !ok {
		if s, okStr := val.(string); okStr {
			return entity.Role(s), nil
		}
		return "", errors.New("role tidak valid")
	}
	return r, nil
}
