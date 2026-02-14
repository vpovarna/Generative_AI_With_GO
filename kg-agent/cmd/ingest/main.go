package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/database"
	"github.com/rs/zerolog/log"
)

func main() {
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

}
