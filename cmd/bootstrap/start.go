package bootstrap

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/gorm"
	"tenant-crud/cmd/server"
	iamContainer "tenant-crud/internal/iam/domain/di"
	"tenant-crud/internal/infra/database/postgres"
)

// Application armazena as dependências centrais da aplicação.
type Application struct {
	container *iamContainer.Container
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

func initContainer(db *gorm.DB) *iamContainer.Container {
	return iamContainer.NewContainer(db)
}

// New prepara a aplicação (config, db, di) e retorna a instância.
func New() (*Application, error) {
	Environment()
	log.Println("[BOOTSTRAP] Configuração de ambiente carregada.")

	db := postgres.InitPostgres()
	log.Println("[BOOTSTRAP] Conexão com o banco de dados inicializada.")

	container := initContainer(db)
	log.Println("[BOOTSTRAP] Contêiner de dependências inicializado.")

	return &Application{
		container: container,
	}, nil
}

func (a *Application) Start() {
	log.Println("[BOOTSTRAP] Iniciando servidor no ambiente:", viper.GetString("environment"))
	server.StartServer(a.container)
}
