package components

import (
	"bytes"
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
		return ErrNilPDF
	}

	if props.Content == "" {
		return ErrEmptyQRContent
	}

	// Generate QR code as PNG in memory
	qrImage, err := qrcode.New(props.Content, qrcode.Medium)
	if err != nil {
		return err
	}

	pngData := qrImage.PNG(256)

	// Register image from bytes
	imageReader := bytes.NewReader(pngData)

	err = qr.pdf.RegisterImageOptionsReader(
		props.Content,
		fpdf.ImageOptions{ImageType: "PNG"},
		imageReader,
	)
	if err != nil {
		return err
	}

	// Draw QR code image on PDF
	qr.pdf.ImageOptions(
		props.Content,
		props.X,
		props.Y,
		props.Size,
		props.Size,
		false,
		fpdf.ImageOptions{ImageType: "PNG"},
	)

	return nil
}

// GetHeight returns the height of the QR code.
func (qr *QRRenderer) GetHeight(props QRProps) float64 {
	return props.Size
}


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
		return ErrNilPDF
	}

	if props.Content == "" {
		return ErrEmptyQRContent
	}

	// Generate QR code as PNG in memory
	qrImage, err := qrcode.New(props.Content, qrcode.Medium)
	if err != nil {
		return err
	}

	pngData := qrImage.PNG(256)

	// Register image from bytes
	imageReader := bytes.NewReader(pngData)

	err = qr.pdf.RegisterImageOptionsReader(
		props.Content,
		fpdf.ImageOptions{ImageType: "PNG"},
		imageReader,
	)
	if err != nil {
		return err
	}

	// Draw QR code image on PDF
	qr.pdf.ImageOptions(
		props.Content,
		props.X,
		props.Y,
		props.Size,
		props.Size,
		false,
		fpdf.ImageOptions{ImageType: "PNG"},
	)

	return nil
}

// GetHeight returns the height of the QR code.
func (qr *QRRenderer) GetHeight(props QRProps) float64 {
	return props.Size
}

import (
	"codeberg.org/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
)

// QRRenderer handles rendering QR code elements
type QRRenderer struct{}
































































}	return n, nil	br.pos += n	n = copy(p, br.data[br.pos:])	}		return 0, nil // EOF	if br.pos >= len(br.data) {func (br *BytesReader) Read(p []byte) (n int, err error) {}	pos  int	data []bytetype BytesReader struct {// BytesReader is a simple reader for bytes}	return sizefunc (qr *QRRenderer) GetHeight(size float64) float64 {// GetHeight returns the height of the QR code}	return nil	}, 0, "")		ImageType: "PNG",	pdf.ImageOptions("qr"+props.Content, props.X, props.Y, props.Size, props.Size, false, fpdf.ImageOptions{	}, &BytesReader{data: pngData})		ImageType: "PNG",	pdf.RegisterImageOptionsReader("qr"+props.Content, fpdf.ImageOptions{	// Register and draw QR code	}		return err	if err != nil {	pngData, err := qrCode.PNG(int(props.Size * 2)) // 2x for better resolution	// Convert to PNG	}		return err	if err != nil {	qrCode, err := qrcode.New(props.Content, qrcode.High)	// Generate QR code	}		return nil	if props.Content == "" {func (qr *QRRenderer) Render(pdf *fpdf.Fpdf, props QRProps) error {// Render renders a QR code element on the PDF}	return &QRRenderer{}func NewQRRenderer() *QRRenderer {// NewQRRenderer creates a new QR code renderer}	Color   RGBColor	Size    float64 // in mm	Y       float64	X       float64	Content stringtype QRProps struct {// QRProps represents QR code element properties