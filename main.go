package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ClientData represents the structure of form data received from the client
type ClientData struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Telephone         string `json:"telephone"`
	Ecole             string `json:"ecole"`
	Filiere           string `json:"filiere"`
	Ville             string `json:"ville"`
	Formations        string `json:"formations"`
	ContactPreference string `json:"contactPreference"`
}

var (
	csvFile   = "client_data.csv"
	fileMutex sync.Mutex
)

func main() {
	// Initialize CSV file with headers if it doesn't exist
	initCSVFile()

	// Handle CORS
	http.HandleFunc("/api/submit", enableCORS(handleFormSubmission))

	// Start the server
	port := ":8080"
	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func initCSVFile() {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	_, err := os.Stat(csvFile)
	if os.IsNotExist(err) {
		file, err := os.Create(csvFile)
		if err != nil {
			log.Fatalf("Failed to create CSV file: %v", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		headers := []string{
			"Timestamp",
			"Name",
			"Email",
			"Telephone",
			"Ecole",
			"Filiere",
			"Ville",
			"Formations",
			"ContactPreference",
		}

		if err := writer.Write(headers); err != nil {
			log.Fatalf("Error writing headers to CSV: %v", err)
		}
		writer.Flush()
	}
}

// enableCORS is a middleware that adds CORS headers to responses
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleFormSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var clientData ClientData
	if err := json.NewDecoder(r.Body).Decode(&clientData); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if clientData.Name == "" || clientData.Email == "" || clientData.Telephone == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := saveToCSV(clientData); err != nil {
		log.Printf("Error saving to CSV: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Data successfully saved"})
}

func saveToCSV(data ClientData) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile(csvFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{
		timestamp,
		data.Name,
		data.Email,
		data.Telephone,
		data.Ecole,
		data.Filiere,
		data.Ville,
		data.Formations,
		data.ContactPreference,
	}

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("error writing to CSV: %v", err)
	}

	return nil
}