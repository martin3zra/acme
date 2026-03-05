package components

import (
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
		return ErrNilPDF
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
		return ErrNilPDF
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

import (
	"codeberg.org/go-pdf/fpdf"
)




































}	return 0.5func (dr *DividerRenderer) GetHeight() float64 {// GetHeight returns the height of the divider}	pdf.SetLineWidth(0.3)	pdf.SetDrawColor(0, 0, 0)	// Reset	pdf.Line(props.X, props.Y, props.X+props.Width, props.Y)	// Draw horizontal line	pdf.SetLineWidth(props.Thickness)	pdf.SetDrawColor(int(props.Color.R), int(props.Color.G), int(props.Color.B))func (dr *DividerRenderer) Render(pdf *fpdf.Fpdf, props DividerProps) {// Render renders a divider element on the PDF}	return &DividerRenderer{}func NewDividerRenderer() *DividerRenderer {// NewDividerRenderer creates a new divider renderer}	Color     RGBColor	Thickness float64	Height    float64	Width     float64	Y         float64	X         float64type DividerProps struct {// DividerProps represents divider element propertiestype DividerRenderer struct{}// DividerRenderer handles rendering horizontal divider/line elements