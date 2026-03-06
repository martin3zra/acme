package components

import (
	"bytes"
	"errors"
	"io"

	"codeberg.org/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
)

// QRProps defines the properties for rendering QR codes on a PDF.
type QRProps struct {
	Content string
	X       float64
	Y       float64
	Size    float64
	Color   RGBColor
}

// BytesReader implements io.Reader interface for byte slices.
type BytesReader struct {
	data   []byte
	offset int
}

// Read implements the io.Reader interface.
func (br *BytesReader) Read(p []byte) (n int, err error) {
	if br.offset >= len(br.data) {
		return 0, io.EOF
	}
	n = copy(p, br.data[br.offset:])
	br.offset += n
	return n, nil
}

// QRRenderer handles QR code rendering on PDF documents.
type QRRenderer struct {
	pdf *fpdf.Fpdf
}

// NewQRRenderer creates a new QRRenderer instance.
func NewQRRenderer(pdf *fpdf.Fpdf) *QRRenderer {
	return &QRRenderer{
		pdf: pdf,
	}
}

// Render generates and draws a QR code on the PDF.
func (qr *QRRenderer) Render(props QRProps) error {
	if qr.pdf == nil {
		return errors.New("ErrNilPDF")
	}

	if props.Content == "" {
		return errors.New("ErrEmptyQRContent")
	}

	// Generate QR code as PNG in memory
	qrImage, err := qrcode.New(props.Content, qrcode.Medium)
	if err != nil {
		return err
	}

	pngData, err := qrImage.PNG(256)
	if err != nil {
		return err
	}

	// Register image from bytes
	imageReader := bytes.NewReader(pngData)

	_ = qr.pdf.RegisterImageOptionsReader(
		props.Content,
		fpdf.ImageOptions{ImageType: "PNG"},
		imageReader,
	)

	// Draw QR code image on PDF
	qr.pdf.ImageOptions(
		props.Content,
		props.X,
		props.Y,
		props.Size,
		props.Size,
		false,
		fpdf.ImageOptions{ImageType: "PNG"},
		0,
		"",
	)

	return nil
}

// GetHeight returns the height of the QR code.
func (qr *QRRenderer) GetHeight(props QRProps) float64 {
	return props.Size
}
