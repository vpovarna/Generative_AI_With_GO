package main

import (
	"context"
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/embedding"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/ingestion"
	"github.com/rs/zerolog/log"
)

func main() {
	filePath := flag.String("filePath", "resources/test-input.txt", "Relative path to the document")

	err := godotenv.Load()

	if err != nil {
		log.Warn().Msg("Unable to load env variables")
	}

	ctx := context.Background()

	config := database.Config{
		Host:     os.Getenv("KG_AGENT_VECTOR_DB_HOST"),
		Port:     os.Getenv("KG_AGENT_VECTOR_DB_PORT"),
		User:     os.Getenv("KG_AGENT_VECTOR_DB_USER"),
		Password: os.Getenv("KG_AGENT_VECTOR_DB_PASSWORD"),
		Database: os.Getenv("KG_AGENT_VECTOR_DB_DATABASE"),
		SSLMode:  os.Getenv("KG_AGENT_VECTOR_DB_SSLMode"),
	}

	db, err := database.NewWithBackoff(ctx, config, 3)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return
	}

	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Database ping failed")
		return
	}

	log.Info().Msg("Connected successfully")

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)

	if err != nil {
		log.Error().Err(err).Msg("Unable to create bedrock client")
		return
	}

	embedder := embedding.NewBedrockEmbedder(bedrockClient.Client)

	// Create document parser
	parser := ingestion.NewParser()
	doc, err := parser.ParseFile(*filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse input file.")
	}

	// Insert Document
	err = db.InsertDocument(ctx, *doc)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to insert document.")
	}
	log.Info().Msg("Document inserted successfully")

	// Create chunks
	chunker := ingestion.NewChunker(50, 10)
	chunks := chunker.ChunkText(doc.Content)
	log.Info().Msg("Chunks created successfully")

	// Generate Embeddings
	var chunkContents []string
	for _, chunk := range chunks {
		chunkContents = append(chunkContents, chunk.Content)
	}

	embeddings, err := embedder.GenerateBatchEmbeddings(ctx, chunkContents)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to generate embeddings.")
	}

	log.Info().Msg("Embeddings generated successfully")

	err = db.InsertChunks(ctx, doc.ID, chunks, embeddings)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to insert chunks")
	}

	log.Info().Int("chunks_inserted", len(chunks)).Msg("Ingestion complete!")

}
