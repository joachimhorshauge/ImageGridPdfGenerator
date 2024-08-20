# Visual Bingo Sheet Generator

This Go script generates customizable visual bingo sheets from a collection of images. Each image represents an item to be spotted on your bingo sheet.

## Features
- **Custom Grid Layout**: Generates bingo sheets with a customizable grid layout (e.g., 5x5).
- **Randomized Images**: Images are randomly placed in the grid to ensure variety across sheets.
- **Optional Overlay**: Add a white square with a black border on the bottom-right corner of each image (useful for branding or identification).

## Usage

### Basic Usage
To generate a PDF with 10 bingo sheets using images from a specified folder:

```bash
go run main.go ./images 10 output.pdf
```

### With Overlay

To add a white square with a black border to the bottom-right corner of each image:

```bash
go run main.go --overlay ./images 10 output.pdf
```

## Requirements

- Go: Ensure you have Go installed on your machine.

Installation

Clone the repository and navigate to the project directory:
```bash
git clone https://github.com/yourusername/visual-bingo-generator.git
cd visual-bingo-generator
```
Customization
- Grid Size: Modify gridRows and gridCols in the script to adjust the grid layout.
- Image Size: Adjust imgSize to control the size of each image in the grid.
