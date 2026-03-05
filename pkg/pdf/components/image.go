package components

import (
	"bytes"
	"io"

	"codeberg.org/go-pdf/fpdf"
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

	err := ir.pdf.RegisterImageOptionsReader(
		props.Source,
		fpdf.ImageOptions{ImageType: "PNG"},
		imageReader,
	)
	if err != nil {
		return err
	}

	// Draw image on PDF
	ir.pdf.ImageOptions(
		props.Source,
		props.X,
		props.Y,
		props.Width,
		props.Height,
		false,
		fpdf.ImageOptions{ImageType: "PNG"},
	)

	return nil
}

// GetHeight returns the height of the image.
func (ir *ImageRenderer) GetHeight(props ImageProps) float64 {
	return props.Height
}















































}	return heightfunc (ir *ImageRenderer) GetHeight(pdf *fpdf.Fpdf, width, height float64) float64 {// GetHeight returns the height of the image}	return pdf.Error()	}, 0, "")		ImageType: "PNG",	pdf.ImageOptions(props.Source, props.X, props.Y, props.Width, props.Height, false, fpdf.ImageOptions{	// Draw image	}, imageReader)		ImageType: "PNG",	pdf.RegisterImageOptionsReader(props.Source, fpdf.ImageOptions{	// In production, would detect format (PNG, JPG, etc.)	// For now, using a simple image registration	imageReader := bytes.NewReader(data)	// Register and draw image from bytes	}		return nil	if len(data) == 0 {func (ir *ImageRenderer) Render(pdf *fpdf.Fpdf, props ImageProps, data []byte) error {// Render renders an image element on the PDF}	return &ImageRenderer{}func NewImageRenderer() *ImageRenderer {// NewImageRenderer creates a new image renderer}	Alt    string	Height float64	Width  float64	Y      float64	X      float64	Source stringtype ImageProps struct {// ImageProps represents image element propertiestype ImageRenderer struct{}// ImageRenderer handles rendering image elements)	"codeberg.org/go-pdf/fpdf"	"bytes"import (