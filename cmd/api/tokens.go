package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/liliang-cn/greenlight/internal/data"
	"github.com/liliang-cn/greenlight/internal/validator"
)

// createAuthenticationTokenHandler 创建认证 Token
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求中的邮箱和密码
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 校验邮箱和密码
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 根据邮箱获取用户，如果找不到则返回 401 错误
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 检查提供的密码是否与真实的密码匹配
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 密码不匹配
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// 密码匹配，生成新的 Token
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 将生成的 Token 返回
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
