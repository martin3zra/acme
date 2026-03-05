package pdf

// Grid implements a 12-column grid system (like Tailwind CSS)
type Grid struct {
	TotalWidth  float64
	Columns     int // Always 12
	ColumnWidth float64
}

// NewGrid creates a new grid system
func NewGrid(pageWidth float64, marginLeft float64, marginRight float64) *Grid {
	totalWidth := pageWidth - marginLeft - marginRight
	return &Grid{
		TotalWidth:  totalWidth,
		Columns:     12,
		ColumnWidth: totalWidth / 12,
	}
}

// CalculateWidth calculates the width of a column based on span
func (g *Grid) CalculateWidth(span int) float64 {
	if span < 1 {
		span = 1
	}
	if span > g.Columns {
		span = g.Columns
	}
	return g.ColumnWidth * float64(span)
}

// ValidateSpan validates that a span is within column bounds
func (g *Grid) ValidateSpan(span int) bool {
	return span > 0 && span <= g.Columns
}

// CalculateXPosition calculates the X position based on column offset
func (g *Grid) CalculateXPosition(offset int, marginLeft float64) float64 {
	return marginLeft + (g.ColumnWidth * float64(offset))
}

// CalculateOffset converts pixel position to column offset
func (g *Grid) CalculateOffset(x float64, marginLeft float64) int {
	return int((x - marginLeft) / g.ColumnWidth)
}
