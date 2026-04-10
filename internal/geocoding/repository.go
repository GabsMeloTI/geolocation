package geocoding

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

// RepositoryInterface define os métodos necessários para interagir com a persistência de geocodificação
type RepositoryInterface interface {
	// ConsultCEP busca as informações de um CEP específico na base local (UniqueCEP)
	ConsultCEP(ctx context.Context, arg string) (db.UniqueCep, error)
}

// Repository implementa a interface RepositoryInterface usando SQLC
type Repository struct {
	conn    *sql.DB
	dbTX    db.DBTX
	queries *db.Queries
	sqlConn *sql.DB
}

// NewRepository cria uma nova instância do repositório de geocodificação
func NewRepository(conn *sql.DB) *Repository {
	q := db.New(conn)
	return &Repository{
		conn:    conn,
		dbTX:    conn,
		queries: q,
		sqlConn: conn,
	}
}

// ConsultCEP executa a query ConsultUniqueCEP para retornar os dados do CEP da base local
func (r *Repository) ConsultCEP(ctx context.Context, arg string) (db.UniqueCep, error) {
	return r.queries.ConsultUniqueCEP(ctx, arg)
}
