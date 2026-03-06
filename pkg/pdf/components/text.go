package components

import (
	"errors"

	"codeberg.org/go-pdf/fpdf"
)

// RGBColor represents an RGB color value.
type RGBColor struct {
	R int
	G int
	B int
}

// TextProps defines the properties for rendering text on a PDF.
type TextProps struct {
	Content  string
	X        float64
	Y        float64
	Width    float64
	Height   float64
	FontSize float64
	Bold     bool
	Italic   bool
	Align    string // "L", "C", "R"
	Color    RGBColor
}

// TextRenderer handles text rendering on PDF documents.
type TextRenderer struct {
	pdf *fpdf.Fpdf
}

// NewTextRenderer creates a new TextRenderer instance.
func NewTextRenderer(pdf *fpdf.Fpdf) *TextRenderer {
	return &TextRenderer{
		pdf: pdf,
	}
}

// Render draws text on the PDF with the specified properties.
func (tr *TextRenderer) Render(props TextProps) error {
	if tr.pdf == nil {
		return errors.New("errNilPDF")
	}

	// Set color
	tr.pdf.SetTextColor(props.Color.R, props.Color.G, props.Color.B)

	// Set font style
	style := ""
	if props.Bold {
		style += "B"
	}
	if props.Italic {
		style += "I"
	}
	tr.pdf.SetFont("Arial", style, props.FontSize)

	// Set text alignment
	align := props.Align
	if align == "" {
		align = "L"
	}

	// Render multi-line text cell
	tr.pdf.SetXY(props.X, props.Y)
	tr.pdf.MultiCell(props.Width, props.Height, props.Content, "", align, false)

	return nil
}

// GetHeight calculates the height needed to render the text.
func (tr *TextRenderer) GetHeight(props TextProps) float64 {
	if tr.pdf == nil {
		return 0
	}

	// Set font to calculate text height
	style := ""
	if props.Bold {
		style += "B"
	}
	if props.Italic {
		style += "I"
	}
	tr.pdf.SetFont("Arial", style, props.FontSize)

	// Get number of lines needed
	lines := tr.pdf.SplitText(props.Content, props.Width)
	height := float64(len(lines)) * props.Height

	return height
}
