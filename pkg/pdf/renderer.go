package pdf

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"codeberg.org/go-pdf/fpdf"
)

// Renderer interface for PDF rendering
type Renderer interface {
	Render(template *TemplateLayout, data map[string]any) ([]byte, error)
}

// Color represents RGB color
type Color struct {
	R *int `json:"r"`
	G *int `json:"g"`
	B *int `json:"b"`
}

// RgbValues returns RGB values with defaults (0 if nil)
func (c *Color) RgbValues() (r, g, b int) {
	if c != nil {
		if c.R != nil {
			r = *c.R
		}
		if c.G != nil {
			g = *c.G
		}
		if c.B != nil {
			b = *c.B
		}
	}
	return
}

// Margins represents page margins
type Margins struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// FontStyle represents font configuration
type FontStyle struct {
	Family *string `json:"family"`
	Style  *string `json:"style"`
	Size   float64 `json:"size"`
	Color  *Color `json:"color"`
}

// FontFamily returns the font family or default "Arial"
func (f *FontStyle) FontFamily() string {
	if f != nil && f.Family != nil && *f.Family != "" {
		return *f.Family
	}
	return "Arial"
}

// FontStyleStr returns the font style or default ""
func (f *FontStyle) FontStyleStr() string {
	if f != nil && f.Style != nil && *f.Style != "" {
		return *f.Style
	}
	return ""
}

// FontSize returns the font size or default 12
func (f *FontStyle) FontSize() float64 {
	if f != nil && f.Size > 0 {
		return f.Size
	}
	return 12
}

// TemplateLayout represents a PDF template
type TemplateLayout struct {
	PageSize    string                `json:"page_size"`
	PageFormat  string                `json:"page_format"`
	Margins     Margins               `json:"margins"`
	Elements    []json.RawMessage     `json:"elements"`
	DefaultFont *FontStyle            `json:"default_font"`
	Width       float64               `json:"width"`
	Height      float64               `json:"height"`
}

// UnmarshalJSON handles custom unmarshaling for TemplateLayout
// This allows null or missing values for required fields
func (tl *TemplateLayout) UnmarshalJSON(data []byte) error {
	type Alias TemplateLayout
	aux := &struct {
		PageSize   *string     `json:"page_size"`
		PageFormat *string     `json:"page_format"`
		*Alias
	}{
		Alias: (*Alias)(tl),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	if aux.PageSize != nil {
		tl.PageSize = *aux.PageSize
	}
	if aux.PageFormat != nil {
		tl.PageFormat = *aux.PageFormat
	}
	
	return nil
}

// fpdfRenderer implements Renderer
type fpdfRenderer struct {
	pdf                *fpdf.Fpdf
	layout             *TemplateLayout
	data               map[string]any
	variableBinder     *VariableBinder
	layoutEngine       *LayoutEngine
	paginationState    *PaginationState
}

// NewRenderer creates a new PDF renderer
func NewRenderer() Renderer {
	return &fpdfRenderer{}
}

// Render renders a template to PDF bytes
func (r *fpdfRenderer) Render(template *TemplateLayout, data map[string]any) ([]byte, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	// Set defaults
	if template.PageSize == "" {
		template.PageSize = "A4"
	}
	if template.PageFormat == "" {
		template.PageFormat = "P"
	}
	if template.Margins.Left == 0 {
		template.Margins.Left = 15
	}
	if template.Margins.Top == 0 {
		template.Margins.Top = 15
	}
	if template.Margins.Right == 0 {
		template.Margins.Right = 15
	}
	if template.Margins.Bottom == 0 {
		template.Margins.Bottom = 15
	}
	if template.Width == 0 {
		template.Width = 210
	}
	if template.Height == 0 {
		template.Height = 297
	}

	// Create PDF
	pdf := fpdf.New(template.PageFormat, "mm", template.PageSize, "")
	pdf.SetMargins(template.Margins.Left, template.Margins.Top, template.Margins.Right)
	pdf.SetAutoPageBreak(true, template.Margins.Bottom)
	pdf.AddPage()

	// Initialize renderer
	r.pdf = pdf
	r.layout = template
	r.data = data
	r.variableBinder = NewVariableBinder(data)
	r.layoutEngine = NewLayoutEngine()
	r.paginationState = NewPaginationState(template)

	// Render elements
	for _, elemJSON := range template.Elements {
		if err := r.renderElement(elemJSON); err != nil {
			return nil, err
		}
	}

	// Output
	w := &BytesWriter{}
	if err := pdf.Output(w); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// renderElement renders a single element
func (r *fpdfRenderer) renderElement(elemJSON json.RawMessage) error {
	var elementData map[string]any
	if err := json.Unmarshal(elemJSON, &elementData); err != nil {
		return fmt.Errorf("parse element: %w", err)
	}

	currentY := r.paginationState.CurrentY
	if err := r.renderElementData(elementData, r.layout.Margins.Left, r.paginationState.GetContentWidth(), &currentY); err != nil {
		return err
	}

	r.paginationState.CurrentY = currentY
	return nil
}

func (r *fpdfRenderer) renderElementData(elementData map[string]any, containerX float64, containerWidth float64, currentY *float64) error {
	elementType := strings.ToLower(getString(elementData, "type", ""))

	switch elementType {
	case "text":
		return r.renderTextElement(elementData, containerX, containerWidth, currentY)
	case "divider":
		return r.renderDividerElement(elementData, containerX, containerWidth, currentY)
	case "grid":
		return r.renderGridElement(elementData, currentY)
	case "table":
		return r.renderTableElement(elementData, containerX, containerWidth, currentY)
	default:
		return nil
	}
}

func (r *fpdfRenderer) renderTextElement(elementData map[string]any, containerX float64, containerWidth float64, currentY *float64) error {
	content := r.variableBinder.Bind(getString(elementData, "content", ""))
	content = cleanUTF8Text(content) // Convert UTF-8 special characters to ASCII equivalents
	fontSize := getFloat(elementData, "size", getFloat(elementData, "font_size", 0))
	if fontSize <= 0 {
		if r.layout.DefaultFont != nil {
			fontSize = r.layout.DefaultFont.FontSize()
		} else {
			fontSize = 10
		}
	}

	lineHeight := getFloat(elementData, "height", 0)
	if lineHeight <= 0 {
		lineHeight = math.Max(4.0, fontSize*0.45)
	}

	if explicitY := getFloat(elementData, "y", 0); explicitY > 0 {
		*currentY = explicitY
	}

	r.ensurePage(currentY, lineHeight)

	drawX := containerX
	if explicitX := getFloat(elementData, "x", 0); explicitX > 0 {
		drawX = explicitX
	}

	drawWidth := getFloat(elementData, "width", 0)
	if drawWidth <= 0 {
		drawWidth = containerWidth
	}
	if drawWidth <= 0 {
		drawWidth = r.paginationState.GetContentWidth()
	}

	fontStyle := ""
	if getBool(elementData, "bold", false) {
		fontStyle += "B"
	}
	if getBool(elementData, "italic", false) {
		fontStyle += "I"
	}

	fontFamily := "Arial"
	if r.layout.DefaultFont != nil {
		fontFamily = r.layout.DefaultFont.FontFamily()
	}

	defaultRed, defaultGreen, defaultBlue := 0, 0, 0
	if r.layout.DefaultFont != nil && r.layout.DefaultFont.Color != nil {
		defaultRed, defaultGreen, defaultBlue = r.layout.DefaultFont.Color.RgbValues()
	}

	textRed, textGreen, textBlue := parseColorValue(elementData["color"], defaultRed, defaultGreen, defaultBlue)
	alignment := normalizeAlign(getString(elementData, "align", "L"))

	r.pdf.SetFont(fontFamily, fontStyle, fontSize)
	r.pdf.SetTextColor(textRed, textGreen, textBlue)
	r.pdf.SetXY(drawX, *currentY)
	r.pdf.MultiCell(drawWidth, lineHeight, content, "", alignment, false)

	nextY := r.pdf.GetY()
	if nextY <= *currentY {
		nextY = *currentY + lineHeight
	}

	*currentY = nextY + getFloat(elementData, "spacing", 0)
	return nil
}

func (r *fpdfRenderer) renderDividerElement(elementData map[string]any, containerX float64, containerWidth float64, currentY *float64) error {
	spacing := getFloat(elementData, "spacing", 5)
	if spacing <= 0 {
		spacing = 5
	}

	thickness := getFloat(elementData, "thickness", 0.3)
	if thickness <= 0 {
		thickness = 0.3
	}

	lineY := *currentY + (spacing / 2)
	r.ensurePage(&lineY, thickness)

	lineRed, lineGreen, lineBlue := parseColorValue(elementData["color"], 200, 200, 200)
	r.pdf.SetDrawColor(lineRed, lineGreen, lineBlue)
	r.pdf.SetLineWidth(thickness)
	r.pdf.Line(containerX, lineY, containerX+containerWidth, lineY)

	*currentY = lineY + (spacing / 2)
	return nil
}

func (r *fpdfRenderer) renderGridElement(elementData map[string]any, currentY *float64) error {
	columnsData, ok := elementData["columns"].([]any)
	if !ok || len(columnsData) == 0 {
		return nil
	}

	grid := NewGrid(r.layout.Width, r.layout.Margins.Left, r.layout.Margins.Right)
	baseY := *currentY
	maxY := baseY
	columnOffset := 0

	for _, columnAny := range columnsData {
		columnData, ok := columnAny.(map[string]any)
		if !ok {
			continue
		}

		span := int(getFloat(columnData, "span", 1))
		if span < 1 {
			span = 1
		}
		if span > grid.Columns {
			span = grid.Columns
		}

		columnX := grid.CalculateXPosition(columnOffset, r.layout.Margins.Left)
		columnWidth := grid.CalculateWidth(span)
		columnY := baseY

		elementsData, ok := columnData["elements"].([]any)
		if ok {
			for _, nestedElementAny := range elementsData {
				nestedElementData, ok := nestedElementAny.(map[string]any)
				if !ok {
					continue
				}
				if err := r.renderElementData(nestedElementData, columnX, columnWidth, &columnY); err != nil {
					return err
				}
			}
		}

		if columnY > maxY {
			maxY = columnY
		}
		columnOffset += span
	}

	*currentY = maxY + getFloat(elementData, "spacing", 0)
	return nil
}

func (r *fpdfRenderer) renderTableElement(elementData map[string]any, containerX float64, containerWidth float64, currentY *float64) error {
	headersData, hasHeaders := elementData["headers"].([]any)
	rowsData, hasRows := elementData["rows"].([]any)
	if !hasRows || len(rowsData) == 0 {
		return nil
	}

	tableX := containerX
	if explicitX := getFloat(elementData, "x", 0); explicitX > 0 {
		tableX = explicitX
	}

	tableWidth := getFloat(elementData, "width", 0)
	if tableWidth <= 0 {
		tableWidth = containerWidth
	}
	if tableWidth <= 0 {
		tableWidth = r.paginationState.GetContentWidth()
	}

	rowHeight := getFloat(elementData, "row_height", 8)
	if rowHeight <= 0 {
		rowHeight = 8
	}

	columnCount := 0
	if hasHeaders && len(headersData) > 0 {
		columnCount = len(headersData)
	} else {
		firstRow, ok := rowsData[0].(map[string]any)
		if ok {
			if firstRowCells, ok := firstRow["cells"].([]any); ok {
				columnCount = len(firstRowCells)
			}
		}
	}
	if columnCount == 0 {
		return nil
	}

	columnWidths := make([]float64, columnCount)
	totalDefinedWidth := 0.0

	if hasHeaders && len(headersData) > 0 {
		for index := 0; index < columnCount; index++ {
			headerData, ok := headersData[index].(map[string]any)
			if !ok {
				continue
			}
			columnWidths[index] = getFloat(headerData, "width", 0)
			totalDefinedWidth += columnWidths[index]
		}
	}

	if totalDefinedWidth <= 0 {
		equalWidth := tableWidth / float64(columnCount)
		for index := range columnWidths {
			columnWidths[index] = equalWidth
		}
	} else if totalDefinedWidth != tableWidth {
		scale := tableWidth / totalDefinedWidth
		for index := range columnWidths {
			columnWidths[index] = columnWidths[index] * scale
		}
	}

	borderRed, borderGreen, borderBlue := parseColorValue(elementData["border_color"], 180, 180, 180)
	headerRed, headerGreen, headerBlue := parseColorValue(elementData["header_bg_color"], 240, 240, 240)
	r.pdf.SetDrawColor(borderRed, borderGreen, borderBlue)

	if hasHeaders && len(headersData) > 0 {
		r.ensurePage(currentY, rowHeight)
		r.pdf.SetXY(tableX, *currentY)
		r.pdf.SetFillColor(headerRed, headerGreen, headerBlue)
		r.pdf.SetTextColor(0, 0, 0)
		r.pdf.SetFont("Arial", "B", 9)

		for index := 0; index < columnCount; index++ {
			headerData, ok := headersData[index].(map[string]any)
			if !ok {
				r.pdf.CellFormat(columnWidths[index], rowHeight, "", "1", 0, "C", true, 0, "")
				continue
			}

			headerLabel := r.variableBinder.Bind(getString(headerData, "label", ""))
			headerAlign := normalizeAlign(getString(headerData, "align", "C"))
			r.pdf.CellFormat(columnWidths[index], rowHeight, headerLabel, "1", 0, headerAlign, true, 0, "")
		}
		r.pdf.Ln(-1)
		*currentY = r.pdf.GetY()
	}

	r.pdf.SetFont("Arial", "", 9)
	r.pdf.SetTextColor(0, 0, 0)

	for _, rowAny := range rowsData {
		rowData, ok := rowAny.(map[string]any)
		if !ok {
			continue
		}
		cellsData, ok := rowData["cells"].([]any)
		if !ok {
			continue
		}

		r.ensurePage(currentY, rowHeight)
		r.pdf.SetXY(tableX, *currentY)

		for index := 0; index < columnCount; index++ {
			cellText := ""
			cellAlign := "L"
			if index < len(cellsData) {
				if cellData, ok := cellsData[index].(map[string]any); ok {
					cellText = r.variableBinder.Bind(getString(cellData, "text", ""))
					cellAlign = normalizeAlign(getString(cellData, "align", "L"))
				}
			}
			r.pdf.CellFormat(columnWidths[index], rowHeight, cellText, "1", 0, cellAlign, false, 0, "")
		}
		r.pdf.Ln(-1)
		*currentY = r.pdf.GetY()
	}

	*currentY += getFloat(elementData, "spacing", 0)
	return nil
}

func (r *fpdfRenderer) ensurePage(currentY *float64, elementHeight float64) {
	if currentY == nil || elementHeight <= 0 {
		return
	}

	maxY := r.layout.Height - r.layout.Margins.Bottom
	if *currentY+elementHeight > maxY {
		r.pdf.AddPage()
		*currentY = r.layout.Margins.Top
		r.paginationState.NewPage()
	}
}

func normalizeAlign(align string) string {
	upperAlign := strings.ToUpper(strings.TrimSpace(align))
	switch upperAlign {
	case "L", "C", "R", "J":
		return upperAlign
	default:
		return "L"
	}
}

func parseColorValue(rawColor any, defaultRed int, defaultGreen int, defaultBlue int) (int, int, int) {
	if rawColor == nil {
		return defaultRed, defaultGreen, defaultBlue
	}

	switch colorValue := rawColor.(type) {
	case string:
		parts := strings.Split(colorValue, ",")
		if len(parts) != 3 {
			return defaultRed, defaultGreen, defaultBlue
		}

		red, redErr := strconv.Atoi(strings.TrimSpace(parts[0]))
		green, greenErr := strconv.Atoi(strings.TrimSpace(parts[1]))
		blue, blueErr := strconv.Atoi(strings.TrimSpace(parts[2]))
		if redErr != nil || greenErr != nil || blueErr != nil {
			return defaultRed, defaultGreen, defaultBlue
		}
		return red, green, blue
	case map[string]any:
		return int(getFloat(colorValue, "r", float64(defaultRed))), int(getFloat(colorValue, "g", float64(defaultGreen))), int(getFloat(colorValue, "b", float64(defaultBlue)))
	default:
		return defaultRed, defaultGreen, defaultBlue
	}
}

// BytesWriter implements io.Writer
type BytesWriter struct {
	buf []byte
}

// cleanUTF8Text converts UTF-8 special characters to ASCII equivalents for fpdf compatibility
func cleanUTF8Text(text string) string {
	// Replace common UTF-8 special characters with ASCII equivalents
	text = strings.ReplaceAll(text, "\u2022", "-")  // Bullet point
	text = strings.ReplaceAll(text, "\u2013", "-")  // En dash
	text = strings.ReplaceAll(text, "\u2014", "-")  // Em dash
	text = strings.ReplaceAll(text, "\u2018", "'")  // Left single quote
	text = strings.ReplaceAll(text, "\u2019", "'")  // Right single quote
	text = strings.ReplaceAll(text, "\u201C", "\"") // Left double quote
	text = strings.ReplaceAll(text, "\u201D", "\"") // Right double quote
	text = strings.ReplaceAll(text, "\u2026", "...") // Ellipsis
	text = strings.ReplaceAll(text, "\u00AE", "(R)") // Registered trademark
	text = strings.ReplaceAll(text, "\u2122", "(TM)") // Trademark
	text = strings.ReplaceAll(text, "\u00A9", "(C)") // Copyright
	return text
}

func (w *BytesWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *BytesWriter) Bytes() []byte {
	return w.buf
}

// Helper functions
func getString(data map[string]any, key string, defaultVal string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultVal
}

func getFloat(data map[string]any, key string, defaultVal float64) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return defaultVal
}

func getBool(data map[string]any, key string, defaultVal bool) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return defaultVal
}

func getColor(data map[string]any, key string) *Color {
	if colorData, ok := data[key].(map[string]any); ok {
		r := int(getFloat(colorData, "r", 0))
		g := int(getFloat(colorData, "g", 0))
		b := int(getFloat(colorData, "b", 0))
		return &Color{
			R: &r,
			G: &g,
			B: &b,
		}
	}
	return nil
}
