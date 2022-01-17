package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// recoverPanic 从panic恢复
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// 限流
func (app *application) rateLimiter(next http.Handler) http.Handler {
	// 全局限流
	// limiter := rate.NewLimiter(2, 4)
	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	// 调用 limiter.Allow() 来确定是否允许请求, 如果不允许，返回 429
	// 	if !limiter.Allow() {
	// 		app.rateLimitExceededResponse(w, r)
	// 		return
	// 	}

	// 	next.ServeHTTP(w, r)
	// })

	// 定义一个客户结构体用来存放 限流器和最近一次使用时间
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// 定时一分钟移除所有老的条目
	go func() {
		for {
			time.Sleep(time.Minute)

			// 加锁避免在清理时限流器做检查
			mu.Lock()

			// 遍历客户端，如果过去的三分钟没有使用，将其移除
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	// return 之前的代码只会执行一次
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// 处理中间件的每个请求都会执行next.ServeHTTP(w, r)之前的代码
			// 从请求中提取客户端的IP地址
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			// 检查IP地址是否在map中，如果不在，初始化一个新的limiter 并将该IP地址添加到map中
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}

			clients[ip].lastSeen = time.Now()

			// 检查当前IP的Allow()方法, 如果不允许，将mutext锁解除并返回429
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// 在这个中间件下游的所有处理程序都返回之前，mutex不会被解锁
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
