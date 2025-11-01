package bootstrap

import (
	"fmt"
	"tenant-crud/internal/infra/database/postgres"

	"github.com/spf13/viper"
)

type Application struct {
}

func Environment() {
	viper.SetConfigName("configs")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error in configuration file: %w\n", err))
	}
}

func New() (*Application, error) {
	Environment()
	postgres.InitPostgres()
	fmt.Println(viper.GetString("environment"))
	return &Application{}, nil
}

func (a *Application) Start() {
}
