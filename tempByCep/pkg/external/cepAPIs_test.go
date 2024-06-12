package external

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JonecoBoy/tempByCep/pkg/utils"
)

func TestViaCep(t *testing.T) {
	cep := "20541155"
	result, err := ViaCep(cep)
	if err != nil {
		t.Errorf("ViaCep() returned an error: %v", err)
	}

	if strings.ReplaceAll(result.Cep, "-", "") != strings.ReplaceAll(cep, "-", "") {
		t.Errorf("ViaCep() returned an invalid CEP: %v", result.Cep)
	}

	fields := []string{
		"Cep", "State", "City", "Neighborhood", "Street", "Source",
	}

	viaCepVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := viaCepVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("ViaCep() did not return a a valid struct with the field %s", field)
		}
	}
}

func TestViaCepZipNotFound(t *testing.T) {
	cep := "90541155"
	_, err := ViaCep(cep)
	if err == nil {
		t.Fatalf("ViaCep() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.ZipNotFoundError.Error() {
		t.Errorf("ViaCep() did not return an " + utils.ZipNotFoundError.Error() + " zipcode error")
	}

}

func TestViaCepInvalidFormat(t *testing.T) {
	cep := "905411551"
	_, err := ViaCep(cep)
	if err == nil {
		t.Fatalf("ViaCep() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.InvalidZipError.Error() {
		t.Errorf("ViaCep() did not return an " + utils.InvalidZipError.Error() + " zipcode error")
	}
}

func TestBrasilApiCep(t *testing.T) {
	cep := "20541155"
	result, err := BrasilApiCep(cep)
	if err != nil {
		t.Errorf("BrasilApiCep() returned an error: %v", err)
	}

	if strings.ReplaceAll(result.Cep, "-", "") != strings.ReplaceAll(cep, "-", "") {
		t.Errorf("BrasilApiCep() returned an invalid CEP: %v", result.Cep)
	}

	fields := []string{
		"Cep", "State", "City", "Neighborhood", "Street", "Source",
	}

	viaCepVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := viaCepVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("BrasilApiCep() did not return a valid struct with the field %s", field)
		}
	}
}

func TestBrasilApiCepZipNotFound(t *testing.T) {
	cep := "90541155"
	_, err := BrasilApiCep(cep)
	if err == nil {
		t.Fatalf("BrasilApi() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.ZipNotFoundError.Error() {
		t.Errorf("BrasilApi() did not return an " + utils.ZipNotFoundError.Error() + " zipcode error")
	}

}

func TestBrasilApiCepInvalidFormat(t *testing.T) {
	cep := "90541A155"
	_, err := BrasilApiCep(cep)
	if err == nil {
		t.Fatalf("ViaCep() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.InvalidZipError.Error() {
		t.Errorf("BrasilApi() did not return an " + utils.InvalidZipError.Error() + " zipcode error")
	}
}
