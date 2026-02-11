package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	modelID := os.Getenv("CLAUDE_MODEL_ID")

	bedrockClient, err := bedrock.NewClient(ctx, region, modelID)

	if err != nil {
		log.Fatal(err)
	}

	// Invoke client
	response, err := bedrockClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      "What is Go Programming Language",
		MaxTokens:   200,
		Temperature: 0.1,
	})

	if err != nil {
		log.Fatalf("Unable to invoke Claude model: %v", err)
	}

	fmt.Printf("Claude response is: %s\n", response.Content)

	log.Println("Finish!!")

}
