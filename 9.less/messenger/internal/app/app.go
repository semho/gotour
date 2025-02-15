package app

import (
	"fmt"
	"log"
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
	db := memory.NewDB(cfg.Storage.MaxUsers, cfg.Storage.MaxMessages, cfg.Storage.MaxChats)

	services := service.NewServices(db, cfg)
	handlers := handler.NewHandlers(services)

	mux := http.NewServeMux()
	// Регистрация маршрутов
	handlers.User.RegisterRoutes(mux)
	handlers.Chat.RegisterRoutes(mux)
	handlers.Message.RegisterRoutes(mux)

	return &App{
		cfg:      cfg,
		db:       db,
		services: services,
		handlers: handlers,
		mux:      mux,
	}, nil
}

func (a *App) Run() error {
	host := fmt.Sprintf("%s", a.cfg.Server.Host)
	addr := fmt.Sprintf(":%s", a.cfg.Server.Port)
	log.Printf("Starting server on %s%s", host, addr)

	server := &http.Server{
		Addr:    addr,
		Handler: a.mux,
	}
	return server.ListenAndServe()
}
