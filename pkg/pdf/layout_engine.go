package pdf

import (
	"encoding/json"
	"fmt"
)

// Element represents a parsed PDF element
type Element struct {
	Type string
	Data interface{}
}

// LayoutEngine parses JSON layout definitions into program structures
type LayoutEngine struct{}

// NewLayoutEngine creates a new layout engine
func NewLayoutEngine() *LayoutEngine {
	return &LayoutEngine{}
}

// ParseElement parses a JSON element definition into an Element
func (le *LayoutEngine) ParseElement(elemJSON json.RawMessage) (*Element, error) {
	var rawElem map[string]interface{}
	if err := json.Unmarshal(elemJSON, &rawElem); err != nil {
		return nil, fmt.Errorf("parse element JSON: %w", err)
	}

	elemType, ok := rawElem["type"].(string)
	if !ok {
		return nil, fmt.Errorf("element missing type field")
	}

	return &Element{
		Type: elemType,
		Data: rawElem,
	}, nil
}

// ParseLayout parses a JSON layout definition
func (le *LayoutEngine) ParseLayout(layoutJSON []byte) (*TemplateLayout, error) {
	var layout TemplateLayout
	if err := json.Unmarshal(layoutJSON, &layout); err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}

	// Validate required fields
	if layout.PageSize == "" {
		layout.PageSize = "A4"
	}
	if layout.PageFormat == "" {
		layout.PageFormat = "P"
	}

	if len(layout.Elements) == 0 {
		return nil, fmt.Errorf("layout has no elements")
	}

	return &layout, nil
}

// ValidateLayout validates a template layout against a schema
func (le *LayoutEngine) ValidateLayout(layout *TemplateLayout) error {
	if layout == nil {
		return fmt.Errorf("layout is nil")
	}

	if layout.PageSize == "" {
		return fmt.Errorf("page_size is required")
	}

	validPageSizes := map[string]bool{
		"A4": true, "A3": true, "A5": true,
		"Letter": true, "Tabloid": true, "Ledger": true,
	}
	if !validPageSizes[layout.PageSize] {
		return fmt.Errorf("invalid page_size: %s", layout.PageSize)
	}

	if layout.PageFormat != "P" && layout.PageFormat != "L" {
		return fmt.Errorf("page_format must be P (portrait) or L (landscape)")
	}

	if len(layout.Elements) == 0 {
		return fmt.Errorf("layout must have at least one element")
	}

	// Validate margins
	if layout.Margins.Top < 0 {
		return fmt.Errorf("top margin cannot be negative")
	}
	if layout.Margins.Left < 0 {
		return fmt.Errorf("left margin cannot be negative")
	}
	if layout.Margins.Right < 0 {
		return fmt.Errorf("right margin cannot be negative")
	}
	if layout.Margins.Bottom < 0 {
		return fmt.Errorf("bottom margin cannot be negative")
	}

	return nil
}

// MeasureElement returns the height of an element for pagination
func (le *LayoutEngine) MeasureElement(elem *Element) float64 {
	// Returns approximate height in mm
	switch elem.Type {
	case "text":
		return 10 // Single line text
	case "image":
		if data, ok := elem.Data.(map[string]any); ok {
			return getFloat(data, "height", 50)
		}
	case "table":
		if data, ok := elem.Data.(map[string]any); ok {
			if rows, ok := data["rows"].([]any); ok {
				rowHeight := getFloat(data, "row_height", 8)
				// Header + rows
				return rowHeight * (1 + float64(len(rows)))
			}
		}
	case "divider":
		return 5
	case "qr":
		if data, ok := elem.Data.(map[string]any); ok {
			return getFloat(data, "size", 30)
		}
	case "grid":
		// Grid height is sum of children
		return 20
	}
	return 10
}

// GetComponentType returns the type of a component
func (le *LayoutEngine) GetComponentType(elem *Element) string {
	return elem.Type
}
