package components

import (
	"github.com/jung-kurt/gofpdf"
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
	pdf *gofpdf.Fpdf
}

// NewTextRenderer creates a new TextRenderer instance.
func NewTextRenderer(pdf *gofpdf.Fpdf) *TextRenderer {
	return &TextRenderer{
		pdf: pdf,
	}
}

// Render draws text on the PDF with the specified properties.
func (tr *TextRenderer) Render(props TextProps) error {
	if tr.pdf == nil {
		return ErrNilPDF
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


































































}	return float64(lines) * fontSize * 0.35	lines := (len(content) / int(width)) + 1	// Approximate: each line is about fontSize * 0.35 in mm	pdf.SetFont("DejaVu", "", fontSize)func (tr *TextRenderer) GetHeight(pdf *fpdf.Fpdf, content string, width float64, fontSize float64) float64 {// GetHeight returns the height needed for the text}	pdf.SetTextColor(0, 0, 0)	// Reset color	pdf.CellFormat(props.Width, props.Height, props.Content, "", 0, props.Align, false, 0, "")	// Draw text	}		props.Align = "L"	if props.Align == "" {	// Align: L=left, C=center, R=right	pdf.SetXY(props.X, props.Y)	// Position	pdf.SetTextColor(int(props.Color.R), int(props.Color.G), int(props.Color.B))	pdf.SetFont("DejaVu", style, props.FontSize)	}		style += "I"	if props.Italic {	}		style += "B"	if props.Bold {	style := ""	// Set fontfunc (tr *TextRenderer) Render(pdf *fpdf.Fpdf, props TextProps) {// Render renders a text element on the PDF}	return &TextRenderer{}func NewTextRenderer() *TextRenderer {// NewTextRenderer creates a new text renderer}	B uint8	G uint8	R uint8type RGBColor struct {// RGBColor represents a color}	Color   RGBColor	Align   string // L, C, R	Italic  bool	Bold    bool	FontSize float64	Height  float64	Width   float64	Y       float64	X       float64	Content stringtype TextProps struct {// TextProps represents text element propertiestype TextRenderer struct{}// TextRenderer handles rendering text elements)	"codeberg.org/go-pdf/fpdf"import (