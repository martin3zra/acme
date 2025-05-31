package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) itemsHandler(ctx *routing.Context) {

	items, err := s.findItems(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Items/Index", inertia.Props{
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("items")),
		"items":        items,
		"units": inertia.Defer(func() (any, error) {
			units, err := s.findUnits(ctx.Request.Context())
			if err != nil {
				ctx.Error(err)
				return nil, nil
			}
			return units, err
		}, "attributes"),
		"taxes": inertia.Defer(func() (any, error) {
			taxes, err := s.findTaxes(ctx.Request.Context())
			if err != nil {
				ctx.Error(err)
				return nil, nil
			}
			return taxes, nil
		}, "attributes"),
	})
}

func (s *Server) storeItemHandler(ctx *routing.Context) {
	var form StoreItemForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.storeItem(ctx.Request.Context(), form)
	if err != nil {
		log.Printf("Error creating item: %v", err)
		s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.item"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.item"}))

	ctx.Redirect("/items")
}

func (s *Server) updateItemHandler(ctx *routing.Context) {
	var form UpdateItemForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.updateItem(ctx.Request.Context(), ctx.Int("id"), form)
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
		ctx.Back()
		return

	}

	s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.item"}))

	ctx.Redirect("/items")
}

func (s *Server) deleteItemHandler(ctx *routing.Context) {
	var form ConfirmsPasswords
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.deleteItem(ctx.Request.Context(), ctx.Int("id"))
	if err != nil {
		s.session.Errors("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.item"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.item"}))

	ctx.Redirect("/items")
}

func (s *Server) changeStatusItemHandler(ctx *routing.Context) {
	var form ConfirmsPasswords
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	item, err := s.findItemByID(ctx.Request.Context(), ctx.Int("id"))
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
		ctx.Back()
		return

	}

	err = s.toggleItemStatus(ctx.Request.Context(), item)
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.item"}))

	ctx.Redirect("/items")
}
