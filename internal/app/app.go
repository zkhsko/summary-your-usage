package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"summary-your-usage/config"
	httpcontroller "summary-your-usage/internal/controller/http"
	"summary-your-usage/internal/repo/persistent"
	"summary-your-usage/internal/usecase"
	"summary-your-usage/migrations"
	"summary-your-usage/pkg/database"
	"summary-your-usage/ui"
)

func Run(ctx context.Context, cfg config.Config) error {
	db, err := database.Open(ctx, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	provider, err := migrations.New(db, cfg.Database.Driver)
	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}
	migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	_, err = provider.Up(migrationCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	users := usecase.NewUser(persistent.NewUser(db))
	router := httpcontroller.NewRouter(users)
	router.Handle("/*", ui.Handler())

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	log.Printf("服务已启动：http://%s", server.Addr)

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP: %w", err)
		}
		<-serveErr
	}
	log.Print("服务已停止")
	return nil
}
