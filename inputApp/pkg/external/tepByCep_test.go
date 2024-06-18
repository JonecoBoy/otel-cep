package external

import (
	"context"
	"reflect"
	"testing"
)

func TestShouldReturnCepWithCityAndTemperature(t *testing.T) {
	cep := "20541155"
	result, err := GetTempByCep(context.Background(), cep)
	if err != nil {
		t.Errorf("getTempByCep() returned an error: %v", err)
	}
	fields := []string{
		"City", "Temp_C", "Temp_F", "Temp_K",
	}

	tempByCep := reflect.ValueOf(result)
	for _, field := range fields {
		val := tempByCep.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("CepConcurrency() did not return a Marine struct with the field %s", field)
		}
	}
	if result.City != "Rio de Janeiro" {
		t.Errorf("CepConcurrency() did not return the correct city")
	}
}

func TestShouldReturnInvalidZipCodeError(t *testing.T) {
	cep := "245A159B"
	result, err := GetTempByCep(context.Background(), cep)
	if err == nil {
		t.Errorf("getTempByCep() did not returned an error: %v", err)
	}
	fields := []string{
		"City", "Temp_C", "Temp_F", "Temp_K",
	}

	tempByCep := reflect.ValueOf(result)
	for _, field := range fields {
		val := tempByCep.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("getTempByCep() did not return a proper struct with the field %s", field)
		}
	}
	if result.City != "" {
		t.Errorf("getTempByCep() did not return an empty city")
	}
	if result.Temp_C != 0 {
		t.Errorf("getTempByCep() did not return an Temp_C as 0")
	}
	if result.Temp_F != 0 {
		t.Errorf("getTempByCep() did not return an Temp_F as 0")
	}
	if result.Temp_K != 0 {
		t.Errorf("getTempByCep() did not return an Temp_K as 0")
	}
	if err.Error() != "422 invalid zipcode" {
		t.Errorf("getTempByCep() did not return the correct error message")
	}
}

func TestShouldReturnCanNotFindZipCode(t *testing.T) {
	//404
	cep := "99900028"
	result, err := GetTempByCep(context.Background(), cep)
	if err == nil {
		t.Errorf("getTempByCep() did not returned an error: %v", err)
	}
	fields := []string{
		"City", "Temp_C", "Temp_F", "Temp_K",
	}

	tempByCep := reflect.ValueOf(result)
	for _, field := range fields {
		val := tempByCep.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("getTempByCep() did not return a proper struct with the field %s", field)
		}
	}
	if result.City != "" {
		t.Errorf("getTempByCep() did not return an empty city")
	}
	if result.Temp_C != 0 {
		t.Errorf("getTempByCep() did not return an Temp_C as 0")
	}
	if result.Temp_F != 0 {
		t.Errorf("getTempByCep() did not return an Temp_F as 0")
	}
	if result.Temp_K != 0 {
		t.Errorf("getTempByCep() did not return an Temp_K as 0")
	}
	if err.Error() != "404 can not find zipcode" {
		t.Errorf("getTempByCep() did not return the correct error message")
	}
}
