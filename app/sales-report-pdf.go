package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"codeberg.org/go-pdf/fpdf"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
)

type SalesReportPDF struct {
	pdf          *fpdf.Fpdf
	trans        *i18n.Translator
	form         *ReportSalesForm
	bottomMargin float64
	topBuffer    float64 // 🔑 new: default spacing below header
}

func NewSalesReportPDF(trans *i18n.Translator, form *ReportSalesForm) (*SalesReportPDF, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 20, 15)
	pdf.SetAutoPageBreak(true, 25)
	pdf.AliasNbPages("")

	if err := registerFonts(pdf); err != nil {
		return nil, err
	}

	pdf.SetFont("DejaVu", "", 11)

	trans.Load("reports", "customers", "items")

	return &SalesReportPDF{
		pdf:          pdf,
		trans:        trans,
		form:         form,
		bottomMargin: 25,
		topBuffer:    5, // 🔑 default buffer spacing
	}, pdf.Error()
}

func (r *SalesReportPDF) AddLogo() {
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

func (r *SalesReportPDF) Header(ctx context.Context) {
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

func (r *SalesReportPDF) Footer(text string) {
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

func (r *SalesReportPDF) RenderReport(ctx context.Context, groups any, footerText string) {
	r.Header(ctx)
	r.Footer(footerText)
	switch r.form.ReportType {
	case "sales_by_customer":
		if r.form.ShowInvoices {
			reportGroups := groups.([]CustomerGroup)
			r.RenderCustomerInvoices(reportGroups)
		} else {
			reportGroups := groups.([]CustomerGroup)
			r.RenderGroups(toReportGroups(reportGroups), r.trans.Trans("reports.pdf.sales.title", i18n.Replacements{"type": "@customers.title"}))
		}
	case "sales_by_item":
		reportGroups := groups.([]ItemGroup)
		r.RenderGroups(toReportGroups(reportGroups), r.trans.Trans("reports.pdf.sales.title", i18n.Replacements{"type": "@items.title"}))
	case "sales_by_date":
		if r.form.ShowInvoices {
			r.RenderDateInvoices(groups.([]DateGroup))
		} else {
			r.RenderGroups(toReportGroups(groups.([]DateGroup)), r.trans.Trans("reports.pdf.sales.title", i18n.Replacements{"type": "@global.date"}))
		}
	default:
		panic(fmt.Sprintf("unsupported report type: %s", r.form.ReportType))
	}
}

// Summary renderer (customers, items, etc.)
func (r *SalesReportPDF) RenderGroups(groups []ReportGroup, title string) {
	if len(groups) == 0 {
		r.RenderEmptyReport(title)
		return
	}

	if r.pdf.PageNo() == 0 {
		r.pdf.AddPage()
	}

	r.pdf.Ln(r.topBuffer)
	r.pdf.SetFont("DejaVu", "B", 14)
	r.pdf.CellFormat(0, 10, title, "", 1, "", false, 0, "")

	var grandTotal float64
	for idx, group := range groups {
		rowHeight := 6.0
		r.ReserveSpace(rowHeight, r.topBuffer)

		r.pdf.SetFont("DejaVu", "", 11)
		r.pdf.CellFormat(155, rowHeight, group.Name(), "", 0, "", false, 0, "")
		r.pdf.CellFormat(25, rowHeight, foundation.FormatAmount(group.TotalAmount()), "", 1, "R", false, 0, "")

		if idx == len(groups)-1 {
			lineY := r.pdf.GetY() + 2
			r.pdf.SetDrawColor(200, 200, 200)
			r.pdf.Line(15, lineY, 195, lineY)
			r.pdf.Ln(3)
		}

		grandTotal += group.TotalAmount()
	}

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(155, 10, r.trans.Trans("reports.pdf.grandTotal"), "", 0, "", false, 0, "")
	r.pdf.CellFormat(25, 10, foundation.FormatAmount(grandTotal), "", 1, "R", false, 0, "")
}

// Detailed renderer (customer invoices)
func (r *SalesReportPDF) RenderCustomerInvoices(groups []CustomerGroup) {
	if len(groups) == 0 {
		r.RenderEmptyReport(r.trans.Trans("reports.pdf.sales.titleWithDetail", i18n.Replacements{"type": "@customers.title"}))
		return
	}

	if r.pdf.PageNo() == 0 {
		r.pdf.AddPage()
	}

	r.pdf.Ln(r.topBuffer)
	r.pdf.SetFont("DejaVu", "B", 14)
	r.pdf.CellFormat(0, 10, r.trans.Trans("reports.pdf.sales.titleWithDetail", i18n.Replacements{"type": "@customers.title"}), "", 1, "", false, 0, "")

	var grandTotal float64
	for _, group := range groups {
		headerHeight := 10.0
		invoiceRowHeight := 8.0

		r.ReserveSpace(headerHeight, r.topBuffer)
		r.pdf.SetFont("DejaVu", "B", 12)
		r.pdf.CellFormat(0, headerHeight, group.CustomerName, "", 1, "", false, 0, "")

		r.pdf.SetFont("DejaVu", "", 10)
		var subtotal float64
		for _, inv := range group.Invoices {
			r.ReserveSpace(invoiceRowHeight, r.topBuffer)
			date := inv.Date.Format("02-01-2006")
			left := fmt.Sprintf("%s (%s)", date, inv.Code)
			right := foundation.FormatAmount(inv.Total)

			r.pdf.CellFormat(155, invoiceRowHeight, left, "", 0, "", false, 0, "")
			r.pdf.CellFormat(25, invoiceRowHeight, right, "", 1, "R", false, 0, "")
			subtotal += inv.Total
		}

		group.Total = subtotal
		lineY := r.pdf.GetY()
		r.pdf.SetDrawColor(200, 200, 200)
		r.pdf.Line(15, lineY, 195, lineY)

		r.pdf.SetFont("DejaVu", "B", 10)
		r.pdf.CellFormat(155, invoiceRowHeight, r.trans.Trans("reports.pdf.sales.totalSales"), "", 0, "", false, 0, "")
		r.pdf.CellFormat(25, invoiceRowHeight, foundation.FormatAmount(group.Total), "", 1, "R", false, 0, "")
		r.pdf.Ln(5)

		grandTotal += group.Total
	}

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(155, 10, r.trans.Trans("reports.pdf.grandTotal"), "", 0, "", false, 0, "")
	r.pdf.CellFormat(25, 10, foundation.FormatAmount(grandTotal), "", 1, "R", false, 0, "")
}

func (r *SalesReportPDF) RenderDateInvoices(groups []DateGroup) {
	if len(groups) == 0 {
		r.RenderEmptyReport(r.trans.Trans("reports.pdf.sales.titleWithDetail", i18n.Replacements{"type": "@global.date"}))
		return
	}

	if r.pdf.PageNo() == 0 {
		r.pdf.AddPage()
	}

	r.pdf.Ln(r.topBuffer)
	r.pdf.SetFont("DejaVu", "B", 14)
	r.pdf.CellFormat(0, 10, r.trans.Trans("reports.pdf.sales.titleWithDetail", i18n.Replacements{"type": "@global.date"}), "", 1, "", false, 0, "")

	var grandTotal float64
	for _, group := range groups {
		headerHeight := 10.0
		invoiceRowHeight := 8.0

		r.ReserveSpace(headerHeight, r.topBuffer)
		r.pdf.SetFont("DejaVu", "B", 12)
		r.pdf.CellFormat(0, headerHeight, group.Date.Format("02-Jan-2006"), "", 1, "", false, 0, "")

		r.pdf.SetFont("DejaVu", "", 10)
		var subtotal float64
		for _, inv := range group.Invoices {
			r.ReserveSpace(invoiceRowHeight, r.topBuffer)
			left := inv.Code
			right := foundation.FormatAmount(inv.Total)

			r.pdf.CellFormat(155, invoiceRowHeight, left, "", 0, "", false, 0, "")
			r.pdf.CellFormat(25, invoiceRowHeight, right, "", 1, "R", false, 0, "")
			subtotal += inv.Total
		}

		lineY := r.pdf.GetY()
		r.pdf.SetDrawColor(200, 200, 200)
		r.pdf.Line(15, lineY, 195, lineY)

		r.pdf.SetFont("DejaVu", "B", 10)
		r.pdf.CellFormat(155, invoiceRowHeight, r.trans.Trans("reports.pdf.sales.totalSales"), "", 0, "", false, 0, "")
		r.pdf.CellFormat(25, invoiceRowHeight, foundation.FormatAmount(subtotal), "", 1, "R", false, 0, "")
		r.pdf.Ln(5)

		grandTotal += subtotal
	}

	r.pdf.SetFont("DejaVu", "B", 12)
	r.pdf.CellFormat(155, 10, r.trans.Trans("reports.pdf.grandTotal"), "", 0, "", false, 0, "")
	r.pdf.CellFormat(25, 10, foundation.FormatAmount(grandTotal), "", 1, "R", false, 0, "")
}

func (r *SalesReportPDF) RenderEmptyReport(title string) {
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
func (r *SalesReportPDF) NeedsNewPage(requiredHeight, buffer float64) bool {
	y := r.pdf.GetY()
	return y+requiredHeight+buffer > r.UsableHeight()
}

// ReserveSpace ensures there's enough room for the next block.
// If not, it adds a new page and applies buffer spacing below the header.
func (r *SalesReportPDF) ReserveSpace(requiredHeight, buffer float64) {
	if r.NeedsNewPage(requiredHeight, buffer) {
		r.pdf.AddPage()
		if buffer > 0 {
			r.pdf.Ln(buffer)
		}
	}
}

// UsableHeight returns the vertical space available on the current page
// after accounting for the bottom margin.
func (r *SalesReportPDF) UsableHeight() float64 {
	_, pageHeight := r.pdf.GetPageSize()
	return pageHeight - r.bottomMargin
}
