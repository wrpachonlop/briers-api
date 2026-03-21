package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"briers-api/internal/models"
)

// ─── Profile Repository ───────────────────────────────────────────────────────

type ProfileRepository struct{ db *gorm.DB }

func NewProfileRepository(db *gorm.DB) *ProfileRepository { return &ProfileRepository{db} }

func (r *ProfileRepository) FindByID(ctx context.Context, id string) (*models.Profile, error) {
	var p models.Profile
	return &p, r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
}

func (r *ProfileRepository) FindAll(ctx context.Context) ([]models.Profile, error) {
	var profiles []models.Profile
	return profiles, r.db.WithContext(ctx).Order("full_name").Find(&profiles).Error
}

func (r *ProfileRepository) UpdateRole(ctx context.Context, id, role string) error {
	return r.db.WithContext(ctx).Model(&models.Profile{}).
		Where("id = ?", id).Update("role", role).Error
}

// ─── Product Repository ───────────────────────────────────────────────────────

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db} }

func (r *ProductRepository) Search(ctx context.Context, q, productType string) ([]models.Product, error) {
	var products []models.Product
	db := r.db.WithContext(ctx).
		Preload("FabricPrices").
		Where("is_active = true")

	if q != "" {
		pattern := "%" + strings.ToLower(q) + "%"
		db = db.Where("lower(provider_name) LIKE ? OR lower(store_name) LIKE ?", pattern, pattern)
	}
	if productType != "" {
		db = db.Where("product_type = ?", productType)
	}
	return products, db.Order("store_name ASC").Find(&products).Error
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	err := r.db.WithContext(ctx).
		Preload("Sections", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("FabricPrices", func(db *gorm.DB) *gorm.DB { return db.Order("grade ASC") }).
		First(&p, "id = ?", id).Error
	return &p, err
}

func (r *ProductRepository) Create(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProductRepository) Update(ctx context.Context, p *models.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ?", id).Update("is_active", false).Error
}

// ─── Section Repository ───────────────────────────────────────────────────────

type SectionRepository struct{ db *gorm.DB }

func NewSectionRepository(db *gorm.DB) *SectionRepository { return &SectionRepository{db} }

func (r *SectionRepository) FindByProductID(ctx context.Context, productID string) ([]models.Section, error) {
	var sections []models.Section
	return sections, r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("sort_order ASC").
		Find(&sections).Error
}

func (r *SectionRepository) FindByID(ctx context.Context, id string) (*models.Section, error) {
	var s models.Section
	return &s, r.db.WithContext(ctx).First(&s, "id = ?", id).Error
}

func (r *SectionRepository) Create(ctx context.Context, s *models.Section) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SectionRepository) Update(ctx context.Context, s *models.Section) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *SectionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Section{}, "id = ?", id).Error
}

// ─── FabricPrice Repository ───────────────────────────────────────────────────

type FabricPriceRepository struct{ db *gorm.DB }

func NewFabricPriceRepository(db *gorm.DB) *FabricPriceRepository { return &FabricPriceRepository{db} }

func (r *FabricPriceRepository) FindByProductID(ctx context.Context, productID string) ([]models.FabricPrice, error) {
	var fps []models.FabricPrice
	return fps, r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("grade ASC").
		Find(&fps).Error
}

func (r *FabricPriceRepository) FindByProductAndGrade(ctx context.Context, productID string, grade int) (*models.FabricPrice, error) {
	var fp models.FabricPrice
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND grade = ?", productID, grade).
		First(&fp).Error
	return &fp, err
}

func (r *FabricPriceRepository) Upsert(ctx context.Context, productID string, grade int, cost float64) error {
	pid, err := uuid.Parse(productID)
	if err != nil {
		return fmt.Errorf("invalid product id")
	}
	fp := models.FabricPrice{
		ProductID:    pid,
		Grade:        grade,
		SupplierCost: cost,
	}
	return r.db.WithContext(ctx).
		Where(models.FabricPrice{ProductID: pid, Grade: grade}).
		Assign(models.FabricPrice{SupplierCost: cost}).
		FirstOrCreate(&fp).Error
}

// ─── ExtraCharge Repository ───────────────────────────────────────────────────

type ExtraChargeRepository struct{ db *gorm.DB }

func NewExtraChargeRepository(db *gorm.DB) *ExtraChargeRepository { return &ExtraChargeRepository{db} }

func (r *ExtraChargeRepository) FindAll(ctx context.Context) ([]models.ExtraCharge, error) {
	var charges []models.ExtraCharge
	return charges, r.db.WithContext(ctx).Where("is_active = true").Order("name").Find(&charges).Error
}

func (r *ExtraChargeRepository) Create(ctx context.Context, ec *models.ExtraCharge) error {
	return r.db.WithContext(ctx).Create(ec).Error
}

func (r *ExtraChargeRepository) Update(ctx context.Context, ec *models.ExtraCharge) error {
	return r.db.WithContext(ctx).Save(ec).Error
}

func (r *ExtraChargeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.ExtraCharge{}).
		Where("id = ?", id).Update("is_active", false).Error
}

// ─── Quote Repository ─────────────────────────────────────────────────────────

type QuoteRepository struct{ db *gorm.DB }

func NewQuoteRepository(db *gorm.DB) *QuoteRepository { return &QuoteRepository{db} }

func (r *QuoteRepository) FindAll(ctx context.Context, userID string, isAdmin bool) ([]models.Quote, error) {
	var quotes []models.Quote
	db := r.db.WithContext(ctx).Preload("Product")
	if !isAdmin {
		db = db.Where("created_by = ?", userID)
	}
	return quotes, db.Order("created_at DESC").Find(&quotes).Error
}

func (r *QuoteRepository) FindByID(ctx context.Context, id string) (*models.Quote, error) {
	var q models.Quote
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("QuoteSections.Section").
		Preload("ExtraCharges").
		First(&q, "id = ?", id).Error
	return &q, err
}

func (r *QuoteRepository) Create(ctx context.Context, q *models.Quote) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *QuoteRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).Model(&models.Quote{}).
		Where("id = ?", id).Update("status", status).Error
}
