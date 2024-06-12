package main

import (
	"strings"
	"testing"
)

func TestShouldReturnInvalidZipCode(t *testing.T) {
	cep := "123456789"
	cep = strings.ReplaceAll(cep, "-", "")

	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	err := validateCep(cep)
	if err == nil {
		t.Errorf("InputApp() did not returned an 422 error")
	}
	if err.Error() != "cep must contain exactly 8 characters" {
		t.Errorf("InputApp() did not returned the 8 character cep error")
	}
}
func TestShouldReturnValidZip(t *testing.T) {
	cep := "12345678"
	cep = strings.ReplaceAll(cep, "-", "")

	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	err := validateCep(cep)
	if err != nil {
		t.Errorf("InputApp() invalid zip code")
	}
}
