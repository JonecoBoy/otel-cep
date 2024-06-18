package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel-cep/tempByCep/pkg/external"
	"github.com/JonecoBoy/otel-cep/tempByCep/pkg/infra/telemetry"
	"github.com/JonecoBoy/otel-cep/tempByCep/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdown, err := telemetry.SetupProvider(ctx, "tempByCep")
	if err != nil {
		return
	}

	srv := &http.Server{
		Addr:         ":8090",
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

	handleFunc("/cep/", cepHandler)
	handleFunc("/temp/", tempHandler)
	//handler := otelhttp.NewHandler(mux, "/")
	return mux
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	log.Print(carrier)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	t := otel.Tracer("cep")
	ctx, span := t.Start(ctx, "cepHandler")
	defer span.End()

	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	cep := path[2]
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	c, err := CepConcurrency(ctx, cep)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusGatewayTimeout) // 504
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
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	t := otel.Tracer("temp")
	ctx, span := t.Start(ctx, "tempHandler")
	defer span.End()
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	cep := path[2]
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	c, err := CepConcurrency(ctx, cep)
	if err != nil {
		if err.Error() == "404 can not find zipcode" {
			w.WriteHeader(http.StatusNotFound) // 422
		}
		if err.Error() == "can not find zipcode" {
			w.WriteHeader(http.StatusUnprocessableEntity) // 404
		}
		w.WriteHeader(http.StatusGatewayTimeout) // 504
		w.Write([]byte(err.Error()))
		return
	}

	q := strings.Join([]string{utils.RemoveAccents(c.City), utils.RemoveAccents(c.State), "brazil"}, "-")

	temp, err := external.CurrentWeather(ctx, q, "pt")
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

func CepConcurrency(ctx context.Context, cep string) (external.Address, error) {
	ctx, internalSpan := otel.GetTracerProvider().Tracer("cep").Start(ctx, "concurrency-cep")
	defer internalSpan.End()
	c1 := make(chan Result)
	c2 := make(chan Result)

	go func() {
		data, err := external.BrasilApiCep(ctx, cep)
		c1 <- Result{Address: data, Err: err}
	}()
	go func() {
		data, err := external.ViaCep(ctx, cep)
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
	case <-time.After(time.Second * 30):
		return external.Address{}, errors.New("Timeout Reached, no API returned in time. CEP: " + cep)
	}
}
