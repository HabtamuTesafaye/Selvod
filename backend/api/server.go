package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/selvod/selvod/handler"
	"github.com/selvod/selvod/hooks"
	"github.com/selvod/selvod/queue"
	"github.com/selvod/selvod/signer"
	"github.com/selvod/selvod/store"
	"github.com/selvod/selvod/storage"
	"github.com/selvod/selvod/transcoder"
	customMiddleware "github.com/selvod/selvod/middleware"
)

type Server struct {
	router      *chi.Mux
	httpServer  *http.Server
	handler     *handler.VideoHandler
	apiKey      string
	playbackKey string
}

func NewServer(s store.Store, st storage.Provider, sig *signer.SecureSigner, q *queue.WorkerPool, h *hooks.Registry, t transcoder.Transcoder, storageDir, apiKey, playbackKey string) *Server {
	vh := &handler.VideoHandler{
		Store:      s,
		Storage:    st,
		Signer:     sig,
		Queue:      q,
		Hooks:      h,
		Transcoder: t,
		StorageDir: storageDir,
	}

	srv := &Server{
		router:      chi.NewRouter(),
		handler:     vh,
		apiKey:      apiKey,
		playbackKey: playbackKey,
	}

	srv.routes()
	return srv
}

func (s *Server) routes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Use(customMiddleware.Logger)
	s.router.Use(middleware.Throttle(50)) 

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(customMiddleware.ScopedAuth(s.apiKey, s.playbackKey))

		r.Get("/videos/{id}/stream", s.handler.HandleSign)

		r.Group(func(r chi.Router) {
			r.Post("/videos", s.handler.HandleUpload)
			r.Get("/videos", s.handler.HandleList)
			r.Get("/videos/{id}", s.handler.HandleGet)
			r.Delete("/videos/{id}", s.handler.HandleDelete)
		})
	})

	s.router.Get("/health", s.handler.HandleHealth)
}

func (s *Server) Start(port string) error {
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("server starting", "port", port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
