package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	First string
	Last  string
}

func main() {
	postUser()
	multipartPost()
}

func postUser() {
	// Create a user
	u := User{First: "Adam", Last: "Woodbeck"}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&u)
	if err != nil {
		fmt.Println("Error encoding user:", err)
		return
	}

	// Make a POST request with JSON content
	resp, err := http.Post("https://httpbin.org/post", "application/json", buf)
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and display the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Printf("Response:\n%s\n", body)
}

func multipartPost() {
	// Create a multipart form request body
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)

	// Add form fields
	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form fields with attached files",
	} {
		err := w.WriteField(k, v)
		if err != nil {
			fmt.Println("Error writing field:", err)
			return
		}
	}

	// Add files to the request
	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1), filepath.Base(file))
		if err != nil {
			fmt.Println("Error creating form file:", err)
			return
		}

		f, err := os.Open(file)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			fmt.Println("Error copying file to request:", err)
			return
		}
	}

	// Close the writer
	err := w.Close()
	if err != nil {
		fmt.Println("Error closing writer:", err)
		return
	}

	// Set a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", reqBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	// Perform the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Expected status %d; actual status %d\n", http.StatusOK, resp.StatusCode)
	} else {
		fmt.Printf("Response:\n%s\n", body)
	}
}
