package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 使用 shutDownError 通道来接收 Shutdown() 函数返回的错误
	shutdownError := make(chan error)

	go func() {
		// 新建 channel 用来携带系统信号
		quit := make(chan os.Signal, 1)

		// 使用signal.Notify()来监听传入的SIGINT和SIGTERM信号，并将它们转发到通道 channel，其他信号都不会被signal.Notify()捕捉到
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// 从channel中读值, 会阻塞直至读取到值
		s := <-quit

		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
