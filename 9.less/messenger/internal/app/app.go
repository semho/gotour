package app

import (
	"messenger/internal/config"
	"messenger/internal/handler"
	"messenger/internal/service"
	"messenger/internal/storage/memory"
	"net/http"
)

type App struct {
	cfg      *config.Config
	db       *memory.DB
	services *service.Services
	handlers *handler.Handlers
	mux      *http.ServeMux
}

func New(cfg *config.Config) (*App, error) {
	db := memory.NewDB()

	services := service.NewServices(db)
	handlers := handler.NewHandlers(services)

	mux := http.NewServeMux()
	mux.Handle("/users", handlers.User)
	mux.Handle("/users/", handlers.User)
	mux.Handle("/chats", handlers.Chat)
	mux.Handle("/chats/", handlers.Chat)
	// другие маршруты

	return &App{
		cfg:      cfg,
		db:       db,
		services: services,
		handlers: handlers,
		mux:      mux,
	}, nil
}

func (a *App) Run() error {
	server := &http.Server{
		Addr:    ":" + a.cfg.Server.Port,
		Handler: a.mux,
	}
	return server.ListenAndServe()
}
