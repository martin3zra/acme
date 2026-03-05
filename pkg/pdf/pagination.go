package pdf

// PaginationState tracks pagination state for page breaking
type PaginationState struct {
	CurrentPage  int
	CurrentY     float64
	PageHeight   float64
	PageWidth    float64
	BottomMargin float64
	TopMargin    float64
	LeftMargin   float64
	RightMargin  float64
}

// NewPaginationState creates a new pagination state
func NewPaginationState(layout *TemplateLayout) *PaginationState {
	return &PaginationState{
		CurrentPage:  1,
		CurrentY:     layout.Margins.Top,
		PageHeight:   layout.Height,
		PageWidth:    layout.Width,
		BottomMargin: layout.Margins.Bottom,
		TopMargin:    layout.Margins.Top,
		LeftMargin:   layout.Margins.Left,
		RightMargin:  layout.Margins.Right,
	}
}

// NeedsNewPage checks if a new page is needed
func (ps *PaginationState) NeedsNewPage(elementHeight float64) bool {
	return ps.CurrentY+elementHeight > ps.PageHeight-ps.BottomMargin
}

// NewPage creates a new page
func (ps *PaginationState) NewPage() {
	ps.CurrentPage++
	ps.ResetY()
}

// UpdateY updates the current Y position
func (ps *PaginationState) UpdateY(elementHeight float64) {
	ps.CurrentY += elementHeight
}

// ResetY resets Y to top margin
func (ps *PaginationState) ResetY() {
	ps.CurrentY = ps.TopMargin
}

// GetAvailableHeight returns available height
func (ps *PaginationState) GetAvailableHeight() float64 {
	return ps.PageHeight - ps.BottomMargin - ps.CurrentY
}

// GetContentWidth returns available width
func (ps *PaginationState) GetContentWidth() float64 {
	return ps.PageWidth - ps.LeftMargin - ps.RightMargin
}

// GetCursorX returns left margin X position
func (ps *PaginationState) GetCursorX() float64 {
	return ps.LeftMargin
}