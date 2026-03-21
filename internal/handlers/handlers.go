package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"briers-api/internal/dto"
	"briers-api/internal/middleware"
	"briers-api/internal/models"
	"briers-api/internal/repository"
	"briers-api/internal/services"
	"briers-api/pkg/validator"
)

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, dto.ErrorResponse{Error: msg})
}

// ─── Product Handler ──────────────────────────────────────────────────────────

type ProductHandler struct {
	repo *repository.ProductRepository
}

func NewProductHandler(r *repository.ProductRepository) *ProductHandler { return &ProductHandler{r} }

func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	pt := r.URL.Query().Get("type")
	products, err := h.repo.Search(r.Context(), q, pt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search failed")
		return
	}
	respond(w, http.StatusOK, products)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	product, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "product not found")
		return
	}
	respond(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ProviderName == "" || req.StoreName == "" {
		respondError(w, http.StatusBadRequest, "provider_name and store_name are required")
		return
	}

	claims, _ := middleware.GetClaims(r)
	createdBy, _ := uuid.Parse(claims.Sub)

	product := &models.Product{
		ProviderName: req.ProviderName,
		StoreName:    req.StoreName,
		ProductType:  models.ProductType(req.ProductType),
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		CreatedBy:    &createdBy,
	}

	if err := h.repo.Create(r.Context(), product); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create product")
		return
	}
	respond(w, http.StatusCreated, product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	product, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "product not found")
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ProviderName != "" {
		product.ProviderName = req.ProviderName
	}
	if req.StoreName != "" {
		product.StoreName = req.StoreName
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := h.repo.Update(r.Context(), product); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update product")
		return
	}
	respond(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// ─── Section Handler ──────────────────────────────────────────────────────────

type SectionHandler struct {
	repo *repository.SectionRepository
}

func NewSectionHandler(r *repository.SectionRepository) *SectionHandler { return &SectionHandler{r} }

func (h *SectionHandler) List(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	sections, err := h.repo.FindByProductID(r.Context(), productID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch sections")
		return
	}
	respond(w, http.StatusOK, sections)
}

func (h *SectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	pid, err := uuid.Parse(productID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req dto.CreateSectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	section := &models.Section{
		ProductID:   pid,
		Name:        req.Name,
		WidthCm:     req.WidthCm,
		HeightCm:    req.HeightCm,
		DepthCm:     req.DepthCm,
		FabricYards: req.FabricYards,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
	}

	if err := h.repo.Create(r.Context(), section); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create section")
		return
	}
	respond(w, http.StatusCreated, section)
}

func (h *SectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	section, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "section not found")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(section); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.Update(r.Context(), section); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update section")
		return
	}
	respond(w, http.StatusOK, section)
}

func (h *SectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete section")
		return
	}
	respond(w, http.StatusNoContent, nil)
}

// ─── Fabric Handler ───────────────────────────────────────────────────────────

type FabricHandler struct {
	repo       *repository.FabricPriceRepository
	pricingSvc *services.PricingService
}

func NewFabricHandler(r *repository.FabricPriceRepository, ps *services.PricingService) *FabricHandler {
	return &FabricHandler{r, ps}
}

func (h *FabricHandler) List(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")
	prices, err := h.repo.FindByProductID(r.Context(), productID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch fabric prices")
		return
	}

	// Enrich with final price
	type enriched struct {
		models.FabricPrice
		FinalPrice float64 `json:"final_price"`
	}
	result := make([]enriched, len(prices))
	for i, p := range prices {
		result[i] = enriched{p, h.pricingSvc.CalculateFinalPrice(p.SupplierCost)}
	}
	respond(w, http.StatusOK, result)
}

func (h *FabricHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")

	var req dto.UpsertFabricPricesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for _, entry := range req.Prices {
		if err := validator.ValidateFabricGrade(entry.Grade); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := h.repo.Upsert(r.Context(), productID, entry.Grade, entry.SupplierCost); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save fabric prices")
			return
		}
	}
	respond(w, http.StatusOK, map[string]string{"status": "saved"})
}

// ─── Configurator Handler ─────────────────────────────────────────────────────

type ConfiguratorHandler struct {
	svc *services.ConfiguratorService
}

func NewConfiguratorHandler(s *services.ConfiguratorService) *ConfiguratorHandler {
	return &ConfiguratorHandler{s}
}

func (h *ConfiguratorHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfiguratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Calculate(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusOK, result)
}

// ─── Quote Handler ────────────────────────────────────────────────────────────

type QuoteHandler struct {
	repo           *repository.QuoteRepository
	configuratorSvc *services.ConfiguratorService
}

func NewQuoteHandler(r *repository.QuoteRepository, cs *services.ConfiguratorService) *QuoteHandler {
	return &QuoteHandler{r, cs}
}

func (h *QuoteHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	// TODO: check role for admin visibility
	quotes, err := h.repo.FindAll(r.Context(), claims.Sub, false)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch quotes")
		return
	}
	respond(w, http.StatusOK, quotes)
}

func (h *QuoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	quote, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "quote not found")
		return
	}
	respond(w, http.StatusOK, quote)
}

func (h *QuoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	configReq := dto.ConfiguratorRequest{
		ProductID:      req.ProductID,
		FabricGrade:    req.FabricGrade,
		PlacedSections: req.PlacedSections,
		ExtraCharges:   req.ExtraCharges,
	}
	result, err := h.configuratorSvc.Calculate(r.Context(), configReq)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	claims, _ := middleware.GetClaims(r)
	createdBy, _ := uuid.Parse(claims.Sub)
	productID, _ := uuid.Parse(req.ProductID)

	quote := &models.Quote{
		ProductID:        productID,
		CreatedBy:        createdBy,
		CustomerName:     req.CustomerName,
		FabricGrade:      req.FabricGrade,
		SupplierCost:     result.SupplierCost,
		FinalPrice:       result.GrandTotal,
		TotalWidthCm:     result.TotalWidthCm,
		TotalDepthCm:     result.TotalDepthCm,
		TotalFabricYards: result.TotalFabricYards,
		Notes:            req.Notes,
		Status:           "draft",
	}

	for _, ps := range req.PlacedSections {
		sid, _ := uuid.Parse(ps.SectionID)
		quote.QuoteSections = append(quote.QuoteSections, models.QuoteSection{
			SectionID: sid,
			Quantity:  ps.Quantity,
			Rotation:  ps.Rotation,
			PositionX: ps.PositionX,
			PositionY: ps.PositionY,
		})
	}

	for _, ec := range req.ExtraCharges {
		qec := models.QuoteExtraCharge{
			Name:     ec.Name,
			Amount:   ec.Amount,
			Quantity: ec.Quantity,
		}
		if ec.ID != "" {
			parsed, err := uuid.Parse(ec.ID)
			if err == nil {
				qec.ExtraChargeID = &parsed
			}
		}
		quote.ExtraCharges = append(quote.ExtraCharges, qec)
	}

	if err := h.repo.Create(r.Context(), quote); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create quote")
		return
	}
	respond(w, http.StatusCreated, quote)
}

func (h *QuoteHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateQuoteStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update status")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": req.Status})
}

// ─── Extra Charge Handler ─────────────────────────────────────────────────────

type ExtraChargeHandler struct {
	repo *repository.ExtraChargeRepository
}

func NewExtraChargeHandler(r *repository.ExtraChargeRepository) *ExtraChargeHandler {
	return &ExtraChargeHandler{r}
}

func (h *ExtraChargeHandler) List(w http.ResponseWriter, r *http.Request) {
	charges, err := h.repo.FindAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch charges")
		return
	}
	respond(w, http.StatusOK, charges)
}

func (h *ExtraChargeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ec models.ExtraCharge
	if err := json.NewDecoder(r.Body).Decode(&ec); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.repo.Create(r.Context(), &ec); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create charge")
		return
	}
	respond(w, http.StatusCreated, ec)
}

func (h *ExtraChargeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var ec models.ExtraCharge
	if err := json.NewDecoder(r.Body).Decode(&ec); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ec.ID = parsed
	if err := h.repo.Update(r.Context(), &ec); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update charge")
		return
	}
	respond(w, http.StatusOK, ec)
}

func (h *ExtraChargeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete charge")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "deactivated"})
}

// ─── Profile Handler ──────────────────────────────────────────────────────────

type ProfileHandler struct {
	repo *repository.ProfileRepository
}

func NewProfileHandler(r *repository.ProfileRepository) *ProfileHandler { return &ProfileHandler{r} }

func (h *ProfileHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	profile, err := h.repo.FindByID(r.Context(), claims.Sub)
	if err != nil {
		respondError(w, http.StatusNotFound, "profile not found")
		return
	}
	respond(w, http.StatusOK, profile)
}

func (h *ProfileHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.repo.FindAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch profiles")
		return
	}
	respond(w, http.StatusOK, profiles)
}

func (h *ProfileHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	validRoles := map[string]bool{"admin": true, "manager": true, "seller": true}
	if !validRoles[body.Role] {
		respondError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if err := h.repo.UpdateRole(r.Context(), id, body.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update role")
		return
	}
	respond(w, http.StatusOK, map[string]string{"role": body.Role})
}

// Ensure time import is used (referenced in Quote model)
var _ = time.Now
