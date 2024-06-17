package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel-cep/tempByCep/pkg/external"
	"github.com/JonecoBoy/otel-cep/tempByCep/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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

	shutdown, err := initProvider("tempByCep", "otel-collector:4317")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	// allow insecure connection
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	mux := http.NewServeMux()
	// podia ter passado anonima
	mux.HandleFunc("/cep/", cepHandler)
	mux.HandleFunc("/temp/", tempHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", HomeHandler)
	log.Print("Listening...")
	http.ListenAndServe(":8090", mux)

}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// header pra propagar nas requests futuras
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	//span := trace.SpanFromContext(ctx)
	span := trace.SpanFromContext(ctx)
	//spanCtx := span.SpanContext()
	span.SetName("HomeHandler")

	defer span.End()
	//injetando o header
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	//tracer := otel.Tracer("microservice-tracer")
	w.Write([]byte("pong"))
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	//span := trace.SpanFromContext(ctx)
	span := trace.SpanFromContext(ctx)
	//spanCtx := span.SpanContext()
	span.SetName("HomeHandler")

	defer span.End()
	//injetando o header
	otel.GetTextMapPropagator().Inject(ctx, carrier)
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

func initProvider(serviceName, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// troca o contexto de atributo para timeout, tenta conectar em 1 segundo se não for vai dar timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	// config do exporter do trace. Poderia ser feito via http, mas estou dizendo que será feito via GRPC
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// bsp = batch span processor
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	// o provider via pegar o recurso e o span processor. Ele é o cara que irá fazer com que as infos serão consolidadas. Pelo exporter ele sabe oq vai falar com o grpc
	tracerProvider := sdktrace.NewTracerProvider(
		// with Sampler é a amostragem que quer ter par aenviar o trace, em dev ele envia para cada requisicao o trace pra gente (AlwaysSample), se fosse prod para não ter q enviar tudo, poderíamos controlar o exemplo de trace
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// vai propagar a informação usando dados de tracing
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// o shutdown desliga de forma graciosa
	return tracerProvider.Shutdown, nil
}
