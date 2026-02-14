package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/ingestion"
)

func (db *DB) InsertDocument(ctx context.Context, doc ingestion.Document) error {
	query := `INSERT INTO documents(id, title, content, metadata, created_at, updated_at) VALUES($1, $2, $3, $4, NOW(), NOW())`

	metadataJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal document. Error: %w", err)
	}

	_, err = db.Pool.Exec(ctx, query, doc.ID, doc.Title, doc.Content, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert document. Error: %w", err)
	}

	return nil
}

func (db *DB) InsertChunk(ctx context.Context, documentID string, chunk ingestion.Chunk, embedding []float32) error {

	query := `INSERT INTO chunks (id, document_id, chunk_index, content, embedding, metadata, created_at) VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())`

	metadata := map[string]any{
		"start": chunk.Start,
		"end":   chunk.End,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata. Error: %w", err)
	}

	// Convert []float32 to pgvector.Vector
	vector := pgvector.NewVector(embedding)

	_, err = db.Pool.Exec(ctx, query, documentID, chunk.Index, chunk.Content, vector, metadataJSON)

	if err != nil {
		return fmt.Errorf("unable to insert chunks. Error: %w", err)
	}

	return nil
}

func (db *DB) InsertChunks(ctx context.Context, documentID string, chunks []ingestion.Chunk, embeddings [][]float32) error {
	// Use a transaction for atomicity
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if we don't commit

	query := `
        INSERT INTO document_chunks (id, document_id, chunk_index, content, embedding, metadata, created_at)
        VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())
    `

	for i, chunk := range chunks {
		metadata := map[string]any{
			"start": chunk.Start,
			"end":   chunk.End,
		}
		metadataJSON, _ := json.Marshal(metadata)

		// Convert []float32 to pgvector.Vector
		vector := pgvector.NewVector(embeddings[i])

		_, err := tx.Exec(ctx, query,
			documentID,
			chunk.Index,
			chunk.Content,
			vector,
			metadataJSON,
		)

		if err != nil {
			return fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
