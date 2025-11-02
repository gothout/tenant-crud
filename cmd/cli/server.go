package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"tenant-crud/cmd/bootstrap"
	"tenant-crud/internal/infra/database/postgres"
	"tenant-crud/internal/pkg/system"
)

const pidFile = "run/server.pid"

func startServer() error {
	app, err := bootstrap.New()
	if err != nil {
		return fmt.Errorf("não foi possível criar a aplicação: %w", err)
	}

	if err := system.SavePID(pidFile, os.Getpid()); err != nil {
		return err
	}
	defer system.RemovePID(pidFile)
	defer postgres.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Iniciando servidor...")
	return app.Start(ctx)
}

func stopServer() error {
	pid, err := system.LoadPID(pidFile)
	if err != nil {
		return err
	}

	if err := system.TerminateProcess(pid); err != nil {
		return err
	}

	system.RemovePID(pidFile)
	return nil
}
