package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juanMaAV92/go-utils/env"
	"github.com/juanMaAV92/go-utils/platform/config"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

func New(config *config.BasicConfig) (*Server, error) {
	server := echo.New()
	server.HideBanner = true
	server.Debug = config.Environment == env.LocalEnvironment
	decimal.MarshalJSONWithoutQuotes = true

	return &Server{
		Instance:    server,
		BasicConfig: config,
	}, nil
}

func (s *Server) Run() <-chan error {

	errC := make(chan error, 1)
	s.initGracefulShutdown(errC)

	go func() {
		if err := s.Instance.Start(":" + s.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()

	return errC
}

func (s *Server) initGracefulShutdown(errChannel chan error) {

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-ctx.Done()
		ctxTimeout, cancel := context.WithTimeout(context.Background(), s.GracefullTime+(1*time.Second))
		defer func() {
			time.AfterFunc(s.GracefullTime, func() {
				s.Instance.Server.SetKeepAlivesEnabled(false)
				if err := s.Instance.Shutdown(ctxTimeout); err != nil {
					errChannel <- err
				}
				cancel()
				stop()
				close(errChannel)
			})
		}()

	}()
}
