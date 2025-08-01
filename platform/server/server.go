package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (s *Server) Run(port string, gracefulTime time.Duration) <-chan error {

	errC := make(chan error, 1)
	s.initGracefulShutdown(gracefulTime, errC)

	go func() {
		if err := s.Echo.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()

	return errC
}

func (s *Server) initGracefulShutdown(gracefulTime time.Duration, errChannel chan error) {

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-ctx.Done()
		ctxTimeout, cancel := context.WithTimeout(context.Background(), gracefulTime+(1*time.Second))
		defer func() {
			time.AfterFunc(gracefulTime, func() {
				s.Echo.Server.SetKeepAlivesEnabled(false)
				if err := s.Echo.Shutdown(ctxTimeout); err != nil {
					errChannel <- err
				}
				cancel()
				stop()
				close(errChannel)
			})
		}()

	}()
}
