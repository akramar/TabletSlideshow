package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

// Page object
type Page struct {
	Title string
	Body  []byte
}

func viewFrameHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view", nil)
}

func getImgHandler(w http.ResponseWriter, r *http.Request) {
	// List all the files
	//TODO: should only pull image file types, not everything, that's dangerous
	files, _ := ioutil.ReadDir("./uploads")
	//for _, f := range files {
	//	fmt.Println(f.Name())
	//}

	if len(files) == 0 || (len(files) == 1 && files[0].Name() == ".gitignore") {
		fmt.Fprintf(w, "No images available!")
		return
	}

	// Pick a random in the array

	rand.Seed(time.Now().Unix())

	var chosenFile os.FileInfo

	for {
		randIdx := rand.Intn(len(files))
		chosenFile = files[randIdx]
		fmt.Println("chosen file: " + chosenFile.Name())

		// Change when filtering file list above
		if chosenFile.Name() != ".gitignore" {
			break
		}
	}

	// file to byte stream?

	//fileBytes, _ := ioutil.ReadFile("./uploads/" + chosenFile.Name())
	//fmt.Println(fileBytes)

	infile, err := os.Open("./uploads/" + chosenFile.Name())
	if err != nil {
		// replace this with real error handling
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
	}

	// respond with bytes

	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, src, nil)
	if err != nil {
		log.Println("unable to encode image.")
	}

	imgStr := base64.StdEncoding.EncodeToString(buffer.Bytes())

	w.Header().Set("Content-Type", "image/jpeg")
	//w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	w.Header().Set("Content-Length", strconv.Itoa(len(imgStr)))

	_, err = w.Write([]byte(imgStr))
	//_, err = w.Write(buffer.Bytes())
	if err != nil {
		log.Println("unable to write image. " + err.Error())
	}

}

func saveImgHandler(w http.ResponseWriter, r *http.Request) {
	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	defer file.Close()

	filename := header.Filename

	out, err := os.Create("uploads/" + filename)
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege - "+
			err.Error())
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "upload", nil)
}

var templates = template.Must(template.ParseFiles("upload.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(upload|save|view|get-img|)/$")

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(viewFrameHandler))
	http.HandleFunc("/upload/", makeHandler(uploadPageHandler))
	http.HandleFunc("/save/", makeHandler(saveImgHandler))
	http.HandleFunc("/get-img/", makeHandler(getImgHandler))

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	//http.ListenAndServe(":8080", nil)
}
