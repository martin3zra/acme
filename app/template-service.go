package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/martin3zra/acme/pkg/pdf"
)

// TemplateService orchestrates template operations
type TemplateService struct {
	repository *TemplateRepository
	renderer   pdf.Renderer
}

// NewTemplateService creates a new template service
func NewTemplateService(repo *TemplateRepository, renderer pdf.Renderer) *TemplateService {
	if renderer == nil {
		renderer = pdf.NewRenderer()
	}
	return &TemplateService{
		repository: repo,
		renderer:   renderer,
	}
}

// TemplateRepository represents the template storage interface
type TemplateRepository struct {
	server *Server
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(server *Server) *TemplateRepository {
	return &TemplateRepository{server: server}
}

// RenderTemplate renders a template with data
func (ts *TemplateService) RenderTemplate(ctx context.Context, templateID int, data map[string]any) ([]byte, error) {
	// Get published template version
	template, err := ts.repository.server.findTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}

	if template.CurrentVersionID == nil {
		return nil, fmt.Errorf("template has no current version")
	}

	// Get the published version
	version, err := ts.repository.server.getPublishedVersion(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("get published version: %w", err)
	}

	// Parse layout JSON
	var layout pdf.TemplateLayout
	if err := json.Unmarshal(version.LayoutJSON, &layout); err != nil {
		return nil, fmt.Errorf("parse layout JSON: %w", err)
	}

	// Render PDF
	pdfBytes, err := ts.renderer.Render(&layout, data)
	if err != nil {
		return nil, fmt.Errorf("render PDF: %w", err)
	}

	return pdfBytes, nil
}

// NormalizeLayoutJSON returns a canonical layout object JSON.
// It supports:
// 1) direct layout objects
// 2) stringified layout JSON
// 3) wrapper payloads with {"template": {...}, "data": {...}}
func NormalizeLayoutJSON(layoutJSON json.RawMessage) (json.RawMessage, error) {
	var rawData any
	if err := json.Unmarshal(layoutJSON, &rawData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if str, ok := rawData.(string); ok {
		return NormalizeLayoutJSON(json.RawMessage(str))
	}

	if obj, ok := rawData.(map[string]any); ok {
		if templateValue, hasTemplate := obj["template"]; hasTemplate {
			if templateValue == nil {
				return nil, fmt.Errorf("template is required")
			}

			if templateString, isString := templateValue.(string); isString {
				return NormalizeLayoutJSON(json.RawMessage(templateString))
			}

			rawData = templateValue
		}
	}

	normalized, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON structure: %w", err)
	}

	return json.RawMessage(normalized), nil
}

// UnmarshalLayoutJSON properly handles string, object, and wrapped JSON layouts
func UnmarshalLayoutJSON(layoutJSON json.RawMessage) (*pdf.TemplateLayout, error) {
	normalizedLayoutJSON, err := NormalizeLayoutJSON(layoutJSON)
	if err != nil {
		return nil, err
	}

	var layout pdf.TemplateLayout
	if err := json.Unmarshal(normalizedLayoutJSON, &layout); err != nil {
		return nil, fmt.Errorf("invalid layout structure: %w", err)
	}

	return &layout, nil
}

// ValidateTemplate validates a layout JSON
func (ts *TemplateService) ValidateTemplate(layoutJSON json.RawMessage) error {
	layout, err := UnmarshalLayoutJSON(layoutJSON)
	if err != nil {
		return err
	}

	engine := pdf.NewLayoutEngine()
	return engine.ValidateLayout(layout)
}

// PublishTemplate publishes a template version
func (ts *TemplateService) PublishTemplate(ctx context.Context, templateID int) (*TemplateVersion, error) {
	return ts.repository.server.publishTemplate(ctx, templateID)
}

// SaveDraft saves a draft version
func (ts *TemplateService) SaveDraft(ctx context.Context, templateID int, form *StoreTemplateForm) (*Template, error) {
	return ts.repository.server.updateTemplate(ctx, templateID, form)
}

// CreateTemplate creates a new template
func (ts *TemplateService) CreateTemplate(ctx context.Context, form *StoreTemplateForm) (*Template, error) {
	return ts.repository.server.storeTemplate(ctx, form)
}

// GetTemplate retrieves a template by ID
func (ts *TemplateService) GetTemplate(ctx context.Context, templateID int) (*Template, error) {
	return ts.repository.server.findTemplateByID(ctx, templateID)
}

// ListTemplates lists all templates for a company
func (ts *TemplateService) ListTemplates(ctx context.Context) ([]*Template, error) {
	return ts.repository.server.listTemplates(ctx)
}

// GetVersionHistory retrieves version history
func (ts *TemplateService) GetVersionHistory(ctx context.Context, templateID int) ([]*TemplateVersion, error) {
	return ts.repository.server.listTemplateVersions(ctx, templateID)
}

// DeleteTemplate deletes a template
func (ts *TemplateService) DeleteTemplate(ctx context.Context, templateID int) error {
	return ts.repository.server.deleteTemplate(ctx, templateID)
}

// DuplicateTemplate duplicates a template
func (ts *TemplateService) DuplicateTemplate(ctx context.Context, templateID int, newName string) (*Template, error) {
	return ts.repository.server.duplicateTemplate(ctx, templateID, newName)
}
