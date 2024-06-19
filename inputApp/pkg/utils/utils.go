package utils

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type HttpError struct {
	Code    int
	Message string
}

// Error implements error.
//
//	func (e HttpError) Error() string {
//		return strings.Join([]string{fmt.Sprint(e.Code), e.Message}, " ")
//	}
func (e HttpError) Error() string {
	return e.Message
}

var InvalidZipError = HttpError{
	Code:    http.StatusUnprocessableEntity,
	Message: "422 invalid zipcode",
}

var ZipNotFoundError = HttpError{
	Code:    http.StatusNotFound,
	Message: "404 can not find zipcode",
}

func ValidateCep(cep string) error {
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

func IsMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(IsMn), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}
