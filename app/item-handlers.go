package app

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/martin3zra/acme/pkg/cache"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
	"golang.org/x/text/transform"
)

func (s *Server) itemsHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "item") {
		return
	}

	itemType := ItemType(ctx.Query("itemType"))
	if err := itemType.Validate(); err != nil {
		itemType = "all"
	}

	selectedUUID := ctx.Query("id")
	selectedAction := ctx.Query("action")
	if selectedAction != "edit" && selectedAction != "view" && selectedAction != "trash" {
		selectedAction = "view"
	}

	items, err := s.findItems(ctx.Request.Context(), itemType)
	if err != nil {
		ctx.Error(err)
		return
	}
	mode := ctx.Query("mode")
	props := inertia.Props{
		"openState":             mode == "creating",
		"translations":          trans("items"),
		"items":                 items,
		"currentItemTypeFilter": itemType,
	}

	if selectedUUID != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:item:%s", selectedUUID)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (*item, error) {
			selectedItem, err := s.findItemByUUID(ctx.Request.Context(), selectedUUID)
			if err != nil {
				return nil, err
			}

			if selectedItem.ItemType == "product" {
				setup, err := s.findItemVariantSetup(ctx.Request.Context(), selectedItem.ID)
				if err != nil {
					return nil, err
				}
				selectedItem.VariantSetup = setup
			}

			return selectedItem, nil
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		props["item"] = data
		props["selectedAction"] = selectedAction
		props["openState"] = true
	}

	// only add units defer if not creating
	if mode != "creating" {
		props["units"] = inertia.Defer(func() (any, error) {
			return s.findUnits(ctx.Request.Context())
		}, "attributes")
		props["taxes"] = inertia.Defer(func() (any, error) {
			return s.findTaxes(ctx.Request.Context())
		}, "attributes")
		props["attributes"] = inertia.Defer(func() (any, error) {
			return s.findAttributesWithValues(ctx.Request.Context())
		}, "attributes")
		props["warehouses"] = inertia.Defer(func() (any, error) {
			return s.findWarehouses(ctx.Request.Context())
		}, "attributes")
	} else {
		units, err := s.findUnits(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		props["units"] = units
		taxes, err := s.findTaxes(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		props["taxes"] = taxes
		attributes, err := s.findAttributesWithValues(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		props["attributes"] = attributes
		warehouses, err := s.findWarehouses(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		props["warehouses"] = warehouses
	}

	ctx.Render("Items/Index", props)
}

func (s *Server) itemVariantSetupHandler(ctx *routing.Context) {
	itemUUID := ctx.Param("id")

	item, err := s.findItemByUUID(ctx.Request.Context(), itemUUID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if item.ItemType != "product" {
		ctx.JSON(http.StatusOK, map[string]any{
			"has_variants":                 false,
			"attribute_ids":                []int{},
			"selected_values_by_attribute": map[int][]int{},
			"variants":                     []any{},
		})
		return
	}

	setup, err := s.findItemVariantSetup(ctx.Request.Context(), item.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, setup)
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
		itemUUID := ctx.Param("id")
		item, err := s.findItemByUUID(ctx.Request.Context(), itemUUID)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		err = s.updateItem(ctx.Request.Context(), item.ID, form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:item:%s", itemUUID)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}

func (s *Server) deleteItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		itemUUID := ctx.Param("id")
		item, err := s.findItemByUUID(ctx.Request.Context(), itemUUID)
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		err = s.deleteItem(ctx.Request.Context(), item.ID)
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:item:%s", itemUUID)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.item"}))

		ctx.Redirect("/items")
	})
}

func (s *Server) changeStatusItemHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		itemUUID := ctx.Param("id")

		item, err := s.findItemByUUID(ctx.Request.Context(), itemUUID)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return

		}

		err = s.toggleItemStatus(ctx.Request.Context(), item)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.item"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:item:%s", itemUUID)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
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
				"status": s.trans("global.onlyCsvOrTxtAllowed"),
			})
			return
		}

		// 2️⃣ Generate ID
		uploadID := uuid.NewString()

		// 3️⃣ Create upload directory
		dir := filepath.Join("uploads", uploadID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Println("failed to create upload dir: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": s.trans("global.somethingWentWrong"),
			})
			return
		}

		if err := s.storeUploadSession(&UploadSession{
			ID:        uploadID,
			UserID:    int64(ctx.User().Id),
			Filename:  form.Filename,
			FileSize:  form.Size,
			Delimiter: form.Delimiter,
			Encoding:  form.Encoding,
			Status:    "pending",
		}); err != nil {
			log.Println("Something went wrong starting the upoload.: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": s.trans("global.somethingWentWrong"),
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
			log.Println("Invalid total_chunks value:", form.TotalChunks)
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": s.trans("global.invalid.totalChunkSize"),
			})
			return
		}

		uploadSession, err := s.findUploadSession(form.UploadId)
		if err != nil || uploadSession == nil {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": s.trans("global.invalid.uploadId"),
			})
			return
		}

		// 🔥 Set total_chunks ONCE
		if !uploadSession.TotalChunks.Valid {
			if err := s.updateTotalChunks(form.UploadId, form.TotalChunks); err != nil {
				log.Println("failed to set total_chunks:", err)
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"status": s.trans("global.somethingWentWrong"),
				})
				return
			}
		}

		if uploadSession.Status == "completed" {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": s.trans("global.invalid.uploadAlreadyCompleted"),
			})
			return
		}

		if uploadSession.Status == "pending" {
			if err := s.updateUploadStatus(form.UploadId, "uploading"); err != nil {
				log.Println("updating uploaded status: ", err)
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"status": s.trans("global.somethingWentWrong"),
				})
				return
			}
		}

		dstPath := filepath.Join("uploads", form.UploadId, fmt.Sprintf("%d.part", form.ChunkIndex))
		dst, err := os.Create(dstPath)
		if err != nil {
			log.Println("cannot write chunk: ", err)
			s.failUpload(form.UploadId, "cannot write chunk: "+err.Error())
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": s.trans("global.somethingWentWrong"),
			})
			return
		}
		defer dst.Close()

		io.Copy(dst, form.Chunk)

		if err := s.incrementUploadedChunks(form.UploadId); err != nil {
			log.Println("incrementing uploaded: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": s.trans("global.somethingWentWrong"),
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
				"status": s.trans("global.invalid.onlyCsvOrTxtAllowed"),
			})
			return
		}

		uploadSession, err := s.findUploadSession(form.UploadID)
		if err != nil || uploadSession == nil {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": s.trans("global.invalid.uploadId"),
			})
			return
		}

		if uploadSession.UploadedChunks != int(uploadSession.TotalChunks.Int64) {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"status": s.trans("global.invalid.missingChunks"),
			})
			return
		}

		parts, _ := filepath.Glob(filepath.Join("uploads", form.UploadID, "*.part"))
		if len(parts) == 0 {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"status": s.trans("global.invalid.notFoundChunks"),
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
				"status": s.trans("global.somethingWentWrong"),
			})
			return
		}

		ctx.JSON(http.StatusOK, nil)
	})
}

func (s *Server) startImportHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ImportForm) {

		importID := uuid.New()
		if err := s.storeImport(importID.String(), form); err != nil {
			log.Println("Error starting import: ", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": s.trans("global.somethingWentWrong"),
			})
			return
		}

		go s.processImport(CurrentCompany(ctx.Request.Context()).ID, importID.String(), form)

		ctx.JSON(http.StatusOK, map[string]any{
			"import_id": importID,
			"status":    "queued",
		})
	})
}

func (s *Server) processImport(companyID int, importID string, form *ImportForm) {
	defer func() {
		if r := recover(); r != nil {
			s.failImport(importID, fmt.Sprint(r))
		}
	}()

	uSess, err := s.findUploadSession(form.UploadID)
	if err != nil {
		s.failImport(importID, fmt.Sprint(err))
		return
	}

	s.markStarted(importID)

	filePath := resolveUploadPath(fmt.Sprintf("%s/%s", form.UploadID, uSess.Filename))

	delimiter := rune(uSess.Delimiter[0])
	emit(importID, ImportEvent{"phase", "reading_file"})
	encoding := normalizeGivengEncoding(uSess.Encoding)
	totalRows, err := countCSVRows(filePath, encoding, delimiter)
	if err != nil {
		s.failImport(importID, "Failed counting rows")
		return
	}

	s.updateTotalRows(importID, totalRows)

	// 2️⃣ Open fresh reader for import
	file, utf8Reader, err := openForImport(filePath, encoding)
	if err != nil {
		s.failImport(importID, "Failed opening file")
		return
	}
	defer file.Close()

	emit(importID, ImportEvent{"phase", "mapping_columns"})
	csvReader := newCSVReader(utf8Reader, delimiter)

	headers, err := csvReader.Read()
	if err != nil {
		log.Println("Invalid CSV header", err)
		s.failImport(importID, "Invalid CSV header")
		return
	}

	if len(headers) <= 1 {
		sampleLines, _ := readSampleLines(filePath, 20)
		detected := DetectDelimiter(sampleLines)

		if detected != rune(uSess.Delimiter[0]) {
			s.failImport(
				importID,
				fmt.Sprintf(
					"Delimiter %q produced 1 column. %q works better.",
					uSess.Delimiter,
					string(detected),
				),
			)
			return
		}
		// 🚨 Hard stop – delimiter probably wrong
		s.failImport(
			importID,
			fmt.Sprintf(
				"Delimiter %q produced only one column. Please choose the correct delimiter.",
				uSess.Delimiter,
			),
		)
		return
	}

	columnMap, err := mapHeaders(headers, form.Type)
	if err != nil {
		s.failImport(importID, err.Error())
		return
	}

	emit(importID, ImportEvent{"phase", "importing_rows"})

	if err := s.processRows(companyID, importID, form.Type, csvReader, columnMap, totalRows); err != nil {
		log.Println("Something wrong happens processing the records", err)
		s.failImport(importID, err.Error())
		emit(importID, ImportEvent{"type", "failed"})
		return
	}

	emit(importID, ImportEvent{
		Type: "progress",
		Data: map[string]int{
			"processed": totalRows,
			"total":     totalRows,
		},
	})

	importFile, err := s.findImportByID(importID)
	if err != nil {
		log.Println("Something wrong happens processing the records", err)
		s.failImport(importID, err.Error())
		emit(importID, ImportEvent{"type", "failed"})
	}

	s.completeImport(importFile)
}

func resolveUploadPath(uploadID string) string {
	// Example: uploads stored as /data/uploads/{upload_id}.csv
	return filepath.Join("uploads", uploadID)
}

func openForImport(path string, enc UploadEncoding) (*os.File, io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	if dec := encodingDecoders[enc]; dec != nil {
		reader := transform.NewReader(file, dec)
		return file, reader, nil
	}

	return file, file, nil
}

func normalizeGivengEncoding(enc string) UploadEncoding {
	e := strings.ToLower(strings.TrimSpace(enc))

	switch e {
	case "utf-8", "utf8":
		return EncodingUTF8

	case "latin-1", "latin1", "iso-8859-1":
		// Treat Latin-1 as Windows-1252 (REALITY)
		return EncodingWin1252

	case "windows-1252", "win1252", "cp1252":
		return EncodingWin1252
	}

	// Default fallback (safe)
	return EncodingUTF8
}

func newCSVReader(r io.Reader, delimiter rune) *csv.Reader {
	cr := csv.NewReader(r)
	cr.Comma = delimiter
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true
	return cr
}

func countCSVRows(path string, enc UploadEncoding, delimiter rune) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("error open file: ", err)
		return 0, err
	}
	defer file.Close()

	var r io.Reader = file
	if dec := encodingDecoders[enc]; dec != nil {
		r = transform.NewReader(file, dec)
	}

	reader := newCSVReader(r, delimiter)

	count := 0
	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// ⚠️ Do NOT fail — row counting should be forgiving
			log.Println("row read error:", err)
			continue
		}
		count++
	}

	// minus header
	if count > 0 {
		count--
	}

	return count, nil
}

func mapHeaders(headers []string, source string) (map[int]string, error) {
	var mapping map[string]string

	if source == "items" {
		mapping = map[string]string{
			"NOMBRE":      "name",
			"DESCRIPCION": "description",
			"PRECIO":      "price",
			"ITBIS":       "tax_rate",
			"TIPO":        "item_type",
			"SKU":         "sku",
			"CODIGO":      "code",
			"BARRA":       "barcode",
			"REFERENCIA":  "reference",
			"REF_SUP":     "vendor_reference",
		}
	}

	if source == "customers" {
		mapping = map[string]string{
			"NOMBRE":          "name",
			"NOMBRE_CONTACTO": "contact_name",
			"TELEFONO":        "phone",
			"CORREO":          "email",
			"PAGO":            "payment_method",
			"LIMITE":          "credit_limit",
			"CONDICIONES":     "payment_terms",
			"TIPO_NCFTP":      "tax_receipt_id",
			"CODIGO":          "code",
			"LIMITE_CRE":      "credit_limited",
		}
	}

	result := map[int]string{}

	for i, h := range headers {
		key := strings.ToUpper(strings.TrimSpace(h))
		if v, ok := mapping[key]; ok {
			result[i] = v
		}
	}

	if _, ok := result[0]; !ok {
		return nil, errors.New("Missing NAME column")
	}

	return result, nil
}

func (s *Server) importEventsHandler(w http.ResponseWriter, r *http.Request) {
	importID := strings.TrimPrefix(r.URL.Path, "/sse/imports/")
	if importID == "" {
		http.Error(w, "missing import id", http.StatusBadRequest)
		return
	}
	// Required SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // 🔥 Nginx hint

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan ImportEvent, 10)
	importStreams.Store(importID, ch)
	defer importStreams.Delete(importID)

	for {
		select {
		case ev := <-ch:
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Helper to emit events
func emit(importID string, event ImportEvent) {
	if ch, ok := importStreams.Load(importID); ok {
		ch.(chan ImportEvent) <- event
	}
}

func DetectDelimiter(lines []string) rune {
	candidates := []rune{',', ';', '\t', '|'}
	best := ','
	bestScore := -1

	for _, d := range candidates {
		counts := make([]int, 0, 10)
		for i := 0; i < len(lines) && i < 10; i++ {
			counts = append(counts, len(strings.Split(lines[i], string(d))))
		}

		score := 0
		if counts[0] > 1 {
			score += 2
		}

		for _, c := range counts {
			if c == counts[0] {
				score++
			}
		}

		if score > bestScore {
			bestScore = score
			best = d
		}
	}

	return best
}

func readSampleLines(filePath string, maxLines int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Increase buffer size in case of long CSV lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	lines := make([]string, 0, maxLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lines = append(lines, line)

		if len(lines) >= maxLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}
