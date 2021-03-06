package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/liliang-cn/greenlight/internal/validator"
)

// readIDParam 从 URL 中读取 id
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// 将 json 内容添加一个父级字段, 结构化返回 json
type envelope map[string]interface{}

// writeJSON 将 JSON 写入响应中
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// readJSON 读取请求中的 JSON
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// 使用 http.MaxBytesReader() 来限制请求体大小不超过1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// 初始化 json.Decoder 后在 decode 前调用 DisallowUnknownFields() 方法
	// 如果请求中包含不能被赋值给目标的字段，将会抛出错误，而不是忽略多余字段
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	// 再次调用 Decode() 方法, 检查是否只有一个 json 对象
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString 从 url 的 querystring 中提取字符
func (app *application) readString(qs url.Values, key string, defaultString string) string {
	// 没有值的话会返回 ""
	s := qs.Get(key)

	if s == "" {
		return defaultString
	}

	return s
}

// readCSV 从 url 的 querystring 中提取逗号分隔的值
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readInt 从 url 的 querystring 中提取数值, 如果不能转换，Validator 中提示
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// background 通过 goroutine 执行传入函数，若发生 panic 中会恢复
func (app *application) background(fn func()) {
	// WaitGroup 计数加1
	app.wg.Add(1)
	// 运行后台 goroutine
	go func() {
		// goroutine 退出前 WaitGroup 计数减1
		defer app.wg.Done()
		// 恢复 panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// 执行传入函数
		fn()
	}()
}
