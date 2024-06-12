package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JonecoBoy/tempByCep/pkg/external"
	"github.com/JonecoBoy/tempByCep/pkg/utils"
)

type Result struct {
	Address external.Address
	Err     error
}

type TempResponse struct {
	// Location *external.Location `json:"location"`
	City   string  `json:"city"`
	Temp_C float32 `json:"temp_c"`
	Temp_F float32 `json:"temp_f"`
	Temp_K float32 `json:"temp_k"`
}

func main() {
	// allow insecure connection
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	mux := http.NewServeMux()
	// podia ter passado anonima
	mux.HandleFunc("/cep/", cepHandler)
	mux.HandleFunc("/temp/", tempHandler)
	mux.HandleFunc("/", HomeHandler)

	log.Print("Listening...")
	http.ListenAndServe(":8090", mux)

}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	cep := path[2]
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	c, err := CepConcurrency(cep)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(string(jsonData))
	w.Write(jsonData)
}

func tempHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	cep := path[2]
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	c, err := CepConcurrency(cep)
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

	q := strings.Join([]string{utils.RemoveAccents(c.City), utils.RemoveAccents(c.State), "brazil"}, "-")

	temp, err := external.CurrentWeather(q, "pt")
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte(err.Error()))
		return

	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	TempResponse := TempResponse{
		//Location: temp.Location,
		City:   temp.Location.Name,
		Temp_C: temp.Current.TempC,
		Temp_F: temp.Current.TempF,
		Temp_K: temp.Current.TempC + 273,
	}
	jsonData, err := json.Marshal(TempResponse)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print(string(jsonData))
	w.Write(jsonData)
}

func CepConcurrency(cep string) (external.Address, error) {
	c1 := make(chan Result)
	c2 := make(chan Result)

	go func() {
		data, err := external.BrasilApiCep(cep)
		c1 <- Result{Address: data, Err: err}
	}()
	go func() {
		data, err := external.ViaCep(cep)
		c2 <- Result{Address: data, Err: err}
	}()

	select {
	case res := <-c1:
		if res.Err != nil {
			return external.Address{}, res.Err
		}
		return res.Address, nil
	case res := <-c2:
		if res.Err != nil {
			return external.Address{}, res.Err
		}
		return res.Address, nil
	case <-time.After(time.Second * 1):
		return external.Address{}, errors.New("Timeout Reached, no API returned in time. CEP: " + cep)
	}
}
