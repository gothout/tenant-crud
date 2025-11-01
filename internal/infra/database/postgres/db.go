package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"

	"log"
	"sync"
)

var (
	db   *sql.DB
	once sync.Once
)

// InitPostgres inicializa a conexão com o banco de dados PostgreSQL.
// A função utiliza sync.Once para garantir que a conexão seja criada apenas uma vez.
func InitPostgres() *sql.DB {
	once.Do(func() {
		dsn := buildDSN()
		var err error

		db, err = sql.Open("pgx", dsn)
		if err != nil {
			log.Fatalf("[DATABASE] erro ao abrir conexão com o banco de dados: %v", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("[DATABASE] erro ao testar conexão com o banco de dados: %v", err)
		}

		log.Println("[DATABASE] Conexão com o banco de dados PostgreSQL estabelecida com sucesso.")
	})

	return db
}

// GetDB retorna a instância atual da conexão com o banco de dados.
// Gera erro fatal se InitPostgres() não tiver sido chamado anteriormente.
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("[DATABASE] a conexão com o banco de dados não foi inicializada. Chame InitPostgres() primeiro.")
	}
	return db
}

// buildDSN monta a string de conexão (Data Source Name) para o PostgreSQL
// utilizando as variáveis de ambiente definidas no pacote env.
func buildDSN() string {
	host := viper.GetString("databases.postgres.host")
	port := viper.GetInt("databases.postgres.port")
	user := viper.GetString("databases.postgres.user")
	pass := viper.GetString("databases.postgres.pwd")
	name := viper.GetString("databases.postgres.db_name")
	if name == "" {
		name = "appdb"
	}
	ssl := viper.GetString("databases.postgres.ssl_mode")
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, ssl,
	)
}
