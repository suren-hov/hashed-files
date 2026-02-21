package main

import (
    "fmt"
    "io"
    "os"
    "crypto/sha256"
    "net/http"
    "errors"
)

const STORAGE_PATH = "./storage"

func handleGet(w http.ResponseWriter, r *http.Request) {
    hash := r.PathValue("hash")
    filePath := fmt.Sprintf("%s/%s", STORAGE_PATH, hash)
    file, err := os.Open(filePath) 

    if err != nil {
		if errors.Is(err, os.ErrNotExist) {
            http.Error(w, "File not exists", http.StatusBadRequest)
            return
        }
        panic(err)
    }
    defer file.Close()

    _, err = io.Copy(w, file) 
    if err != nil {
        http.Error(w, "Can't send the file", http.StatusBadRequest)
        panic(err)
    }
}

func handleSave(w http.ResponseWriter, r *http.Request) {
    
    tempFile, err :=  os.CreateTemp(STORAGE_PATH, "temp-data-*")
    if err != nil {
        panic(err)
    }

    defer tempFile.Close()
    defer os.Remove(tempFile.Name())
    fmt.Printf("Created tempfile: %s\n", tempFile.Name())

    hasher := sha256.New()
    multiWriter := io.MultiWriter(tempFile, hasher)

    if _, err := io.Copy(multiWriter, r.Body); err != nil {
        panic(err) 
    }

    hash := hasher.Sum(nil)
    hashStr := fmt.Sprintf("%x", hash)[:16]
    
    filePath := fmt.Sprintf("%s/%s", STORAGE_PATH, hashStr)
    err = os.Rename(tempFile.Name(), filePath) 

    if err != nil {
        panic(err)
    }

    fmt.Fprintf(w, hashStr)
}

func main() {
    fmt.Println("SuperDB started")

    err := os.MkdirAll(STORAGE_PATH, 0755)

    if err != nil {
        fmt.Println("Failed to create storage folder")
        panic(err)
    }

    mux := http.NewServeMux()

    mux.HandleFunc("GET /{hash}", handleGet)
    mux.HandleFunc("POST /", handleSave)

    http.ListenAndServe(":8080", mux)
}
