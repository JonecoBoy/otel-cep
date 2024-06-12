package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JonecoBoy/tempByCep/pkg/utils"
)

const requestExpirationTime = 10 * time.Second

type Address struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Source       string `json:"source"`
}

type errorResponse struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func BrasilApiCep(cep string) (Address, error) {
	err := utils.ValidateCep(cep)
	if err != nil {
		return Address{}, utils.InvalidZipError
	}
	ctx := context.Background()
	// o contexto expira em 1 segundo!
	ctx, cancel := context.WithTimeout(ctx, requestExpirationTime)
	defer cancel() // de alguma forma nosso contexto será cancelado
	req, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/"+cep, nil)

	if err != nil {
		return Address{}, err
	}

	// faz a request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return Address{}, err
	}

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNotFound {
			return Address{}, utils.ZipNotFoundError
		}

		return Address{}, errors.New("unkown error")

	}

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Api fetch timeout exceeed.")
		return Address{}, errors.New("api fetch timeout exceeed")
	}

	// depois de tudo termina e faz o body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Address{}, err
	}
	var addressData Address
	err = json.Unmarshal(body, &addressData)
	if err != nil {
		return Address{}, err
	}

	//empty struct = valid format but no data
	if (addressData == Address{}) {
		return Address{}, utils.ZipNotFoundError
	}

	addressData.Source = "brasilAPI"
	return addressData, nil
}
