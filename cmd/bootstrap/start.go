package bootstrap

import (
	"context"
	"fmt"
	"log"
	"tenant-crud/internal/infra/jwt"
	"time"

	"tenant-crud/cmd/server"
	iamContainer "tenant-crud/internal/iam/di"
	"tenant-crud/internal/infra/database/postgres"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Application armazena as dependências centrais da aplicação.
type Application struct {
	container *iamContainer.Container
	server    *server.HTTPServer
}

// Environment configura e lê o arquivo de configuração (configs.json)
func Environment() {
	viper.SetConfigName("configs")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/") // Para ambientes de produção
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error in configuration file: %w", err))
	}
}

func initContainer(db *gorm.DB, jwtInstance *jwt.TokenGenerator) *iamContainer.Container {
	return iamContainer.NewContainer(db, jwtInstance)
}

// New prepara a aplicação (config, db, di) e retorna a instância.
func New() (*Application, error) {
	Environment()
	log.Println("[BOOTSTRAP] Configuração de ambiente carregada.")
	jwtConfig := jwt.Config{
		AccessSecret:  viper.GetString("security.jwt_access_secret"),
		RefreshSecret: viper.GetString("security.jwt_refresh_secret"),
		Issuer:        viper.GetString("app.name"),
		AccessExpiry:  time.Duration(viper.GetInt64("security.jwt_access_expiry_min")) * time.Minute,
	}

	tokenGen, err := jwt.NewTokenGenerator(jwtConfig)
	if err != nil {
		// Erro fatal, a aplicação não pode subir sem o gerador de token
		return nil, fmt.Errorf("falha ao criar gerador de token: %w", err)
	}
	log.Println("[BOOTSTRAP] Gerador de token inicializado.")
	db := postgres.InitPostgres()
	log.Println("[BOOTSTRAP] Conexão com o banco de dados inicializada.")

	container := initContainer(db, tokenGen)
	log.Println("[BOOTSTRAP] Contêiner de dependências inicializado.")

	return &Application{
		container: container,
		server:    server.NewHTTPServer(container),
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	log.Println("[BOOTSTRAP] Iniciando servidor no ambiente:", viper.GetString("environment"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.server.Start()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("falha ao encerrar servidor: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}
