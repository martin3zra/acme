package app

func mapPurchaseToStoreForm(p *purchase, lines []*line) *StorePurchaseForm {
	formLines := make([]*Line, 0)
	for _, line := range lines {
		var l Line
		l.ID = int(line.ID)
		l.Unit = int(line.Unit.ID)
		l.Qty = int(line.Qty)
		l.Price = line.Price
		l.Rate = line.Tax.Rate
		l.Action = LineActions.Added
		l.tax = line.Tax.Amount
		l.amount = line.Amount
		l.discount = 0
		l.total = line.Total
		formLines = append(formLines, &l)
	}

	return &StorePurchaseForm{
		Kind:          p.Kind,
		Date:          p.Date,
		Terms:         p.Terms,
		VendorID:      p.Vendor.ID,
		amount:        p.Amount,
		amountDue:     p.AmountDue,
		Discount:      p.Discount,
		tax:           p.Tax,
		total:         p.Total,
		Notes:         p.Notes,
		paymentStatus: PaidStatuses.UnPaid,
		Source:        &PurchaseSource{ID: p.UUID, Type: p.Kind, Code: p.Number},
		Lines:         formLines,
	}
}
