package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/external"
	"github.com/JonecoBoy/otel-cep/inputApp/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
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
	shutdown, err := initProvider("inputApp", "otel-collector:4317")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
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

	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	//span := trace.SpanFromContext(ctx)
	//span := trace.SpanFromContext(ctx)
	//spanCtx := span.SpanContext()
	//span.SetName("HomeHandler")
	tracer := otel.Tracer("microservice-tracer")
	ctx, span := tracer.Start(ctx, "iniciando")
	defer span.End()
	//injetando o header
	otel.GetTextMapPropagator().Inject(ctx, carrier)

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
