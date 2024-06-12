package external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/JonecoBoy/tempByCep/pkg/utils"
)

type AddressDataViaCep struct {
	Cep          string `json:"cep"`
	State        string `json:"uf"`
	City         string `json:"localidade"`
	Neighborhood string `json:"bairro"`
	Street       string `json:"logradouro"`
}

func ViaCep(cep string) (Address, error) {
	err := utils.ValidateCep(cep)
	if err != nil {
		return Address{}, utils.InvalidZipError
	}
	ctx := context.Background()
	// o contexto expira em 1 segundo!
	ctx, cancel := context.WithTimeout(ctx, requestExpirationTime)
	defer cancel() // de alguma forma nosso contexto será cancelado
	req, err := http.NewRequestWithContext(ctx, "GET", "http://viacep.com.br/ws/"+cep+"/json/", nil)

	if err != nil {
		return Address{}, err
	}

	// faz a request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return Address{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Address{}, utils.ZipNotFoundError

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
	var jsonData AddressDataViaCep
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return Address{}, err
	}

	//empty struct = valid format but no data
	if (jsonData == AddressDataViaCep{}) {
		return Address{}, utils.ZipNotFoundError
	}

	addressData := Address{
		Cep:          jsonData.Cep,
		State:        jsonData.State,
		City:         jsonData.City,
		Neighborhood: jsonData.Neighborhood,
		Street:       jsonData.Street,
		Source:       "ViaCEP",
	}

	return addressData, nil

}
