package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lanping/lp-server/internal/auth"
	"github.com/lanping/lp-server/internal/response"
)

const CtxClaims = "claims"

func JWT(m *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			response.Abort(c, http.StatusUnauthorized, "未登录")
			return
		}
		claims, err := m.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			response.Abort(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			return
		}
		c.Set(CtxClaims, claims)
		c.Next()
	}
}

func Roles(roles ...string) gin.HandlerFunc {
	set := map[string]struct{}{}
	for _, r := range roles {
		set[r] = struct{}{}
	}
	return func(c *gin.Context) {
		v, ok := c.Get(CtxClaims)
		if !ok {
			response.Abort(c, http.StatusUnauthorized, "未登录")
			return
		}
		claims := v.(*auth.Claims)
		if _, ok := set[claims.Role]; !ok {
			response.Abort(c, http.StatusForbidden, "无权访问")
			return
		}
		c.Next()
	}
}

func GetClaims(c *gin.Context) *auth.Claims {
	v, _ := c.Get(CtxClaims)
	if v == nil {
		return nil
	}
	return v.(*auth.Claims)
}
