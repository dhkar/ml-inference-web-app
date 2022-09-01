package main

import (
	"fmt"
	"html/template"
	"io"

	"ml-inference/inference"
	"net/http"
	"os"
	"path"
)

var ImgFilename string
var classResult string

func main() {

	// Route root/main (/)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var filepath = path.Join("assets", "index.html")
		var template, error = template.ParseFiles(filepath)

		if error != nil {
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}

		var data = map[string]interface{}{}

		template.Execute(w, data)
	})

	// Route to handle image submission
	http.HandleFunc("/submit_image", func(w http.ResponseWriter, r *http.Request) {

		// Parse image File from form to imgFile variable
		r.ParseMultipartForm(10 << 20)
		imgFile, handler, error := r.FormFile("imgFile")

		// Error Handler
		if error != nil {
			fmt.Println("File Cannot be Uploaded")
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}
		ImgFilename = handler.Filename
		defer imgFile.Close()

		// Create new empty file as image file template
		dir, _ := os.Getwd()
		imgpath := path.Join(dir, "assets", "imgs", ImgFilename)
		fmt.Println(imgpath)
		imgTargetFile, _ := os.OpenFile(imgpath, os.O_WRONLY|os.O_CREATE, 0666)
		defer imgTargetFile.Close()

		// Copy imgFile to image file template (imgTargetFile)
		_, errCopy := io.Copy(imgTargetFile, imgFile)
		if errCopy != nil {
			fmt.Println("File Cannot be Copied")
			http.Error(w, errCopy.Error(), http.StatusInternalServerError)
		}

		// Show Page *output.html
		var filepath = path.Join("assets", "output.html")
		var template, err2 = template.ParseFiles(filepath)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}

		imgpath = fmt.Sprintf("%s%s", "/assets/imgs/", ImgFilename)

		// Image Inference to Tensorflow Model
		imgpathRelative := path.Join("./assets", "imgs", ImgFilename)

		fmt.Println(imgpathRelative)
		modelpath := "assets/models/RPSClassificationModelv4"

		prediction := inference.InferenceImage(modelpath, imgpathRelative)
		pred := prediction.Value()

		// Extract Interface{} data type pred
		pred_arr := pred.([][]float32)

		// Determine index array that has max value
		max := pred_arr[0][0]
		idxmax := 0
		for idx, value := range pred_arr[0] {
			if value > max {
				max = value
				idxmax = idx
			}
		}

		// Convert idxmax to class
		switch idxmax {
		case 0:
			classResult = "paper"
		case 1:
			classResult = "rock"
		case 2:
			classResult = "scissors"
		default:
			classResult = "Cannot Determine"
		}

		// Pass classResilt Value to html
		var data = map[string]interface{}{
			"filename":    ImgFilename,
			"img_path":    imgpath,
			"classResult": classResult,
		}

		template.Execute(w, data)
	})
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.ListenAndServe(":7777", nil)

}
