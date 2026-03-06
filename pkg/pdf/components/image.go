package components

import (
	"bytes"
	"errors"

	"codeberg.org/go-pdf/fpdf"
)

var (
	ErrNilPDF         = errors.New("PDF instance is nil")
	ErrEmptyImageData = errors.New("image data is empty")
)

// ImageProps defines the properties for rendering images on a PDF.
type ImageProps struct {
	Source string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Alt    string
}

// ImageRenderer handles image rendering on PDF documents.
type ImageRenderer struct {
	pdf *fpdf.Fpdf
}

// NewImageRenderer creates a new ImageRenderer instance.
func NewImageRenderer(pdf *fpdf.Fpdf) *ImageRenderer {
	return &ImageRenderer{
		pdf: pdf,
	}
}

// Render draws an image on the PDF with the specified properties.
func (ir *ImageRenderer) Render(props ImageProps, data []byte) error {
	if ir.pdf == nil {
		return ErrNilPDF
	}

	if len(data) == 0 {
		return ErrEmptyImageData
	}

	// Register image from bytes
	imageReader := bytes.NewReader(data)

	_ = ir.pdf.RegisterImageOptionsReader(
		props.Source,
		fpdf.ImageOptions{ImageType: "PNG"},
		imageReader,
	)

	// Draw image on PDF
	ir.pdf.ImageOptions(
		props.Source,
		props.X,
		props.Y,
		props.Width,
		props.Height,
		false,
		fpdf.ImageOptions{ImageType: "PNG"},
		0,
		"",
	)

	return nil
}

// GetHeight returns the height of the image.
func (ir *ImageRenderer) GetHeight(props ImageProps) float64 {
	return props.Height
}
