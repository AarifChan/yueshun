package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"zhizhang-server/internal/pkg/errors"
	"zhizhang-server/internal/pkg/response"
)

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string
	AccessTTL  int // seconds
	RefreshTTL int // seconds
}

var jwtCfg *JWTConfig

// InitJWT 初始化 JWT 配置
func InitJWT(cfg *JWTConfig) {
	jwtCfg = cfg
}

// Claims JWT 声明
type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	RoleID    uint   `json:"role_id"`
	DeptID    uint   `json:"dept_id"`
	CompanyID uint   `json:"company_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成访问令牌
func GenerateToken(userID uint, username string, roleID, deptID, companyID uint) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Username:  username,
		RoleID:    roleID,
		DeptID:    deptID,
		CompanyID: companyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtCfg.AccessTTL) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "zhizhang-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtCfg.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtCfg.RefreshTTL) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtCfg.Secret + "_refresh"))
}

// GetAccessTTL 返回访问令牌有效期（秒）
func GetAccessTTL() int {
	return jwtCfg.AccessTTL
}

// GetRefreshTTL 返回刷新令牌有效期（秒）
func GetRefreshTTL() int {
	return jwtCfg.RefreshTTL
}

// ParseRefreshToken 解析刷新令牌，返回用户 ID
func ParseRefreshToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtCfg.Secret + "_refresh"), nil
	})
	if err != nil {
		if err == jwt.ErrTokenExpired {
			return 0, errors.ErrExpiredToken
		}
		return 0, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, errors.ErrInvalidToken
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.ErrInvalidToken
	}
	return uint(id), nil
}

// ParseToken 解析令牌
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtCfg.Secret), nil
	})
	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, errors.ErrExpiredToken
		}
		return nil, errors.ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.ErrInvalidToken
}

// JWTMiddleware Gin JWT 中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少 Authorization 头")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "Authorization 格式错误")
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			if err == errors.ErrExpiredToken {
				response.Unauthorized(c, "令牌已过期")
			} else {
				response.Unauthorized(c, "无效的令牌")
			}
			c.Abort()
			return
		}

		// 将用户信息注入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleID", claims.RoleID)
		c.Set("deptID", claims.DeptID)
		c.Set("companyID", claims.CompanyID)

		c.Next()
	}
}

// GetCompanyID 从上下文获取公司 ID
func GetCompanyID(c *gin.Context) uint {
	if v, exists := c.Get("companyID"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) uint {
	if v, exists := c.Get("userID"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetDeptID 从上下文获取部门 ID
func GetDeptID(c *gin.Context) uint {
	if v, exists := c.Get("deptID"); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
