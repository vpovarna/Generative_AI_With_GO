package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/batch"
	"github.com/povarna/generative-ai-agents/eval-agent/internal/setup"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	startTime := time.Now()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	input := flag.String("input", "", "Input file relative path")
	output := flag.String("output", "", "Output file relative path")
	format := flag.String("format", "jsonl", "Output file format. Supported formats: 'jsonl', 'summary'")
	summary := flag.String("summary", "", "Optional separate summary file")
	workers := flag.Int("workers", 5, "Concurrent evaluators workers")
	continueOnError := flag.Bool("continue-on-error", true, "Continue on evaluation failures")
	dryRun := flag.Bool("dry-run", false, "Validate input without evaluating")
	// progressInterval := flag.Int("progress-interval", 5, "Progress log interval (seconds)")

	flag.Parse()

	if *input == "" {
		log.Fatal().Msg("required flag -input not provided")
	}
	formatValidator(format)

	if err := godotenv.Load(); err != nil {
		log.Warn().Msg("No .env file found, using environment variables")
	}

	ctx, cancel := setupGracefulShutdown()
	defer cancel()

	cfg := setup.LoadConfig()

	deps, err := setup.Wire(ctx, cfg, &log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to wire dependencies")
	}

	// Open input file
	var inputFile io.Reader
	if *input == "-" {
		inputFile = os.Stdin
		log.Info().Msg("Reading from stdin")
	} else {
		f, err := os.Open(*input)
		if err != nil {
			log.Fatal().Err(err).Str("file", *input).Msg("Failed to open input file")
		}
		defer f.Close()
		inputFile = f
		log.Info().Str("file", *input).Msg("Reading input file")
	}

	// Read records
	reader := batch.NewReader(inputFile, deps.Logger)
	recordsCh := reader.ReadAll(ctx)

	var records []batch.InputRecord
	for record := range recordsCh {
		records = append(records, record)
	}

	log.Info().Int("total", len(records)).Msg("Input file parsed")

	// Dry run validation
	if *dryRun {
		dryRunAndExit(records)
	}

	// Open output file
	var outputFile io.Writer
	if *output == "" {
		outputFile = os.Stdout
		log.Info().Msg("Writing to stdout")
	} else {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatal().Err(err).Str("file", *output).Msg("Failed to create output file")
		}
		defer f.Close()
		outputFile = f
		log.Info().Str("file", *output).Msg("Writing to output file")
	}

	// Create writer
	writer, err := batch.NewWriter(outputFile, *format, deps.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create writer")
	}
	defer writer.Close()

	// Process with worker pool
	processor := batch.NewProcessor(deps.Executor, *workers, deps.Logger)
	results := processor.Process(ctx, records)

	// Write results
	successCount := 0
	errorCount := 0

	for result := range results {
		if err := writer.Write(result); err != nil {
			log.Error().Err(err).Str("id", result.ID).Msg("Failed to write result")
			errorCount++

			if !*continueOnError {
				log.Fatal().Msg("Stopping due to write error")
			}
		} else {
			successCount++
		}
	}

	log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Dur("duration", time.Since(startTime)).
		Msg("Processing complete")

	if *summary != "" {
		writeSummary(summary)
	}

	log.Info().Msg("Batch processing complete")
}

func setupGracefulShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Warn().Msg("Received interrupt signal, finishing current work...")
		cancel()
	}()

	return ctx, cancel
}

func formatValidator(format *string) {
	validFormats := map[string]bool{"jsonl": true, "summary": true}
	if !validFormats[*format] {
		log.Fatal().
			Str("format", *format).
			Msg("Invalid format. Supported: jsonl, summary")
	}
}

func writeSummary(summary *string) {
	summaryFile, err := os.Create(*summary)
	if err != nil {
		log.Fatal().Err(err).Str("file", *summary).Msg("Failed to create summary file")
	}
	defer summaryFile.Close()

	// TODO: Write summary stats (can reuse summary writer logic)
	log.Info().Str("file", *summary).Msg("Summary written")
}

func dryRunAndExit(records []batch.InputRecord) {
	errorCount := 0
	for _, record := range records {
		if record.Error != nil {
			log.Error().
				Int("line", record.LineNumber).
				Err(record.Error).
				Msg("Validation error")
			errorCount++
		}
	}

	if errorCount > 0 {
		log.Fatal().Int("errors", errorCount).Msg("Validation failed")
	}

	log.Info().Msg("Validation successful")
	os.Exit(0)
}
