package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/pflag"
	"gorm.io/gorm"

	"tenant-crud/cmd/bootstrap"
	"tenant-crud/internal/infra/database/admin"
	"tenant-crud/internal/infra/database/migrations"
	"tenant-crud/internal/infra/database/postgres"
	"tenant-crud/internal/pkg/system"
)

const pidFile = "run/server.pid"

func main() {
	var (
		flagStart         bool
		flagStop          bool
		flagSeed          bool
		flagUpdate        bool
		flagDBCheck       bool
		flagDBDelete      bool
		flagDBBackup      bool
		backupDestination string
	)

	pflag.BoolVar(&flagStart, "start", false, "Inicia o servidor HTTP")
	pflag.BoolVar(&flagStop, "stop", false, "Finaliza o servidor HTTP")
	pflag.BoolVar(&flagSeed, "migration-seed", false, "Aplica migrations de seed")
	pflag.BoolVar(&flagUpdate, "migration-update", false, "Aplica migrations de atualização")
	pflag.BoolVar(&flagDBCheck, "db-check", false, "Checa status do banco de dados")
	pflag.BoolVar(&flagDBDelete, "db-delete", false, "Remove todas as tabelas do banco de dados")
	pflag.BoolVar(&flagDBBackup, "db-backup", false, "Realiza backup do banco de dados")
	pflag.StringVar(&backupDestination, "local", "", "Diretório de destino para o backup do banco")
	pflag.Parse()

	bootstrap.Environment()

	if flagStop {
		if err := stopServer(); err != nil {
			log.Fatalf("falha ao parar servidor: %v", err)
		}
		log.Println("Servidor finalizado com sucesso.")
		return
	}

	var dbInitialized bool
	var dbInstance = func() *gorm.DB {
		if !dbInitialized {
			db := postgres.InitPostgres()
			dbInitialized = true
			return db
		}
		return postgres.GetDB()
	}

	defer func() {
		if dbInitialized {
			postgres.Close()
		}
	}()

	manager := func() *migrations.Manager {
		return migrations.NewManager(dbInstance())
	}

	operationsExecuted := false

	if flagSeed {
		if err := manager().ApplySeed(); err != nil {
			log.Fatalf("falha ao aplicar migrations de seed: %v", err)
		}
		log.Println("Migrations de seed aplicadas com sucesso.")
		operationsExecuted = true
	}

	if flagUpdate {
		if err := manager().ApplyUpdate(); err != nil {
			log.Fatalf("falha ao aplicar migrations de atualização: %v", err)
		}
		log.Println("Migrations de atualização aplicadas com sucesso.")
		operationsExecuted = true
	}

	if flagDBCheck {
		status, err := admin.Check(dbInstance())
		if err != nil {
			log.Fatalf("falha ao checar banco de dados: %v", err)
		}
		log.Printf("Banco de dados ativo. Tabelas encontradas (%d): %v", len(status.Tables), status.Tables)
		operationsExecuted = true
	}

	if flagDBDelete {
		if err := admin.DeleteAll(dbInstance()); err != nil {
			log.Fatalf("falha ao deletar tabelas do banco: %v", err)
		}
		log.Println("Todas as tabelas foram removidas com sucesso.")
		operationsExecuted = true
	}

	if flagDBBackup {
		if backupDestination == "" {
			log.Fatal("para executar o backup informe o destino com --local=<caminho>")
		}
		dest := backupDestination
		if !filepath.IsAbs(dest) {
			abs, err := filepath.Abs(dest)
			if err == nil {
				dest = abs
			}
		}
		if err := admin.Backup(admin.BackupOptions{Destination: dest}); err != nil {
			log.Fatalf("falha ao executar backup: %v", err)
		}
		log.Printf("Backup gerado em %s", dest)
		operationsExecuted = true
	}

	if !operationsExecuted && !flagStart {
		flagStart = true
	}

	if flagStart {
		if err := startServer(); err != nil {
			log.Fatalf("falha ao iniciar servidor: %v", err)
		}
	}
}

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
