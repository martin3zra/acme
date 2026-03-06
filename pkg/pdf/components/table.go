package components

import (
	"codeberg.org/go-pdf/fpdf"
)

// TableProps defines the properties for rendering tables on a PDF.
type TableProps struct {
	X           float64
	Y           float64
	Width       float64
	Headers     []string
	Rows        [][]string
	RowHeight   float64
	BorderColor RGBColor
	HeaderColor RGBColor
	TextColor   RGBColor
	HeaderAlign string
	RowAlign    string
}

// TableRenderer handles table rendering on PDF documents.
type TableRenderer struct {
	pdf *fpdf.Fpdf
}

// NewTableRenderer creates a new TableRenderer instance.
func NewTableRenderer(pdf *fpdf.Fpdf) *TableRenderer {
	return &TableRenderer{
		pdf: pdf,
	}
}

// Render draws a table on the PDF with the specified properties.
func (tr *TableRenderer) Render(props TableProps) error {
	if tr.pdf == nil {
		return ErrNilPDF
	}

	// Set border color
	tr.pdf.SetDrawColor(props.BorderColor.R, props.BorderColor.G, props.BorderColor.B)

	// Calculate column width
	columnWidth := props.Width / float64(len(props.Headers))

	// Render header
	tr.pdf.SetXY(props.X, props.Y)
	tr.pdf.SetFillColor(props.HeaderColor.R, props.HeaderColor.G, props.HeaderColor.B)
	tr.pdf.SetTextColor(255, 255, 255) // White text for header
	tr.pdf.SetFont("Arial", "B", 10)

	headerAlign := props.HeaderAlign
	if headerAlign == "" {
		headerAlign = "C"
	}

	for _, header := range props.Headers {
		tr.pdf.CellFormat(columnWidth, props.RowHeight, header, "1", 0, headerAlign, true, 0, "")
	}
	tr.pdf.Ln(-1)

	// Render rows
	tr.pdf.SetFillColor(255, 255, 255)
	tr.pdf.SetTextColor(props.TextColor.R, props.TextColor.G, props.TextColor.B)
	tr.pdf.SetFont("Arial", "", 10)

	rowAlign := props.RowAlign
	if rowAlign == "" {
		rowAlign = "L"
	}

	for _, row := range props.Rows {
		tr.pdf.SetXY(props.X, tr.pdf.GetY())
		for _, cell := range row {
			tr.pdf.CellFormat(columnWidth, props.RowHeight, cell, "1", 0, rowAlign, false, 0, "")
		}
		tr.pdf.Ln(-1)
	}

	return nil
}

// GetHeight calculates the total height needed to render the table.
func (tr *TableRenderer) GetHeight(props TableProps) float64 {
	// Header row + data rows
	rowCount := float64(len(props.Rows) + 1) // +1 for header
	return rowCount * props.RowHeight
}
