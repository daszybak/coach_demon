package main

import (
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"coach_demon/internal/app"
	"coach_demon/internal/fetcher"
	"coach_demon/internal/openai"
	"coach_demon/internal/server"
	"coach_demon/internal/storage"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("couldn't read config file, falling back to env only")
	}
}

func main() {
	// Pretty console output if in dev mode
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	initConfig()

	// Setup logger with some defaults (ISO timestamp)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	mStore, err := storage.NewMongoManager(viper.GetString("MONGODB_URI"), &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("storage setup failed")
	}

	aiCfg := openai.Config{
		APIKey:       viper.GetString("OPENAI_API_KEY"),
		Model:        viper.GetString("OPENAI_MODEL"),
		SystemPrompt: viper.GetString("OPENAI_SYSTEM_PROMPT"),
	}
	httpClient := &http.Client{}
	aiClient, err := openai.NewClient(aiCfg, httpClient)
	if err != nil {
		logger.Fatal().Err(err).Msg("openai client setup failed")
	}

	fetchURL := viper.GetString("FETCHER_ENDPOINT")
	if fetchURL == "" {
		logger.Fatal().Msg("FETCHER_ENDPOINT missing in config")
	}
	fetchTok := viper.GetString("FETCHER_TOKEN")
	fetchSvc := fetcher.NewBrowserless(fetchURL, fetchTok)

	appCtx := &app.App{
		Store:  mStore,
		AI:     aiClient,
		Fetch:  fetchSvc,
		Logger: &logger,
	}

	addr := ":" + viper.GetString("PORT")
	if addr == ":" {
		addr = ":12345"
	}
	logger.Info().Str("addr", addr).Msg("🚀 Coach server starting")

	if err := http.ListenAndServe(addr, server.New(appCtx)); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}
}
