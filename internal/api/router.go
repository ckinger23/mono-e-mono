package api

import (
	"net/http"

	"github.com/ckinger23/mono-e-mono/internal/auth"
	"github.com/ckinger23/mono-e-mono/internal/db"
	"github.com/ckinger23/mono-e-mono/internal/game"
	"github.com/ckinger23/mono-e-mono/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds the API configuration
type Config struct {
	FrontendURL string
	JWTSecret   string
	OAuth       *auth.OAuthConfig
}

// Server holds all the dependencies for the API
type Server struct {
	router       *chi.Mux
	pool         *pgxpool.Pool
	queries      *db.Queries
	jwtManager   *auth.JWTManager
	oauthManager *auth.OAuthManager
	config       *Config
	hub          *ws.Hub
	draftManager *game.WeeklyDraftManager
	playerDB     *game.PlayerDB
}

// NewServer creates a new API server
func NewServer(pool *pgxpool.Pool, config *Config) *Server {
	// Load player database
	playerDB, err := game.LoadPlayerDB()
	if err != nil {
		// Log error but continue - drafts will fail but other endpoints work
		playerDB = nil
	}

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create draft manager
	var draftManager *game.WeeklyDraftManager
	if playerDB != nil {
		draftManager = game.NewWeeklyDraftManager(playerDB, pool)
	}

	s := &Server{
		router:       chi.NewRouter(),
		pool:         pool,
		queries:      db.NewQueries(pool),
		jwtManager:   auth.NewJWTManager(auth.DefaultJWTConfig(config.JWTSecret)),
		oauthManager: auth.NewOAuthManager(config.OAuth),
		config:       config,
		hub:          hub,
		draftManager: draftManager,
		playerDB:     playerDB,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Basic middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.config.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// WebSocket route for drafts (token passed as query param)
	s.router.Get("/ws/draft/{draftID}", s.handleDraftWebSocket)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/refresh", s.handleRefresh)

			// OAuth routes
			r.Get("/google", s.handleOAuthGoogle)
			r.Get("/google/callback", s.handleOAuthGoogleCallback)
			r.Get("/github", s.handleOAuthGitHub)
			r.Get("/github/callback", s.handleOAuthGitHubCallback)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(s.AuthMiddleware)

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", s.handleGetCurrentUser)
				r.Put("/me", s.handleUpdateCurrentUser)
			})

			// Draft routes (REST)
			r.Route("/drafts", func(r chi.Router) {
				r.Get("/{draftID}/roster", s.handleGetDraftRoster)
			})

			// League routes
			r.Route("/leagues", func(r chi.Router) {
				r.Get("/", s.handleGetLeagues)
				r.Post("/", s.handleCreateLeague)
				r.Post("/join", s.handleJoinLeague)

				r.Route("/{leagueID}", func(r chi.Router) {
					r.Get("/", s.handleGetLeague)
					r.Put("/", s.handleUpdateLeague)
					r.Delete("/", s.handleDeleteLeague)
					r.Get("/members", s.handleGetLeagueMembers)
					r.Get("/standings", s.handleGetLeagueStandings)

					// Season routes
					r.Route("/seasons", func(r chi.Router) {
						r.Get("/", s.handleGetSeasons)
						r.Post("/", s.handleCreateSeason)
						r.Get("/active", s.handleGetActiveSeason)

						r.Route("/{seasonID}", func(r chi.Router) {
							r.Get("/", s.handleGetSeason)
							r.Get("/standings", s.handleGetSeasonStandings)

							// Weekly results
							r.Get("/week/{week}/results", s.handleGetWeeklyResults)

							// Draft routes
							r.Route("/week/{week}/draft", func(r chi.Router) {
								r.Get("/", s.handleGetWeeklyDraft)
								r.Post("/", s.handleStartWeeklyDraft)
							})
						})
					})
				})
			})

			// NFL data routes (authenticated, read-only)
			r.Route("/nfl", func(r chi.Router) {
				r.Get("/state", s.handleGetNFLState)
				r.Get("/teams", s.handleGetNFLTeams)
				r.Get("/players", s.handleGetNFLPlayers)
				r.Get("/leaders", s.handleGetWeeklyLeaders)
			})

			// Admin routes (should be protected by admin middleware in production)
			r.Route("/admin", func(r chi.Router) {
				// Player/Stats sync
				r.Post("/sync-players", s.handleSyncPlayers)
				r.Post("/sync-stats", s.handleSyncWeeklyStats)

				// Scoring
				r.Post("/seasons/{seasonID}/calculate-scores", s.handleCalculateScores)
				r.Post("/seasons/{seasonID}/recalculate-all", s.handleRecalculateAllWeeks)
			})
		})
	})
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}
