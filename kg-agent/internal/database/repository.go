package database

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (db *DB) DeleteDocument(ctx context.Context, docId string) error {

	query := `DELETE FROM documents WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, docId)
	if err != nil {
		return fmt.Errorf("Failed to delete document id: %s, error: %w", docId, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Warn().Str("doc_id", docId).Msg("Document not found")
	} else {
		log.Info().Str("doc_id", docId).Msg("Document deleted")
	}

	return nil
}

func (db *DB) GetAllDocs(ctx context.Context) ([]DocumentEntityResponse, error) {
	query := `SELECT id, title from documents`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Unable to fetch document ids from DB")
	}

	defer rows.Close()

	var documentsResponse []DocumentEntityResponse

	for rows.Next() {
		var document DocumentEntityResponse

		if err := rows.Scan(&document.Id, &document.Title); err != nil {
			return nil, fmt.Errorf("Failed to scan id: %w", err)
		}

		documentsResponse = append(documentsResponse, document)
	}

	return documentsResponse, nil
}
