package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/external"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type TempResponse struct {
	// Location *external.Location `json:"location"`
	City   string  `json:"city"`
	Temp_C float32 `json:"temp_c"`
	Temp_F float32 `json:"temp_f"`
	Temp_K float32 `json:"temp_k"`
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	mux := http.NewServeMux()

	mux.HandleFunc("/", tempHandler)

	log.Print("Listening...")
	http.ListenAndServe(":8091", mux)
}

func validateCep(cep string) error {
	cep = strings.ReplaceAll(cep, "-", "")
	if len(cep) != 8 {
		return errors.New("cep must contain exactly 8 characters")
	}

	_, err := strconv.Atoi(cep)
	if err != nil {
		return errors.New("cep must contain only numbers")
	}

	return nil
}
func tempHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var body map[string]string
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	// Get the cep from the body
	cep, ok := body["cep"]
	if !ok {
		http.Error(w, "cep not provided", http.StatusBadRequest)
		return
	}
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	err = validateCep(cep)
	if err != nil {
		w.WriteHeader(utils.InvalidZipError.Code) // 422
		w.Write([]byte(utils.InvalidZipError.Message))
		return
	}

	c, err := external.GetTempByCep(cep)
	if err != nil {
		if err.Error() == "404 can not find zipcode" {
			w.WriteHeader(http.StatusNotFound) // 422
		}
		if err.Error() == "can not find zipcode" {
			w.WriteHeader(http.StatusUnprocessableEntity) // 404
		}

		w.Write([]byte(err.Error()))
		return
	}

	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte(err.Error()))
		return

	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	tempResponse := TempResponse{
		//Location: temp.Location,
		City:   c.City,
		Temp_C: c.Temp_C,
		Temp_F: c.Temp_F,
		Temp_K: c.Temp_K,
	}
	jsonData, err := json.Marshal(tempResponse)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print(string(jsonData))
	w.Write(jsonData)
}
