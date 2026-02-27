package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
)

type ProfitLostReportPDF struct {
	pdf          *fpdf.Fpdf
	trans        *i18n.Translator
	form         *ReportForm
	bottomMargin float64
	topBuffer    float64 // 🔑 new: default spacing below header
}

func NewProfitLostReportPDF(trans *i18n.Translator, form *ReportForm) (*ProfitLostReportPDF, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 20, 15)
	pdf.SetAutoPageBreak(true, 25)
	pdf.AliasNbPages("")

	if err := registerFonts(pdf); err != nil {
		return nil, err
	}

	pdf.SetFont("DejaVu", "", 11)

	trans.Load("reports")

	return &ProfitLostReportPDF{
		pdf:          pdf,
		trans:        trans,
		form:         form,
		bottomMargin: 25,
		topBuffer:    5, // 🔑 default buffer spacing
	}, pdf.Error()
}

func (r *ProfitLostReportPDF) AddLogo() {
	logoBytes, err := base64.StdEncoding.DecodeString(logoBase64)
	if err == nil {
		r.pdf.RegisterImageOptionsReader("logo", fpdf.ImageOptions{
			ImageType: "PNG",
		}, bytes.NewReader(logoBytes))
		r.pdf.ImageOptions("logo", 15, 10, 30, 0, false, fpdf.ImageOptions{
			ImageType: "PNG",
		}, 0, "")
	}
}

func (r *ProfitLostReportPDF) Header(ctx context.Context) {
	r.pdf.SetHeaderFunc(func() {
		r.AddLogo()

		company := CurrentCompany(ctx)
		// Company Info (top left)
		r.pdf.SetFont("DejaVu", "", 10)
		r.pdf.SetXY(50, 10)
		r.pdf.MultiCell(80, 5, fmt.Sprintf("%s\n%s\n%s", company.Name, company.Address+"\n"+company.City, ""), "", "", false)
		r.pdf.SetXY(50, 25)
		r.pdf.CellFormat(80, 6, fmt.Sprintf("%s: %s", r.trans.Trans("companies.single.rnc_short"), company.Identifier), "", 1, "", false, 0, "")

		// Info (top right)
		r.pdf.SetXY(150, 10)
		r.pdf.CellFormat(0, 6, fmt.Sprintf("%s - %s", r.form.From.Format("2006-01-02"), r.form.To.Format("2006-01-02")), "", 1, "R", false, 0, "")

		r.pdf.Ln(12)
	})
}

func (r *ProfitLostReportPDF) Footer(text string) {
	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-15)
		r.pdf.SetFont("DejaVu", "I", 8)
		r.pdf.SetTextColor(0, 0, 0)
		footer := fmt.Sprintf("%s %d/{nb}", r.trans.Trans("global.pagination.page"), r.pdf.PageNo())
		r.pdf.CellFormat(
			10, 10, r.trans.Trans("reports.pdf.generatedBy", map[string]string{"platform": text}), "", 0, "L", false, 0, "",
		)
		r.pdf.CellFormat(
			0, 10, footer, "", 0, "C", false, 0, "",
		)
	})
}

func (r *ProfitLostReportPDF) RenderReport(ctx context.Context, totalSales float64, report []*expenseCategory, from, to time.Time, footerText string) {
	r.Header(ctx)
	r.Footer(footerText)

	if len(report) == 0 && totalSales == 0 {
		r.RenderEmptyReport(r.trans.Trans("reports.pdf.profit-lost.title"))
		return
	}

	if r.pdf.PageNo() == 0 {
		r.pdf.AddPage()
	}
	r.pdf.Ln(r.topBuffer)

	r.pdf.SetFont("DejaVu", "B", 14)
	r.pdf.CellFormat(0, 10, r.trans.Trans("reports.pdf.profit-lost.title"), "", 1, "L", false, 0, "")

	r.pdf.Ln(4)

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(155, 10, r.trans.Trans("reports.income"), "", 0, "", false, 0, "")
	r.pdf.CellFormat(25, 10, foundation.FormatAmount(totalSales), "", 1, "R", false, 0, "")

	r.pdf.Ln(1)

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(0, 10, r.trans.Trans("reports.expenses"), "0", 0, "L", false, 0, "")
	r.pdf.Ln(10)

	x := r.pdf.GetX()
	y := r.pdf.GetY()

	var grandTotal float64
	r.pdf.SetFont("DejaVu", "", 12)
	for idx, row := range report {
		// Add inner padding manually
		r.pdf.SetX(x + 4)
		rowHeight := 6.0
		r.ReserveSpace(rowHeight, r.topBuffer)

		r.pdf.CellFormat(155-4, 10, row.Name, "0", 0, "L", false, 0, "")
		r.pdf.CellFormat(25, 10, fmt.Sprintf("%s", foundation.FormatAmount(row.TotalAmount)), "0", 1, "R", false, 0, "")

		if idx == len(report)-1 {
			lineY := r.pdf.GetY() + 2
			r.pdf.SetDrawColor(200, 200, 200)
			r.pdf.Line(15, lineY, 195, lineY)
			r.pdf.Ln(3)
		}

		grandTotal += row.TotalAmount
	}

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(155, 10, r.trans.Trans("reports.pdf.grandTotal"), "", 0, "", false, 0, "")
	r.pdf.CellFormat(25, 10, foundation.FormatAmount(grandTotal), "", 1, "R", false, 0, "")

	r.pdf.Ln(10)

	// net := totalSales - grandTotal
	// if net >= 0 {
	// 	r.pdf.SetFillColor(220, 255, 220) // green
	// } else {
	// 	r.pdf.SetFillColor(255, 220, 220) // red
	// }
	// r.pdf.Ln(3)
	// r.pdf.SetFont("DejaVu", "B", 16)
	// r.pdf.CellFormat(155, 10+2, r.trans.Trans("reports.pdf.netProfit"), "", 0, "", true, 0, "")
	// r.pdf.CellFormat(25, 10+2, foundation.FormatAmount(net), "", 1, "R", true, 0, "")
	// r.pdf.Ln(3)
	// // reset fill color
	// r.pdf.SetFillColor(255, 255, 255)
	net := totalSales - grandTotal

	x = r.pdf.GetX()
	y = r.pdf.GetY()
	width := 180.0
	height := 14.0

	if net >= 0 {
		r.pdf.SetFillColor(220, 255, 220)
	} else {
		r.pdf.SetFillColor(255, 220, 220)
	}

	// Draw background block
	r.pdf.Rect(x, y, width, height, "F")

	// Add inner padding manually
	r.pdf.SetXY(x+4, y+4)

	r.pdf.SetFont("DejaVu", "B", 16)

	r.pdf.CellFormat(147, 6, r.trans.Trans("reports.pdf.netProfit"), "", 0, "L", false, 0, "")
	r.pdf.CellFormat(25, 6, foundation.FormatAmount(net), "", 1, "R", false, 0, "")

	// Move cursor below block
	r.pdf.SetY(y + height)

	// Reset fill
	r.pdf.SetFillColor(255, 255, 255)
}

func (r *ProfitLostReportPDF) RenderEmptyReport(title string) {
	if r.pdf.PageNo() == 0 {
		r.pdf.AddPage()
	}

	r.pdf.Ln(4)
	r.pdf.SetFont("DejaVu", "B", 14)
	r.pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")

	r.pdf.Ln(10)
	r.pdf.SetFont("DejaVu", "", 12)
	msg := r.trans.Trans("reports.pdf.noDataInRange", map[string]string{"start": r.form.From.Format("2006-01-02"), "end": r.form.To.Format("2006-01-02")})
	r.pdf.MultiCell(0, 10, msg, "", "C", false)
}

// NeedsNewPage checks if the required height fits in the remaining space.
// Buffer lets you reserve extra space (e.g. below header).
func (r *ProfitLostReportPDF) NeedsNewPage(requiredHeight, buffer float64) bool {
	y := r.pdf.GetY()
	return y+requiredHeight+buffer > r.UsableHeight()
}

// ReserveSpace ensures there's enough room for the next block.
// If not, it adds a new page and applies buffer spacing below the header.
func (r *ProfitLostReportPDF) ReserveSpace(requiredHeight, buffer float64) {
	if r.NeedsNewPage(requiredHeight, buffer) {
		r.pdf.AddPage()
		if buffer > 0 {
			r.pdf.Ln(buffer)
		}
	}
}

// UsableHeight returns the vertical space available on the current page
// after accounting for the bottom margin.
func (r *ProfitLostReportPDF) UsableHeight() float64 {
	_, pageHeight := r.pdf.GetPageSize()
	return pageHeight - r.bottomMargin
}
