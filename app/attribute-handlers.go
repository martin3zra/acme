package app

import (
	"fmt"
	"time"

	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
)

// attributesHandler returns the list of attributes.
func (s *Server) attributesHandler(ctx *routing.Context) {
	attributes, err := s.findAttributes(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Attributes/Index", map[string]any{
		"translations": trans("attributes"),
		"attributes":   attributes,
	})
}

// storeAttributeHandler creates a new attribute.
func (s *Server) storeAttributeHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeForm) {
		if err := s.storeAttribute(ctx.Request.Context(), form); err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.attribute"}))
		ctx.Redirect("/attributes")
	})
}

// updateAttributeHandler updates an existing attribute.
func (s *Server) updateAttributeHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeForm) {
		attributeID := ctx.Int("id")

		if err := s.updateAttribute(ctx.Request.Context(), attributeID, form); err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.attribute"}))
		ctx.Redirect("/attributes")
	})
}

// deleteAttributeHandler soft-deletes an attribute (and its values) when it is
// not in use by any product.
func (s *Server) deleteAttributeHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		companyID := CurrentCompany(ctx.Request.Context()).ID
		attributeID := ctx.Int("id")

		pdb, err := s.play()
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		count, err := pdb.Model(&productAttributeRead{}).
			WhereEq("company_id", companyID).
			WhereEq("attribute_id", attributeID).
			Count(ctx.Request.Context())
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		if count > 0 {
			ctx.BackWith("error", "Cannot delete attribute that is in use by products")
			return
		}

		// Bulk soft delete: an attribute with no values legitimately matches no row,
		// so this one is not guarded. attributeValueRead's softdelete tag supplies the
		// `deleted_at IS NULL` the raw statement carried.
		if _, err := pdb.Model(&attributeValueRead{}).
			WhereEq("company_id", companyID).
			WhereEq("attribute_id", attributeID).
			Update(ctx.Request.Context(), map[string]any{"deleted_at": time.Now()}); err != nil {
			ctx.BackWithError(err)
			return
		}

		// The attribute itself must match exactly one row.
		affected, err := pdb.Model(&attributeRead{}).
			WhereEq("company_id", companyID).
			WhereEq("id", attributeID).
			Update(ctx.Request.Context(), map[string]any{"deleted_at": time.Now()})
		if err := mustAffectRows(affected, err, "attribute"); err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.attribute"}))
		ctx.Redirect("/attributes")
	})
}

// attributeValuesHandler returns the list of values for an attribute.
func (s *Server) attributeValuesHandler(ctx *routing.Context) {
	attributeID := ctx.Param("id")

	attribute, err := s.findAttributeByID(ctx.Request.Context(), attributeID)
	if err != nil {
		ctx.Error(err)
		return
	}

	values, err := s.findAttributeValuesByAttribute(ctx.Request.Context(), attribute.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Attributes/Values/Index", map[string]any{
		"translations": trans("attributes"),
		"attribute":    attribute,
		"values":       values,
	})
}

// storeAttributeValueHandler creates a new attribute value.
func (s *Server) storeAttributeValueHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeValueForm) {
		attributeUUID := ctx.Param("id")

		attribute, err := s.findAttributeByID(ctx.Request.Context(), attributeUUID)
		if err != nil {
			ctx.Error(err)
			return
		}
		form.AttributeID = attribute.ID

		if err = s.storeAttributeValue(ctx.Request.Context(), form); err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.attributeValue"}))
		ctx.Redirect(fmt.Sprintf("/attributes/%s/values", attributeUUID))
	})
}

// updateAttributeValueHandler updates an attribute value.
func (s *Server) updateAttributeValueHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeValueForm) {
		valueUUID := ctx.Param("uuid")

		av, err := s.findAttributeValueByUUID(ctx.Request.Context(), valueUUID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		form.AttributeID = av.AttributeID

		if err = s.updateAttributeValue(ctx.Request.Context(), valueUUID, form); err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.attributeValue"}))

		attr, err := s.findAttributeByIntID(ctx.Request.Context(), av.AttributeID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		ctx.Redirect(fmt.Sprintf("/attributes/%s/values", attr.UUID))
	})
}

// deleteAttributeValueHandler soft-deletes an attribute value when it is not used
// by any variant.
func (s *Server) deleteAttributeValueHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		valueUUID := ctx.Param("uuid")

		av, err := s.findAttributeValueByUUID(ctx.Request.Context(), valueUUID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		pdb, err := s.play()
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		count, err := pdb.Model(&variantAttributeValueRead{}).
			WhereEq("company_id", CurrentCompany(ctx.Request.Context()).ID).
			WhereEq("attribute_value_id", av.ID).
			Count(ctx.Request.Context())
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		if count > 0 {
			ctx.BackWith("error", "Cannot delete attribute value that is used by variants")
			return
		}

		if err = s.deleteAttributeValue(ctx.Request.Context(), valueUUID); err != nil {
			ctx.BackWithError(err)
			return
		}

		attr, err := s.findAttributeByIntID(ctx.Request.Context(), av.AttributeID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.attributeValue"}))
		ctx.Redirect(fmt.Sprintf("/attributes/%s/values", attr.UUID))
	})
}
