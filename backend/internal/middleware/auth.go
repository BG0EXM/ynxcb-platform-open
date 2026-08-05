package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"ynxcb-platform/internal/auth"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextUsername contextKey = "username"
	ContextRealName contextKey = "real_name"
	ContextRoleCode contextKey = "role_code"
)

// Auth 认证中间件，校验 JWT
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"未登录"}`, http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"认证格式错误"}`, http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"登录已过期，请重新登录"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextUsername, claims.Username)
		ctx = context.WithValue(ctx, ContextRealName, claims.RealName)
		ctx = context.WithValue(ctx, ContextRoleCode, claims.RoleCode)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole 角色校验中间件
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleCode, _ := r.Context().Value(ContextRoleCode).(string)
			for _, role := range roles {
				if role == roleCode {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"无权访问"}`, http.StatusForbidden)
		})
	}
}

// JSON 输出 JSON 响应
func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
