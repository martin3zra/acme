package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) itemsHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "item") {
		return
	}

	itemType := ItemType(ctx.Query("itemType"))
	if err := itemType.Validate(); err != nil {
		itemType = "all"
	}

	items, err := s.findItems(ctx.Request.Context(), itemType)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Items/Index", inertia.Props{
		"translations":          trans("items"),
		"items":                 items,
		"currentItemTypeFilter": itemType,
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

func (s *Server) startUploadChunkHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UploadSessionForm) {

		// 1️⃣ Validate extension
		ext := strings.ToLower(filepath.Ext(form.Filename))
		if ext != ".csv" && ext != ".txt" {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "only CSV or TXT allowed",
			})
			return
		}

		// 2️⃣ Generate ID
		uploadID := uuid.NewString()

		// 3️⃣ Create upload directory
		dir := filepath.Join("uploads", uploadID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": "failed to create upload dir",
			})
			return
		}

		if err := s.storeUploadSession(&UploadSession{
			ID:       uploadID,
			UserID:   int64(ctx.User().Id),
			Filename: form.Filename,
			FileSize: form.Size,
			Status:   "pending",
		}); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": "failed to create upload dir",
			})
			return
		}

		ctx.JSON(http.StatusOK, map[string]any{
			"upload_id": uploadID,
		})
	})
}

func (s *Server) uploadChunkHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UploadChunkForm) {
		if form.Chunk != nil {
			defer form.Chunk.Close()
		}

		if form.TotalChunks <= 0 {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "invalid total_chunks",
			})
			return
		}

		uploadSession, err := s.findUploadSession(form.UploadId, int64(ctx.User().Id))
		if err != nil || uploadSession == nil {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": "invalid upload id",
			})
			return
		}

		// 🔥 Set total_chunks ONCE
		if !uploadSession.TotalChunks.Valid {
			if err := s.updateTotalChunks(form.UploadId, form.TotalChunks); err != nil {
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"status": "failed to set total_chunks",
				})
				return
			}
		}

		if uploadSession.Status == "completed" {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "upload already completed",
			})
			return
		}

		if uploadSession.Status == "pending" {
			if err := s.updateUploadStatus(form.UploadId, "uploading"); err != nil {
				log.Println("updating uploaded status: ", err)
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"status": "something wrong happens",
				})
				return
			}
		}

		dstPath := filepath.Join("uploads", form.UploadId, fmt.Sprintf("%d.part", form.ChunkIndex))
		dst, err := os.Create(dstPath)
		if err != nil {
			s.failUpload(form.UploadId, "cannot write chunk: "+err.Error())
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "cannot write chunk",
			})
			return
		}
		defer dst.Close()

		io.Copy(dst, form.Chunk)

		if err := s.incrementUploadedChunks(form.UploadId); err != nil {
			log.Println("incrementing uploaded: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": "something wrong happens",
			})
			return
		}

		ctx.Response.WriteHeader(http.StatusOK)
	})
}

func (s *Server) completeUploadChunkHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *CompleteUploadForm) {

		// 1️⃣ Validate extension
		ext := strings.ToLower(filepath.Ext(form.Filename))
		if ext != ".csv" && ext != ".txt" {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "only CSV or TXT allowed",
			})
			return
		}

		uploadSession, err := s.findUploadSession(form.UploadID, int64(ctx.User().Id))
		if err != nil || uploadSession == nil {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": "invalid upload id",
			})
			return
		}

		if uploadSession.UploadedChunks != int(uploadSession.TotalChunks.Int64) {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": "missing chunks",
			})
			return
		}

		parts, _ := filepath.Glob(filepath.Join("uploads", form.UploadID, "*.part"))
		if len(parts) == 0 {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": "no chunks uploaded",
			})
			return
		}
		finalPath := filepath.Join("uploads", form.UploadID, form.Filename)
		out, _ := os.Create(finalPath)
		defer out.Close()

		sort.Strings(parts)

		for _, part := range parts {
			f, _ := os.Open(part)
			io.Copy(out, f)
			f.Close()
		}

		if err := s.updateUploadStatus(form.UploadID, "completed"); err != nil {
			log.Println("mutating uploaded status: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": "something wrong happens",
			})
			return
		}

		ctx.JSON(http.StatusOK, nil)
	})
}
