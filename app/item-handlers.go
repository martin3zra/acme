package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) itemsHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "item") {
		return
	}

	items, err := s.findItems(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Items/Index", inertia.Props{
		"translations": trans("items"),
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

func (s *Server) storeItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreItemForm) {

		err := s.storeItem(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating item: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}

func (s *Server) updateItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateItemForm) {

		err := s.updateItem(ctx.Request.Context(), ctx.Int("id"), form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}

func (s *Server) deleteItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.deleteItem(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}

func (s *Server) changeStatusItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		item, err := s.findItemByID(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return

		}

		err = s.toggleItemStatus(ctx.Request.Context(), item)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}
