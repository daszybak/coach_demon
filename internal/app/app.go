package app

import (
	"log"

	"coach_demon/internal/fetcher"
	"coach_demon/internal/openai"
	"coach_demon/internal/storage"
)

type App struct {
	Store  storage.Storage
	AI     *openai.Client
	Fetch  fetcher.Service
	Logger *log.Logger
}
