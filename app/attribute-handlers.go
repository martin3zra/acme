package app

import (
	"fmt"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/validator/locale"
)

// attributesHandler returns list of attributes
func (s *Server) attributesHandler(ctx *routing.Context) {
	attributes, err := s.findAttributes(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Attributes/Index", map[string]any{
		"attributes": attributes,
	})
}

// storeAttributeHandler creates a new attribute
func (s *Server) storeAttributeHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeForm) {
		err := s.storeAttribute(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.attribute"}))
		ctx.Redirect("/attributes")
	})
}

// updateAttributeHandler updates an existing attribute
func (s *Server) updateAttributeHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeForm) {
		attributeID := ctx.Int("id")

		err := s.updateAttribute(ctx.Request.Context(), attributeID, form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.attribute"}))
		ctx.Redirect("/attributes")
	})
}

// deleteAttributeHandler soft-deletes an attribute
func (s *Server) deleteAttributeHandler() routing.HandlerFunc {
	type form struct {
		ConfirmsPasswords
	}

	return routing.WithRequest(func(ctx *routing.Context, frm *form) {
		attributeID := ctx.Int("id")

		// Check if attribute is in use
		var count int64
		err := s.db.QueryRowContext(
			ctx.Request.Context(),
			`SELECT COUNT(*) FROM product_attributes WHERE attribute_id = $1 AND company_id = $2`,
			attributeID, CurrentCompany(ctx.Request.Context()).ID,
		).Scan(&count)

		if err != nil {
			ctx.BackWithError(err)
			return
		}

		if count > 0 {
			ctx.BackWith("error", "Cannot delete attribute that is in use by products")
			return
		}

		// Soft delete all attribute values first
		_, err = s.db.ExecContext(
			ctx.Request.Context(),
			`UPDATE attribute_values SET deleted_at = NOW() WHERE attribute_id = $1 AND company_id = $2 AND deleted_at IS NULL`,
			attributeID, CurrentCompany(ctx.Request.Context()).ID,
		)

		if err != nil {
			ctx.BackWithError(err)
			return
		}

		// Then soft delete the attribute
		_, err = s.db.ExecContext(
			ctx.Request.Context(),
			`UPDATE attributes SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND company_id = $2`,
			attributeID, CurrentCompany(ctx.Request.Context()).ID,
		)

		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", fmt.Sprintf(locale.SpanishMessages()["messages.deleted"].(string), locale.SpanishMessages()["global.attribute"]))
		ctx.Redirect("/attributes")
	})
}

// attributeValuesHandler returns list of values for an attribute
func (s *Server) attributeValuesHandler(ctx *routing.Context) {
	attributeID := ctx.Int("id")

	attribute, err := s.findAttributeByID(ctx.Request.Context(), attributeID)
	if err != nil {
		ctx.Error(err)
		return
	}

	values, err := s.findAttributeValuesByAttribute(ctx.Request.Context(), attributeID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Attributes/Values", map[string]any{
		"attribute": attribute,
		"values":    values,
	})
}

// storeAttributeValueHandler creates a new attribute value
func (s *Server) storeAttributeValueHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeValueForm) {
		form.AttributeID = ctx.Int("id")

		err := s.storeAttributeValue(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.attributeValue"}))
		ctx.Redirect(fmt.Sprintf("/attributes/%d/values", form.AttributeID))
	})
}

// updateAttributeValueHandler updates an attribute value
func (s *Server) updateAttributeValueHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAttributeValueForm) {
		valueID := ctx.Int("id")

		err := s.updateAttributeValue(ctx.Request.Context(), valueID, form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.attributeValue"}))

		// Get attribute ID to redirect back
		form.AttributeID = ctx.Int("attribute_id")
		ctx.Redirect(fmt.Sprintf("/attributes/%d/values", form.AttributeID))
	})
}

// deleteAttributeValueHandler soft-deletes an attribute value
func (s *Server) deleteAttributeValueHandler() routing.HandlerFunc {
	type form struct {
		ConfirmsPasswords
	}

	return routing.WithRequest(func(ctx *routing.Context, frm *form) {
		valueID := ctx.Int("id")
		attributeID := ctx.Int("attribute_id")

		// Check if value is in use
		var count int64
		err := s.db.QueryRowContext(
			ctx.Request.Context(),
			`SELECT COUNT(*) FROM variant_attribute_values WHERE attribute_value_id = $1 AND company_id = $2`,
			valueID, CurrentCompany(ctx.Request.Context()).ID,
		).Scan(&count)

		if err != nil {
			ctx.BackWithError(err)
			return
		}

		if count > 0 {
			ctx.BackWith("error", "Cannot delete attribute value that is used by variants")
			return
		}

		err = s.deleteAttributeValue(ctx.Request.Context(), valueID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", fmt.Sprintf(locale.SpanishMessages()["messages.deleted"].(string), locale.SpanishMessages()["global.attributeValue"]))
		ctx.Redirect(fmt.Sprintf("/attributes/%d/values", attributeID))
	})
}

