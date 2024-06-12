package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JonecoBoy/tempByCep/pkg/external"
	"github.com/JonecoBoy/tempByCep/pkg/utils"
)

func TestCepConcurrency(t *testing.T) {
	cep := "20541-155"
	result, err := CepConcurrency(cep)
	if err != nil {
		t.Errorf("CepConcurrency() returned an error: %v", err)
	}

	if strings.ReplaceAll(result.Cep, "-", "") != strings.ReplaceAll(cep, "-", "") {
		t.Errorf("CepConcurrency() returned an invalid CEP: %v", result.Cep)
	}

	fields := []string{
		"Cep", "State", "City", "Neighborhood", "Street", "Source",
	}

	concurrencyCepVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := concurrencyCepVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("CepConcurrency() did not return a Marine struct with the field %s", field)
		}
	}
}

func TestCepConcurrencyZipNotFound(t *testing.T) {
	cep := "90541155"
	_, err := CepConcurrency(cep)
	if err == nil {
		t.Fatalf("CepConcurrency() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.ZipNotFoundError.Error() {
		t.Errorf("CepConcurrency() did not return an " + utils.ZipNotFoundError.Error() + " error")
	}

}

func TestCepConcurrencyInvalidZipFormat(t *testing.T) {
	cep := "905411551"
	_, err := CepConcurrency(cep)
	if err == nil {
		t.Fatalf("CepConcurrency() returned a value instead of an err: %v", err)
	}
	if err.Error() != utils.InvalidZipError.Error() {
		t.Errorf("CepConcurrency() did not return an " + utils.InvalidZipError.Error() + " error")
	}
}

func TestGetTempByCep(t *testing.T) {
	cep := "25900-028"
	result, err := CepConcurrency(cep)
	if err != nil {
		t.Errorf("CepConcurrency() returned an error: %v", err)
	}

	if strings.ReplaceAll(result.Cep, "-", "") != strings.ReplaceAll(cep, "-", "") {
		t.Errorf("CepConcurrency() returned an invalid CEP: %v", result.Cep)
	}

	fields := []string{
		"Cep", "State", "City", "Neighborhood", "Street", "Source",
	}

	viaCepVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := viaCepVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("CepConcurrency() did not return a Marine struct with the field %s", field)
		}
	}

	query := strings.Join([]string{utils.RemoveAccents(result.City), utils.RemoveAccents(result.State), "brazil"}, "-")
	lang := "pt"

	result2, err := external.CurrentWeather(query, lang)
	if err != nil {
		t.Errorf("Current() returned an error: %v", err)
	}

	if *result2.Current == (external.Current{}) {
		t.Errorf("Current() returned an empty Current struct")
	}

	if *result2.Current.Condition == (external.Condition{}) {
		t.Errorf("Current() returned an empty Current struct")
	}

	// testando se os principais campos estão de volta
	fields2 := []string{
		"WindMph", "WindKph", "WindDegree", "WindDir", "PressureMb", "PressureIn", "PrecipMm", "PrecipIn",
		"Humidity", "Cloud", "FeelslikeC", "FeelslikeF", "TempC", "TempF", "IsDay", "WindMph",
		"WindKph", "WindDegree", "WindDir", "Humidity", "Uv", "GustMph", "GustKph",
	}

	currentVal := reflect.ValueOf(*result2.Current) // Dereference the pointer here
	for _, field := range fields2 {
		val := currentVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("Current() did not return a Current struct with the field %s", field)
		}
	}
}
