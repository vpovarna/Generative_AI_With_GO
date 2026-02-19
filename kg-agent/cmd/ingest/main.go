package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/embedding"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/ingestion"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// TODO: replace this with cobra cli
	insertDocCommand := flag.Bool("insert-doc", false, "Insert document command")
	filePath := flag.String("filePath", "resources/test-input.txt", "Relative path to the document")
	chunkSize := flag.Int("chunkSize", 500, "Chunk size")
	chunkOverlap := flag.Int("chunkOverlap", 100, "Chunk overlap")
	customMetadata := flag.String("customMetadata", "{}", "Custom metadata")

	deleteDocCommand := flag.Bool("delete-doc", false, "Delete existing document command")
	documentId := flag.String("doc-id", "", "Document id which needs to be delete")

	getAllDocsCommand := flag.Bool("get-docs", false, "Get all documents command")

	flag.Parse()

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

	log.Info().Msg("Database connected")

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)

	if err != nil {
		log.Error().Err(err).Msg("Unable to create bedrock client")
		return
	}

	parser := ingestion.NewParser()
	chunker := ingestion.NewChunker(*chunkSize, *chunkOverlap)
	embedder := embedding.NewBedrockEmbedder(bedrockClient.Client)
	pipeline := ingestion.NewPipeline(parser, chunker, embedder, db.Pool)

	// Input commands parsing
	if *deleteDocCommand {
		// Delete a document by id
		err := db.DeleteDocument(ctx, *documentId)
		if err != nil {
			log.Error().Err(err).Msg("failed to delete document")
		}
		log.Info().Msg("Document delete successfully")
	} else if *getAllDocsCommand {
		// Get all documents and return ids
		documentsResponse, err := db.GetAllDocs(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to fetch documents from DB!")
		}

		for _, documentResponse := range documentsResponse {
			log.Info().Msg(documentResponse.Print())

		}
	} else if *insertDocCommand {
		// Ingest document (atomic operation)
		var customMetadataMap map[string]string
		err := json.Unmarshal([]byte(*customMetadata), &customMetadataMap)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to unmarshal custom metadata")
		}
		if err := pipeline.IngestDocument(ctx, *filePath, customMetadataMap); err != nil {
			log.Fatal().Err(err).Msg("Ingestion failed")
		}
		log.Info().Msg("Ingestion successful!")
	} else {
		log.Fatal().Msg("Unsupported command")
	}
}
