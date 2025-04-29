package utils

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"math"

	"github.com/disintegration/imaging"
)

var (
	ProtanopiaMatrix    = [3][3]float64{{0.56667, 0.43333, 0}, {0.55833, 0.44167, 0}, {0, 0.24167, 0.75833}}
	DeuteranopiaMatrix  = [3][3]float64{{0.625, 0.375, 0}, {0.7, 0.3, 0}, {0, 0.3, 0.7}}
	TritanopiaMatrix    = [3][3]float64{{0.95, 0.05, 0}, {0, 0.43333, 0.56667}, {0, 0.475, 0.525}}
	ProtanomalyMatrix   = [3][3]float64{{0.816, 0.184, 0}, {0.333, 0.667, 0}, {0, 0.125, 0.875}}
	DeuteranomalyMatrix = [3][3]float64{{0.8, 0.2, 0}, {0.258, 0.742, 0}, {0, 0.142, 0.858}}
	TritanomalyMatrix   = [3][3]float64{{0.967, 0.033, 0}, {0, 0.733, 0.267}, {0, 0.183, 0.817}}
	AchromatopsiaMatrix = [3][3]float64{{0.299, 0.587, 0.114}, {0.299, 0.587, 0.114}, {0.299, 0.587, 0.114}}
	MonochromacyMatrix  = [3][3]float64{{0.33, 0.33, 0.33}, {0.33, 0.33, 0.33}, {0.33, 0.33, 0.33}}
)

// Box blur matrix
var boxBlurMatrix = [3][3]float64{
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
}

// Gaussian blur matrix
var gaussianBlurMatrix = [3][3]float64{
	{0.0625, 0.125, 0.0625},
	{0.125, 0.25, 0.125},
	{0.0625, 0.125, 0.0625},
}

// Sobel operators for edge detection
var (
	sobelX = [3][3]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	sobelY = [3][3]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}
)

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

// FlipImage flips the image upside down
func FlipImage(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			out.Set(x, bounds.Max.Y-y-1, img.At(x, y))
		}
	}
	return out
}

// RotateImage rotates the image by the given angle
func RotateImage(img image.Image, angle float64) image.Image {
	return imaging.Rotate(img, angle, color.Transparent)
}

// RotateImageWithShear rotates the image using three shear matrices
func RotateImageWithShear(img image.Image, angle float64) image.Image {
	// Convert angle to radians
	theta := angle * math.Pi / 180.0

	// Calculate shear matrices
	alpha := -math.Tan(theta / 2)
	beta := math.Sin(theta)
	gamma := -math.Tan(theta / 2)

	// Apply shears
	img = applyShear(img, alpha, 0) // First shear
	img = applyShear(img, beta, 1)  // Second shear
	img = applyShear(img, gamma, 0) // Third shear

	return img
}

func applyShear(img image.Image, factor float64, axis int) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var newX, newY int
			if axis == 0 {
				newX = x + int(float64(y)*factor)
				newY = y
			} else {
				newX = x
				newY = y + int(float64(x)*factor)
			}

			if newX >= bounds.Min.X && newX < bounds.Max.X &&
				newY >= bounds.Min.Y && newY < bounds.Max.Y {
				out.Set(newX, newY, img.At(x, y))
			}
		}
	}
	return out
}

// ConvertToGrayscale converts the image to grayscale
func ConvertToGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			gray := uint8(0.299*float64(originalColor.R) +
				0.587*float64(originalColor.G) +
				0.114*float64(originalColor.B))
			out.Set(x, y, color.RGBA{gray, gray, gray, originalColor.A})
		}
	}
	return out
}

// ApplyBoxBlur applies a box blur filter to the image
func ApplyBoxBlur(img image.Image) image.Image {
	return applyConvolution(img, boxBlurMatrix)
}

// ApplyGaussianBlur applies a Gaussian blur filter to the image
func ApplyGaussianBlur(img image.Image) image.Image {
	return applyConvolution(img, gaussianBlurMatrix)
}

// ApplyEdgeDetection applies Sobel edge detection to the image
func ApplyEdgeDetection(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	// Convert to grayscale first
	grayImg := ConvertToGrayscale(img)

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			var gx, gy float64

			// Apply Sobel operators
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := color.RGBAModel.Convert(grayImg.At(x+kx, y+ky)).(color.RGBA)
					gray := float64(pixel.R)
					gx += gray * sobelX[ky+1][kx+1]
					gy += gray * sobelY[ky+1][kx+1]
				}
			}

			// Calculate gradient magnitude
			magnitude := math.Sqrt(gx*gx + gy*gy)
			magnitude = clamp(magnitude)

			out.Set(x, y, color.RGBA{uint8(magnitude), uint8(magnitude), uint8(magnitude), 255})
		}
	}
	return out
}

func applyConvolution(img image.Image, kernel [3][3]float64) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			var r, g, b float64

			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := color.RGBAModel.Convert(img.At(x+kx, y+ky)).(color.RGBA)
					r += float64(pixel.R) * kernel[ky+1][kx+1]
					g += float64(pixel.G) * kernel[ky+1][kx+1]
					b += float64(pixel.B) * kernel[ky+1][kx+1]
				}
			}

			out.Set(x, y, color.RGBA{
				uint8(clamp(r)),
				uint8(clamp(g)),
				uint8(clamp(b)),
				255,
			})
		}
	}
	return out
}

// SimulateColorBlindness applies color blindness simulation
func SimulateColorBlindness(img image.Image, matrix [3][3]float64) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			r, g, b := float64(originalColor.R), float64(originalColor.G), float64(originalColor.B)

			newR := clamp(r*matrix[0][0] + g*matrix[0][1] + b*matrix[0][2])
			newG := clamp(r*matrix[1][0] + g*matrix[1][1] + b*matrix[1][2])
			newB := clamp(r*matrix[2][0] + g*matrix[2][1] + b*matrix[2][2])

			out.Set(x, y, color.RGBA{uint8(newR), uint8(newG), uint8(newB), originalColor.A})
		}
	}
	return out
}

// Daltonize applies daltonization to the image
func Daltonize(img image.Image, cbMatrix [3][3]float64) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			r, g, b := float64(originalColor.R), float64(originalColor.G), float64(originalColor.B)

			simR := clamp(r*cbMatrix[0][0] + g*cbMatrix[0][1] + b*cbMatrix[0][2])
			simG := clamp(r*cbMatrix[1][0] + g*cbMatrix[1][1] + b*cbMatrix[1][2])
			simB := clamp(r*cbMatrix[2][0] + g*cbMatrix[2][1] + b*cbMatrix[2][2])

			errR := r - simR
			errG := g - simG
			errB := b - simB

			newR := clamp(r + errR*0.6)
			newG := clamp(g + errG*0.6)
			newB := clamp(b + errB*0.6)

			out.Set(x, y, color.RGBA{uint8(newR), uint8(newG), uint8(newB), originalColor.A})
		}
	}
	return out
}

// Helper to decode an image from an io.Reader
func DecodeImage(r io.Reader) (image.Image, error) {
	return imaging.Decode(r)
}

// Helper to encode an image to JPEG in a buffer
func EncodeToJPEG(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := imaging.Encode(buf, img, imaging.JPEG)
	return buf.Bytes(), err
}
