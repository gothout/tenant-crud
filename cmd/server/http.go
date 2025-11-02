package server

import (
	"fmt"
	"log"
	"tenant-crud/cmd/server/routes"
	"tenant-crud/internal/iam/domain/di"

	"github.com/spf13/viper"
)

// StartServer inicializa o roteador e inicia o servidor HTTP.
func StartServer(container *di.Container) {
	router := routes.SetupRouter(container)
	log.Println("[SERVER] Roteador HTTP configurado.")
	port := fmt.Sprintf(":%s", viper.GetString("server.http.port"))
	log.Printf("[SERVER] Iniciando servidor na porta %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("[SERVER] Falha ao iniciar o servidor: %v", err)
	}
}
