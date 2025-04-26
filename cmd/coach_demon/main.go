package main

import (
	"log"
	"net/http"
	"os"

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
		log.Printf("warning: couldn't read config file, falling back to env only: %v", err)
	}
}

func main() {
	initConfig()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	mStore, err := storage.NewMongoManager(viper.GetString("MONGODB_URI"), logger)
	if err != nil {
		logger.Fatalf("storage: %v", err)
	}

	aiCfg := openai.Config{
		APIKey:       viper.GetString("OPENAI_API_KEY"),
		Model:        viper.GetString("OPENAI_MODEL"),
		SystemPrompt: viper.GetString("OPENAI_MODEL"),
	}
	httpClient := &http.Client{}
	aiClient, err := openai.NewClient(aiCfg, httpClient)
	if err != nil {
		logger.Fatalf("openai: %v", err)
	}

	fetchURL := viper.GetString("FETCHER_ENDPOINT")
	if fetchURL == "" {
		logger.Fatal("FETCHER_ENDPOINT missing in config")
	}
	fetchTok := viper.GetString("FETCHER_TOKEN")
	fetchSvc := fetcher.NewBrowserless(fetchURL, fetchTok)

	appCtx := &app.App{
		Store:  mStore,
		AI:     aiClient,
		Fetch:  fetchSvc,
		Logger: logger,
	}

	addr := ":" + viper.GetString("PORT")
	if addr == ":" {
		addr = ":12345"
	}
	logger.Printf("🚀 Coach listening on %s", addr)

	if err := http.ListenAndServe(addr, server.New(appCtx)); err != nil {
		logger.Fatalf("http: %v", err)
	}
}
