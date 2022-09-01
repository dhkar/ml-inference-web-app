package inference

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"os"

	tf "github.com/galeone/tensorflow/tensorflow/go"
	tg "github.com/galeone/tfgo"
	"github.com/nfnt/resize"
)

func ImportImage(filepath string) (img_raw image.Image) {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println("File Cant Be Accessed")
	}
	defer f.Close()

	img_raw, _, err = image.Decode(f)
	if err != nil {
		fmt.Println("wadw")

	}

	return img_raw
}

// See Library for More Interp Method: 	"github.com/nfnt/resize"
func ResizeUsingInterp(img_raw image.Image, method resize.InterpolationFunction, img_width int, img_height int) (img image.Image) {
	img = resize.Resize(uint(img_width), uint(img_height), img_raw, method)
	return
}

func ConvertImagetoTF(img image.Image) (imgTensor *tf.Tensor, input_dimension []int64) {

	var rf float32
	var gf float32
	var bf float32

	// Convert image.Image value into array 3D float32 (for tensorflow keras model input spec)
	imgArr := make([][][3]float32, img.Bounds().Size().Y)
	for y := 0; y < len(imgArr); y++ {
		imgArr[y] = make([][3]float32, img.Bounds().Size().X)
		for x := 0; x < len(imgArr[y]); x++ {
			px := x + img.Bounds().Min.X
			py := y + img.Bounds().Min.Y
			r, g, b, _ := img.At(px, py).RGBA()

			// Normalize into 0-1
			rf = float32(r>>8) / float32(255)
			gf = float32(g>>8) / float32(255)
			bf = float32(b>>8) / float32(255)

			// image.Image uses 16-bits for each color.
			// We want 8-bits.
			imgArr[y][x][0] = bf
			imgArr[y][x][1] = gf
			imgArr[y][x][2] = rf
		}
	}

	// Tensorflow input image dimension (array 4D)
	input_dimension = make([]int64, 4)
	input_dimension[0] = 1
	input_dimension[1] = int64(img.Bounds().Size().X)
	input_dimension[2] = int64(img.Bounds().Size().Y)
	input_dimension[3] = 3 // RGB

	imgTensor, err := tf.NewTensor(imgArr)
	if err != nil {
		fmt.Println("Failed to Convert Array To Tensor")
	}

	return
}

func InferenceImage(modelpath string, imgpath string) (prediction *tf.Tensor) {

	// Image Preprocessing for TF Model Input
	img_raw := ImportImage(imgpath)
	img := ResizeUsingInterp(img_raw, resize.Bilinear, 227, 227)
	imgTensor, input_dimensionTF := ConvertImagetoTF(img)
	err := imgTensor.Reshape(input_dimensionTF)
	if err != nil {
		fmt.Println("Failed to Reshape")
	}

	// Load model, Inference
	model := tg.LoadModel(modelpath, []string{"serve"}, nil)

	results := model.Exec([]tf.Output{
		model.Op("StatefulPartitionedCall", 0),
	}, map[tf.Output]*tf.Tensor{
		model.Op("serving_default_conv2d_15_input", 0): imgTensor,
	})

	// Prediction -> *tf.Tensor. use prediction.Value() and fmt.Println to get Result
	prediction = results[0]

	return
}
