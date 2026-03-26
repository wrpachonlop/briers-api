package dto

import "github.com/google/uuid"

// ─── Product DTOs ─────────────────────────────────────────────────────────────

type CreateProductRequest struct {
	ProviderName string `json:"provider_name" validate:"required"`
	StoreName    string `json:"store_name"    validate:"required"`
	ProductType  string `json:"product_type"  validate:"required,oneof=modular fixed"`
	Description  string `json:"description"`
	ImageURL     string `json:"image_url"`
}

type UpdateProductRequest struct {
	ProviderName string `json:"provider_name"`
	StoreName    string `json:"store_name"`
	Description  string `json:"description"`
	ImageURL     string `json:"image_url"`
	IsActive     *bool  `json:"is_active"`
}

// ─── Section DTOs ─────────────────────────────────────────────────────────────

type SectionPriceRequest struct {
	Grade int     `json:"grade"`
	Price float64 `json:"price"`
}
type CreateSectionRequest struct {
	Name        string                `json:"name"         validate:"required"`
	WidthCm     float64               `json:"width_cm"     validate:"required,gt=0"`
	HeightCm    float64               `json:"height_cm"    validate:"required,gt=0"`
	DepthCm     float64               `json:"depth_cm"     validate:"required,gt=0"`
	FabricYards float64               `json:"fabric_yards" validate:"required,gt=0"`
	ImageURL    string                `json:"image_url"    validate:"required"`
	SortOrder   int                   `json:"sort_order"`
	Prices      []SectionPriceRequest `json:"prices"`
}

// ─── Fabric Price DTOs ────────────────────────────────────────────────────────

type UpsertFabricPricesRequest struct {
	Prices []FabricPriceEntry `json:"prices" validate:"required,min=8,max=8"`
}

type FabricPriceEntry struct {
	Grade        int     `json:"grade"          validate:"required"`
	SupplierCost float64 `json:"supplier_cost"  validate:"required,gt=0"`
}

// ─── Configurator DTOs ────────────────────────────────────────────────────────

type ConfiguratorRequest struct {
	ProductID      string            `json:"product_id"      validate:"required"`
	FabricGrade    int               `json:"fabric_grade"    validate:"required"`
	PlacedSections []PlacedSection   `json:"placed_sections" validate:"required,min=1"`
	ExtraCharges   []ExtraChargeLine `json:"extra_charges"`
}

type PlacedSection struct {
	SectionID string `json:"section_id" validate:"required"`
	Quantity  int    `json:"quantity"   validate:"required,min=1"`
	Rotation  int    `json:"rotation"   validate:"oneof=0 90 180 270"`
	PositionX int    `json:"position_x"`
	PositionY int    `json:"position_y"`
}

type ExtraChargeLine struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"     validate:"required"`
	Amount   float64 `json:"amount"   validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
}

// ─── Quote DTOs ───────────────────────────────────────────────────────────────

type CreateQuoteRequest struct {
	ProductID      string            `json:"product_id"      validate:"required"`
	CustomerName   string            `json:"customer_name"`
	FabricGrade    int               `json:"fabric_grade"    validate:"required"`
	PlacedSections []PlacedSection   `json:"placed_sections" validate:"required,min=1"`
	ExtraCharges   []ExtraChargeLine `json:"extra_charges"`
	Notes          string            `json:"notes"`
}

type UpdateQuoteStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft sent accepted declined"`
}

// ─── Responses ───────────────────────────────────────────────────────────────

type ConfigResult struct {
	TotalWidthCm     float64           `json:"total_width_cm"`
	TotalDepthCm     float64           `json:"total_depth_cm"`
	TotalFabricYards float64           `json:"total_fabric_yards"`
	FabricGrade      int               `json:"fabric_grade"`
	SupplierCost     float64           `json:"supplier_cost"`
	FinalPrice       float64           `json:"final_price"`
	ExtraCharges     []ExtraChargeLine `json:"extra_charges"`
	ExtraTotal       float64           `json:"extra_total"`
	GrandTotal       float64           `json:"grand_total"`
}

type PriceResult struct {
	Grade        int     `json:"grade"`
	SupplierCost float64 `json:"supplier_cost"`
	FinalPrice   float64 `json:"final_price"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type IDResponse struct {
	ID uuid.UUID `json:"id"`
}
