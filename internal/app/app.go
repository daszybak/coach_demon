package app

import (
	"coach_demon/internal/fetcher"
	"coach_demon/internal/openai"
	"coach_demon/internal/storage"
	"github.com/rs/zerolog"
)

type App struct {
	Store  storage.Storage
	AI     *openai.Client
	Fetch  fetcher.Service
	Logger *zerolog.Logger
}
