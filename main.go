package main

import (
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/nfnt/resize"
)

const (
	gridRows = 5
	gridCols = 5
)

var (
	imgSize       = 50.0 // size of each image in the grid (in points, for PDF)
	marginTop     = 10.0 // top margin
	marginLeft    = 10.0 // left margin
	cellSpacing   = 2.0  // spacing between cells
	overlaySquare = flag.Bool("overlay", false, "Overlay a white square with a black border on the bottom right of each image")
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 3 {
		fmt.Println("Usage: go run main.go [--overlay] <image_folder_path> <number_of_pages> <output_pdf>")
		return
	}

	imageFolder := flag.Args()[0]
	numPages := atoi(flag.Args()[1])
	outputPDF := flag.Args()[2]

	log.Printf("Loading images from folder: %s", imageFolder)
	images, err := loadAndResizeImages(imageFolder)
	if err != nil {
		log.Fatalf("Failed to load images from folder: %v", err)
	}

	if len(images) == 0 {
		log.Fatalf("No images found in the specified folder.")
	}

	fmt.Printf("\nGenerating PDF with %d pages\n", numPages)
	generatePDF(images, numPages, outputPDF)
	fmt.Printf("\nGenerated %d pages\n", numPages) // Move to a new line after the last update
	log.Printf("PDF generated successfully: %s", outputPDF)
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Failed to convert string to int: %v", err)
	}
	return n
}

func loadAndResizeImages(folder string) ([][]byte, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	var images [][]byte
	var wg sync.WaitGroup
	imageChan := make(chan []byte, len(files))

	totalFiles := len(files)
	processedFiles := 0

	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			wg.Add(1)
			go func(file os.DirEntry) {
				defer wg.Done()
				imagePath := filepath.Join(folder, file.Name())
				imgData, err := resizeImage(imagePath)
				if err != nil {
					log.Printf("Failed to process image %s: %v", imagePath, err)
					return
				}
				imageChan <- imgData
				processedFiles++
				fmt.Printf("\rLoaded and resized %d/%d images", processedFiles, totalFiles)
			}(file)
		}
	}

	go func() {
		wg.Wait()
		close(imageChan)
	}()

	for imgData := range imageChan {
		images = append(images, imgData)
	}

	fmt.Printf("\nLoaded and resized %d images\n", len(images)) // New line after all images are processed
	return images, nil
}

func isImageFile(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func resizeImage(imagePath string) ([]byte, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	cellSize := uint(imgSize)
	resizedImg := resize.Resize(cellSize, cellSize, img, resize.Lanczos3)

	if *overlaySquare {
		resizedImg = addOverlay(resizedImg)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addOverlay(img image.Image) image.Image {
	// Create a new image with the same dimensions as the resized image
	rgba := image.NewRGBA(img.Bounds())

	// Draw the original image onto the new RGBA image
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Define the size of the white square overlay
	overlaySize := int(0.2 * float64(img.Bounds().Dx())) // 20% of the image width

	// Define the position of the square (bottom-right corner)
	rect := image.Rect(rgba.Bounds().Dx()-overlaySize, rgba.Bounds().Dy()-overlaySize, rgba.Bounds().Dx(), rgba.Bounds().Dy())

	// Draw the white square
	white := image.NewUniform(color.White)
	draw.Draw(rgba, rect, white, image.Point{}, draw.Src)

	// Draw the black border
	black := color.RGBA{0, 0, 0, 255}
	for x := rect.Min.X; x < rect.Max.X; x++ {
		rgba.Set(x, rect.Min.Y, black)
		rgba.Set(x, rect.Max.Y-1, black)
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		rgba.Set(rect.Min.X, y, black)
		rgba.Set(rect.Max.X-1, y, black)
	}

	return rgba
}

func generatePDF(images [][]byte, numPages int, outputPDF string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pageWidth, _ := pdf.GetPageSize()

	// Calculate cell width and height to ensure cells are square
	cellSize := (pageWidth - 2*marginLeft - (gridCols-1)*cellSpacing) / gridCols

	for i := 0; i < numPages; i++ {
		pdf.AddPage()
		pdf.SetMargins(marginLeft, marginTop, marginLeft)

		// Shuffle images
		rand.Shuffle(len(images), func(i, j int) {
			images[i], images[j] = images[j], images[i]
		})

		// Add images to the grid
		for row := 0; row < gridRows; row++ {
			for col := 0; col < gridCols; col++ {
				x := marginLeft + float64(col)*(cellSize+cellSpacing)
				y := marginTop + float64(row)*(cellSize+cellSpacing)
				imgIndex := (i*gridRows*gridCols + row*gridCols + col) % len(images)
				addImageToPDF(pdf, images[imgIndex], x, y, cellSize, cellSize)
			}
		}
		fmt.Printf("\rGenerated page %d/%d", i+1, numPages)
	}

	err := pdf.OutputFileAndClose(outputPDF)
	if err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
	}
}

func addImageToPDF(pdf *gofpdf.Fpdf, imgData []byte, x, y, w, h float64) {
	imageName := fmt.Sprintf("img_%x", sha1.Sum(imgData)) // Generate a consistent name for the image based on its content
	if pdf.GetImageInfo(imageName) == nil {
		pdf.RegisterImageOptionsReader(imageName, gofpdf.ImageOptions{ImageType: "JPEG", ReadDpi: true}, bytes.NewReader(imgData))
	}
	pdf.ImageOptions(imageName, x, y, w, h, false, gofpdf.ImageOptions{ImageType: "JPEG", ReadDpi: true}, 0, "")
}
