package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
)

var ImgFilename string

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

		// Parse image File to imgFile
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

		dir, _ := os.Getwd()
		imgpath := path.Join(dir, "assets", "imgs", ImgFilename)
		fmt.Println(imgpath)
		imgTargetFile, _ := os.OpenFile(imgpath, os.O_WRONLY|os.O_CREATE, 0666)
		defer imgTargetFile.Close()

		_, errCopy := io.Copy(imgTargetFile, imgFile)
		if errCopy != nil {
			fmt.Println("File Cannot be Copied")
			http.Error(w, errCopy.Error(), http.StatusInternalServerError)
		}

		// Show Page *showOutput.html
		var filepath = path.Join("assets", "showOutput.html")
		var template, err2 = template.ParseFiles(filepath)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}

		imgpath = fmt.Sprintf("%s%s", "/assets/imgs/", ImgFilename)
		var data = map[string]interface{}{
			"filename": ImgFilename,
			"img_path": imgpath,
		}

		template.Execute(w, data)
	})
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.ListenAndServe(":7777", nil)

}
