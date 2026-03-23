package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"briers-api/internal/config"
	"briers-api/internal/database"
	"briers-api/internal/handlers"
	"briers-api/internal/middleware"
	"briers-api/internal/repository"
	"briers-api/internal/services"
	"briers-api/pkg/validator"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.SupabaseJWTSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET is required")
	}

	db := database.Connect(cfg.DatabaseURL)

	// Repositories
	profileRepo := repository.NewProfileRepository(db)
	productRepo := repository.NewProductRepository(db)
	sectionRepo := repository.NewSectionRepository(db)
	fabricRepo := repository.NewFabricPriceRepository(db)
	quoteRepo := repository.NewQuoteRepository(db)
	extraChargeRepo := repository.NewExtraChargeRepository(db)

	// Services
	pricingSvc := services.NewPricingService(fabricRepo)
	configuratorSvc := services.NewConfiguratorService(sectionRepo, pricingSvc)

	// Handlers
	profileHandler := handlers.NewProfileHandler(profileRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	sectionHandler := handlers.NewSectionHandler(sectionRepo)
	fabricHandler := handlers.NewFabricHandler(fabricRepo, pricingSvc)
	configuratorHandler := handlers.NewConfiguratorHandler(configuratorSvc)
	quoteHandler := handlers.NewQuoteHandler(quoteRepo, configuratorSvc)
	extraChargeHandler := handlers.NewExtraChargeHandler(extraChargeRepo)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"https://briers-frontend*.vercel.app",
			"http://localhost:5173",
			"http://localhost:3000",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// All API routes require JWT
	r.Group(func(r chi.Router) {
		r.Use(middleware.ValidateJWT(cfg.SupabaseJWTSecret))

		// ── Profile ───────────────────────────────────────────
		r.Get("/api/v1/me", profileHandler.Me)

		// ── Fabric Grades (public list for dropdowns) ─────────
		r.Get("/api/v1/grades", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(validator.ValidGradeSlice)
		})

		// ── Products ──────────────────────────────────────────
		r.Route("/api/v1/products", func(r chi.Router) {
			r.Get("/", productHandler.Search)
			r.Get("/{id}", productHandler.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(db, "admin", "manager"))
				r.Post("/", productHandler.Create)
				r.Put("/{id}", productHandler.Update)
				r.Delete("/{id}", productHandler.Delete)
			})
		})

		// ── Sections ──────────────────────────────────────────
		r.Route("/api/v1/products/{productId}/sections", func(r chi.Router) {
			r.Get("/", sectionHandler.List)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(db, "admin", "manager"))
				r.Post("/", sectionHandler.Create)
				r.Put("/{id}", sectionHandler.Update)
				r.Delete("/{id}", sectionHandler.Delete)
			})
		})

		// ── Fabric Prices ─────────────────────────────────────
		r.Route("/api/v1/products/{productId}/fabric-prices", func(r chi.Router) {
			r.Get("/", fabricHandler.List)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(db, "admin", "manager"))
				r.Put("/", fabricHandler.Upsert)
			})
		})

		// ── Extra Charges ─────────────────────────────────────
		r.Route("/api/v1/extra-charges", func(r chi.Router) {
			r.Get("/", extraChargeHandler.List)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole(db, "admin", "manager"))
				r.Post("/", extraChargeHandler.Create)
				r.Put("/{id}", extraChargeHandler.Update)
				r.Delete("/{id}", extraChargeHandler.Delete)
			})
		})

		// ── Configurator ──────────────────────────────────────
		r.Post("/api/v1/configure", configuratorHandler.Calculate)

		// ── Quotes ────────────────────────────────────────────
		r.Route("/api/v1/quotes", func(r chi.Router) {
			r.Get("/", quoteHandler.List)
			r.Post("/", quoteHandler.Create)
			r.Get("/{id}", quoteHandler.GetByID)
			r.Put("/{id}/status", quoteHandler.UpdateStatus)
		})

		// ── Users (admin only) ────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(db, "admin"))
			r.Get("/api/v1/users", profileHandler.ListAll)
			r.Put("/api/v1/users/{id}/role", profileHandler.UpdateRole)
		})
	})

	addr := ":" + cfg.Port
	log.Printf("Briers API running on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
