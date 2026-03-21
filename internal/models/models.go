package models

import (
	"time"

	"github.com/google/uuid"
)

// ─── Profile ────────────────────────────────────────────────────────────────

type Profile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"           json:"id"`
	FullName  string    `gorm:"not null"                       json:"full_name"`
	Role      string    `gorm:"not null"                       json:"role"` // admin | manager | seller
	CreatedAt time.Time `gorm:"autoCreateTime"                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                 json:"updated_at"`
}

func (Profile) TableName() string { return "profiles" }

// ─── Product ─────────────────────────────────────────────────────────────────

type ProductType string

const (
	ProductTypeModular ProductType = "modular"
	ProductTypeFixed   ProductType = "fixed"
)

type Product struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProviderName string      `gorm:"not null"                                        json:"provider_name"`
	StoreName    string      `gorm:"not null"                                        json:"store_name"`
	ProductType  ProductType `gorm:"type:varchar(20);not null"                       json:"product_type"`
	Description  string      `gorm:"type:text"                                       json:"description"`
	ImageURL     string      `gorm:"column:image_url"                                json:"image_url"`
	IsActive     bool        `gorm:"default:true"                                    json:"is_active"`
	CreatedBy    *uuid.UUID  `gorm:"type:uuid"                                       json:"created_by"`
	CreatedAt    time.Time   `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt    time.Time   `gorm:"autoUpdateTime"                                  json:"updated_at"`

	Sections     []Section     `gorm:"foreignKey:ProductID" json:"sections,omitempty"`
	FabricPrices []FabricPrice `gorm:"foreignKey:ProductID" json:"fabric_prices,omitempty"`
}

func (Product) TableName() string { return "products" }

// ─── Section ─────────────────────────────────────────────────────────────────

type Section struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null"                              json:"product_id"`
	Name        string    `gorm:"not null"                                        json:"name"`
	WidthCm     float64   `gorm:"column:width_cm;not null"                        json:"width_cm"`
	HeightCm    float64   `gorm:"column:height_cm;not null"                       json:"height_cm"`
	DepthCm     float64   `gorm:"column:depth_cm;not null"                        json:"depth_cm"`
	FabricYards float64   `gorm:"column:fabric_yards;not null"                    json:"fabric_yards"`
	ImageURL    string    `gorm:"column:image_url;not null"                       json:"image_url"`
	SortOrder   int       `gorm:"default:0"                                       json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
}

func (Section) TableName() string { return "sections" }

// ─── FabricPrice ─────────────────────────────────────────────────────────────

type FabricPrice struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProductID    uuid.UUID `gorm:"type:uuid;not null"                              json:"product_id"`
	Grade        int       `gorm:"not null"                                        json:"grade"`
	SupplierCost float64   `gorm:"column:supplier_cost;not null"                   json:"supplier_cost"`
	CreatedAt    time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (FabricPrice) TableName() string { return "fabric_prices" }

// ─── ExtraCharge ─────────────────────────────────────────────────────────────

type ExtraCharge struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"not null"                                        json:"name"`
	Amount      float64   `gorm:"not null"                                        json:"amount"`
	Description string    `gorm:"type:text"                                       json:"description"`
	IsActive    bool      `gorm:"default:true"                                    json:"is_active"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
}

func (ExtraCharge) TableName() string { return "extra_charges" }

// ─── Quote ───────────────────────────────────────────────────────────────────

type Quote struct {
	ID                uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProductID         uuid.UUID          `gorm:"type:uuid;not null"                              json:"product_id"`
	CreatedBy         uuid.UUID          `gorm:"type:uuid;not null"                              json:"created_by"`
	CustomerName      string             `gorm:"type:text"                                       json:"customer_name"`
	FabricGrade       int                `gorm:"not null"                                        json:"fabric_grade"`
	SupplierCost      float64            `gorm:"column:supplier_cost;not null"                   json:"supplier_cost"`
	FinalPrice        float64            `gorm:"column:final_price;not null"                     json:"final_price"`
	TotalWidthCm      float64            `gorm:"column:total_width_cm;not null"                  json:"total_width_cm"`
	TotalDepthCm      float64            `gorm:"column:total_depth_cm;not null"                  json:"total_depth_cm"`
	TotalFabricYards  float64            `gorm:"column:total_fabric_yards;not null"              json:"total_fabric_yards"`
	Notes             string             `gorm:"type:text"                                       json:"notes"`
	Status            string             `gorm:"default:draft"                                   json:"status"`
	CreatedAt         time.Time          `gorm:"autoCreateTime"                                  json:"created_at"`

	Product      Product            `gorm:"foreignKey:ProductID"  json:"product,omitempty"`
	QuoteSections []QuoteSection    `gorm:"foreignKey:QuoteID"    json:"sections,omitempty"`
	ExtraCharges  []QuoteExtraCharge `gorm:"foreignKey:QuoteID"   json:"extra_charges,omitempty"`
}

func (Quote) TableName() string { return "quotes" }

type QuoteSection struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	QuoteID   uuid.UUID `gorm:"type:uuid;not null"                              json:"quote_id"`
	SectionID uuid.UUID `gorm:"type:uuid;not null"                              json:"section_id"`
	Quantity  int       `gorm:"default:1"                                       json:"quantity"`
	Rotation  int       `gorm:"default:0"                                       json:"rotation"`
	PositionX int       `gorm:"column:position_x"                               json:"position_x"`
	PositionY int       `gorm:"column:position_y"                               json:"position_y"`

	Section Section `gorm:"foreignKey:SectionID" json:"section,omitempty"`
}

func (QuoteSection) TableName() string { return "quote_sections" }

type QuoteExtraCharge struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	QuoteID       uuid.UUID  `gorm:"type:uuid;not null"                              json:"quote_id"`
	ExtraChargeID *uuid.UUID `gorm:"type:uuid"                                       json:"extra_charge_id"`
	Name          string     `gorm:"not null"                                        json:"name"`
	Amount        float64    `gorm:"not null"                                        json:"amount"`
	Quantity      int        `gorm:"default:1"                                       json:"quantity"`
}

func (QuoteExtraCharge) TableName() string { return "quote_extra_charges" }
