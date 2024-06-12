package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/otel/inputApp/pkg/utils"
	"io"
	"net/http"
	"time"
)

const requestExpirationTime = 10 * time.Second

func main() {

}

type TempByCepResponse struct {
	City   string  `json:"city"`
	Temp_C float32 `json:"temp_c"`
	Temp_F float32 `json:"temp_f"`
	Temp_K float32 `json:"temp_k"`
}

type Address struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Source       string `json:"source"`
}

func GetTempByCep(cep string) (TempByCepResponse, error) {
	err := utils.ValidateCep(cep)
	if err != nil {
		return TempByCepResponse{}, utils.InvalidZipError
	}
	ctx := context.Background()
	// o contexto expira em 1 segundo!
	ctx, cancel := context.WithTimeout(ctx, requestExpirationTime)
	defer cancel() // de alguma forma nosso contexto será cancelado
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8090/temp/"+cep, nil)

	if err != nil {
		return TempByCepResponse{}, err
	}

	// faz a request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return TempByCepResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNotFound {
			return TempByCepResponse{}, utils.ZipNotFoundError
		}

		return TempByCepResponse{}, errors.New("unkown error")

	}

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Api fetch timeout exceeed.")
		return TempByCepResponse{}, errors.New("api fetch timeout exceeed")
	}

	// depois de tudo termina e faz o body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TempByCepResponse{}, err
	}
	var tempCepData TempByCepResponse
	err = json.Unmarshal(body, &tempCepData)
	if err != nil {
		return TempByCepResponse{}, err
	}

	//empty struct = valid format but no data
	if (tempCepData == TempByCepResponse{}) {
		return TempByCepResponse{}, utils.ZipNotFoundError
	}

	return tempCepData, nil
}
