package components

import (
	"errors"

	"codeberg.org/go-pdf/fpdf"
)

// DividerProps defines the properties for rendering a divider line on a PDF.
type DividerProps struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	Thickness float64
	Color     RGBColor
}

// DividerRenderer handles divider line rendering on PDF documents.
type DividerRenderer struct {
	pdf *fpdf.Fpdf
}

// NewDividerRenderer creates a new DividerRenderer instance.
func NewDividerRenderer(pdf *fpdf.Fpdf) *DividerRenderer {
	return &DividerRenderer{
		pdf: pdf,
	}
}

// Render draws a horizontal divider line on the PDF.
func (dr *DividerRenderer) Render(props DividerProps) error {
	if dr.pdf == nil {
		return errors.New("errNilPDF")
	}

	// Set line color
	dr.pdf.SetDrawColor(props.Color.R, props.Color.G, props.Color.B)

	// Set line width/thickness
	dr.pdf.SetLineWidth(props.Thickness)

	// Calculate end position
	endX := props.X + props.Width
	endY := props.Y + props.Height

	// Draw horizontal line
	dr.pdf.Line(props.X, props.Y, endX, endY)

	return nil
}

// GetHeight returns the height of the divider.
func (dr *DividerRenderer) GetHeight(props DividerProps) float64 {
	return props.Height + props.Thickness
}
