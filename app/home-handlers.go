package app

import (
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) homeHandler(ctx *routing.Context) {
  stats, err := s.findStats(ctx.Request.Context())
  if err != nil {
    ctx.Error(err)
    return
  }
  dueInvoices, err := s.findLatestDueInvoices(ctx.Request.Context())
  if err != nil {
    ctx.Error(err)
    return 
  }
  
	ctx.Render("Home/Index", map[string]any{
    "translations": trans("dashboard"),
    "stats": []any {
      map[string]any { "label": "dashboard.stats.total_amount_due", "value": foundation.FormatAmount(stats.TotalDueAmount), "icon": "dollar", "bg": "bg-pink-100" },
      map[string]any { "label": "dashboard.stats.customers", "value": stats.TotalCustomers, "icon": "users", "bg": "bg-blue-100" },
      map[string]any { "label": "dashboard.stats.invoices", "value": stats.TotalInvoices, "icon": "invoice", "bg": "bg-blue-200" },
      map[string]any { "label": "", "value": "", "icon": "", "bg": "" },
    },
    "due_invoices": dueInvoices,
  })
}