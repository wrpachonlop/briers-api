package services

import (
	"context"
	"fmt"

	"briers-api/internal/dto"
	"briers-api/internal/repository"
	"briers-api/pkg/validator"
)

// ─── Pricing Service ──────────────────────────────────────────────────────────

type PricingService struct {
	fabricRepo *repository.FabricPriceRepository
}

func NewPricingService(fr *repository.FabricPriceRepository) *PricingService {
	return &PricingService{fabricRepo: fr}
}

// CalculateFinalPrice: (supplierCost * 2) + 200
func (s *PricingService) CalculateFinalPrice(supplierCost float64) float64 {
	return (supplierCost * 2) + 200
}

func (s *PricingService) GetPriceForGrade(ctx context.Context, productID string, grade int) (*dto.PriceResult, error) {
	if err := validator.ValidateFabricGrade(grade); err != nil {
		return nil, err
	}
	fp, err := s.fabricRepo.FindByProductAndGrade(ctx, productID, grade)
	if err != nil {
		return nil, fmt.Errorf("no pricing found for grade %d on this product", grade)
	}
	return &dto.PriceResult{
		Grade:        grade,
		SupplierCost: fp.SupplierCost,
		FinalPrice:   s.CalculateFinalPrice(fp.SupplierCost),
	}, nil
}

// ─── Configurator Service ─────────────────────────────────────────────────────

type ConfiguratorService struct {
	sectionRepo *repository.SectionRepository
	pricingSvc  *PricingService
}

func NewConfiguratorService(sr *repository.SectionRepository, ps *PricingService) *ConfiguratorService {
	return &ConfiguratorService{sectionRepo: sr, pricingSvc: ps}
}

func (s *ConfiguratorService) Calculate(ctx context.Context, req dto.ConfiguratorRequest) (*dto.ConfigResult, error) {
	if err := validator.ValidateFabricGrade(req.FabricGrade); err != nil {
		return nil, err
	}

	var totalWidth, totalDepth, totalFabric float64

	for _, ps := range req.PlacedSections {
		section, err := s.sectionRepo.FindByID(ctx, ps.SectionID)
		if err != nil {
			return nil, fmt.Errorf("section %s not found", ps.SectionID)
		}
		w, d := effectiveDimensions(section.WidthCm, section.DepthCm, ps.Rotation)
		totalWidth += w * float64(ps.Quantity)
		if d > totalDepth {
			totalDepth = d
		}
		totalFabric += section.FabricYards * float64(ps.Quantity)
	}

	price, err := s.pricingSvc.GetPriceForGrade(ctx, req.ProductID, req.FabricGrade)
	if err != nil {
		return nil, err
	}

	var extraTotal float64
	for _, ec := range req.ExtraCharges {
		extraTotal += ec.Amount * float64(ec.Quantity)
	}

	return &dto.ConfigResult{
		TotalWidthCm:     totalWidth,
		TotalDepthCm:     totalDepth,
		TotalFabricYards: totalFabric,
		FabricGrade:      req.FabricGrade,
		SupplierCost:     price.SupplierCost,
		FinalPrice:       price.FinalPrice,
		ExtraCharges:     req.ExtraCharges,
		ExtraTotal:       extraTotal,
		GrandTotal:       price.FinalPrice + extraTotal,
	}, nil
}

// effectiveDimensions swaps width/depth when rotated 90° or 270°.
func effectiveDimensions(width, depth float64, rotation int) (float64, float64) {
	if rotation == 90 || rotation == 270 {
		return depth, width
	}
	return width, depth
}
