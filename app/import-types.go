package app

import (
	"database/sql"
	"mime/multipart"
	"sync"
	"time"

	"github.com/martin3zra/forge/support"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

type UploadSessionForm struct {
	support.FormRequest
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Mime      string `json:"mime"`
	Delimiter string `json:"delimiter"`
	Encoding  string `json:"encoding"`
}

func (UploadSessionForm) Rules() map[string]any {
	return map[string]any{
		"mime":      "required",
		"filename":  "required",
		"size":      "required",
		"delimiter": "required",
		"encoding":  "required",
	}
}

type UploadChunkForm struct {
	support.FormRequest
	UploadId    string         `json:"upload_id"`
	ChunkIndex  int            `json:"chunk_index"`
	TotalChunks int            `json:"total_chunks"`
	Filename    string         `json:"filename"`
	Chunk       multipart.File `json:"chunk"`
}

func (UploadChunkForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		// "chunk_index":  "required|min:0",
		"total_chunks": "required",
		"filename":     "required",
		// "chunk":        "required",
	}
}

type CompleteUploadForm struct {
	support.FormRequest
	UploadID string `json:"upload_id"`
	Filename string `json:"filename"`
}

func (CompleteUploadForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		"filename":  "required",
	}
}

type UploadSession struct {
	ID        string `json:"id"`
	UserID    int64  `json:"user_id"`
	Filename  string `json:"filename"`
	FileSize  int64  `json:"file_size"`
	Delimiter string `json:"delimiter"`
	Encoding  string `json:"encoding"`
	// Type           string         `json:"records_type"`
	Status         string         `json:"status"`
	TotalChunks    sql.NullInt64  `json:"total_chunks"`
	UploadedChunks int            `json:"uploaded_chunks"`
	ErrorMessage   sql.NullString `json:"error_message"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ImportForm struct {
	support.FormRequest
	UploadID string `json:"upload_id"`
	Type     string `json:"type"`
}

func (ImportForm) Rules() map[string]any {
	return map[string]any{
		"upload_id": "required",
		"type":      "required|in:items,customers,vendors",
	}
}

type ImportEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

var importStreams = sync.Map{} // importID → chan ImportEvent

type ImportOptions struct {
	Delimiter rune
}

type UploadEncoding string

const (
	EncodingUTF8    UploadEncoding = "utf-8"
	EncodingLatin1  UploadEncoding = "latin-1"
	EncodingWin1252 UploadEncoding = "windows-1252"
)

var encodingDecoders = map[UploadEncoding]*encoding.Decoder{
	EncodingUTF8:    nil, // no-op
	EncodingLatin1:  charmap.ISO8859_1.NewDecoder(),
	EncodingWin1252: charmap.Windows1252.NewDecoder(),
}
