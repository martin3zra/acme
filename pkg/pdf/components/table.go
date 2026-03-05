package components

import (
	"codeberg.org/go-pdf/fpdf"
)

// TableProps defines the properties for rendering tables on a PDF.
type TableProps struct {
	X            float64
	Y            float64
	Width        float64
	Headers      []string
	Rows         [][]string
	RowHeight    float64
	BorderColor  RGBColor
	HeaderColor  RGBColor
	TextColor    RGBColor
	HeaderAlign  string
	RowAlign     string
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


// TableProps defines the properties for rendering tables on a PDF.
type TableProps struct {
	X            float64
	Y            float64
	Width        float64
	Headers      []string
	Rows         [][]string
	RowHeight    float64
	BorderColor  RGBColor
	HeaderColor  RGBColor
	TextColor    RGBColor
	HeaderAlign  string
	RowAlign     string
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


















































































}	return rowHeight * float64(numRows+1)	// Header row + data rowsfunc (tr *TableRenderer) GetHeight(numRows int, rowHeight float64) float64 {// GetHeight returns the total height of the table}	pdf.SetFillColor(255, 255, 255)	pdf.SetTextColor(0, 0, 0)	// Reset colors	}		pdf.Ln(props.RowHeight)		}			pdf.CellFormat(columnWidth, props.RowHeight, cell, 1, 0, align, false, 0, "")			}				align = "L"			if align == "" {			align := props.RowAlign			}				break			if i >= len(props.Headers) {		for i, cell := range row {	for _, row := range props.Rows {	pdf.SetFillColor(255, 255, 255)	pdf.SetTextColor(int(props.TextColor.R), int(props.TextColor.G), int(props.TextColor.B))	pdf.SetFont("DejaVu", "", 9)	// Draw rows	pdf.Ln(props.RowHeight)	}		pdf.CellFormat(columnWidth, props.RowHeight, header, 1, 0, align, true, 0, "")		}			align = "C"		if align == "" {		align := props.HeaderAlign	for _, header := range props.Headers {	pdf.SetTextColor(255, 255, 255)	pdf.SetFillColor(int(props.HeaderColor.R), int(props.HeaderColor.G), int(props.HeaderColor.B))	pdf.SetFont("DejaVu", "B", 10)	// Draw header	columnWidth := props.Width / float64(len(props.Headers))	// Calculate column width	pdf.SetXY(props.X, props.Y)	}		return	if len(props.Headers) == 0 || len(props.Rows) == 0 {func (tr *TableRenderer) Render(pdf *fpdf.Fpdf, props TableProps) {// Render renders a table element on the PDF}	return &TableRenderer{}func NewTableRenderer() *TableRenderer {// NewTableRenderer creates a new table renderer}	TextColor   RGBColor	HeaderColor RGBColor	BorderColor RGBColor	RowAlign    string	Rows        [][]string	HeaderAlign string	Headers     []string	RowHeight   float64	Width       float64	Y           float64	X           float64type TableProps struct {// TableProps represents table element propertiestype TableRenderer struct{}// TableRenderer handles rendering table elements)	"codeberg.org/go-pdf/fpdf"import (