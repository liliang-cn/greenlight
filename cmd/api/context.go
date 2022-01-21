package main

import (
	"context"
	"net/http"

	"github.com/liliang-cn/greenlight/internal/data"
)

// 基于string定义一个contextType
type contextKey string

// 定义一个常量用来从请求的context中获取/操作用户信息
const userContextKey = contextKey("user")

// contextSetUser 返回一个复制的 request，里面包含添加了 User 结构体的 context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser 从 context 中取 User
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
