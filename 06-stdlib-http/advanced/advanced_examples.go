package advanced

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ParseFormParams demonstrates how to parse URL query parameters and form data.
func ParseFormParams() {
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		// Parse URL query parameters.
		name := r.URL.Query().Get("name")
		age := r.URL.Query().Get("age")

		fmt.Fprintf(w, "Query parameters received: name=%s, age=%s\n", name, age)

		// Parse form data from POST, PUT, and PATCH requests.
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "Error parsing form: %v", err)
				return
			}
			message := r.FormValue("message")
			fmt.Fprintf(w, "Form data received: message=%s\n", message)
		}
	})

	fmt.Println("Starting a server for form parsing on http://localhost:8083")
	fmt.Println("Try: curl \"http://localhost:8083/query?name=john&age=30\"")
	fmt.Println("Or: curl -X POST -d \"message=hello\" http://localhost:8083/query")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// FileUpload demonstrates how to handle file uploads.
func FileUpload() {
	// The upload form.
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
				<form action="/upload" method="post" enctype="multipart/form-data">
					<label for="file">Upload File:</label>
					<input type="file" name="file" id="file">
					<input type="submit" value="Upload">
				</form>
			`)
			return
		}

		if r.Method == http.MethodPost {
			// Parse the multipart form, with a max memory of 10MB.
			r.ParseMultipartForm(10 << 20)

			// Get the file from the form data.
			file, handler, err := r.FormFile("file")
			if err != nil {
				fmt.Println("Error retrieving the file:", err)
				return
			}
			defer file.Close()

			fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			fmt.Printf("File Size: %+v\n", handler.Size)
			fmt.Printf("MIME Header: %+v\n", handler.Header)

			// Create a temporary file to store the uploaded content.
			dst, err := os.Create(handler.Filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy the uploaded file to the destination file.
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Successfully Uploaded File\n")
		}
	})

	fmt.Println("Starting a server for file uploads on http://localhost:8084/upload")
	if err := http.ListenAndServe(":8084", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// FileServer demonstrates how to serve static files from a directory.
func FileServer() {
	// Create a directory and a file to serve.
	os.Mkdir("static", 0755)
	file, _ := os.Create("static/index.html")
	file.WriteString("<h1>Hello from a static file!</h1>")
	file.Close()

	// Create a file server handler.
	fs := http.FileServer(http.Dir("./static"))

	// Register the handler for a path prefix.
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	fmt.Println("Starting a file server on http://localhost:8085/files/")
	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
