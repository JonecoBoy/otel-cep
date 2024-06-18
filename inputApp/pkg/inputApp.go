package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/external"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/infra/telemetry"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type TempResponse struct {
	// Location *external.Location `json:"location"`
	City   string  `json:"city"`
	Temp_C float32 `json:"temp_c"`
	Temp_F float32 `json:"temp_f"`
	Temp_K float32 `json:"temp_k"`
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdown, err := telemetry.SetupProvider(ctx, "inputApp")
	if err != nil {
		return
	}

	srv := &http.Server{
		Addr:         ":8091",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mainHttpHanlder(),
	}
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return

}

func mainHttpHanlder() http.Handler {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	mux.Handle("/metrics", promhttp.Handler())
	handleFunc("/", tempHandler)
	//handler := otelhttp.NewHandler(mux, "/")
	return mux
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
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	t := otel.Tracer("weather")
	ctx, span := t.Start(ctx, "tempHandler")
	defer span.End()

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

	c, err := external.GetTempByCep(ctx, cep)
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
