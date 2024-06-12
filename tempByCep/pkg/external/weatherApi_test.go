package external

import (
	"reflect"
	"testing"
	"time"
)

func TestDoRequest(t *testing.T) {
	method := "GET"
	path := "current.json"
	params := map[string]string{
		"q": "London",
	}

	resp, err := doRequest(method, path, params)
	if err != nil {
		t.Errorf("doRequest() returned an error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("doRequest() returned status code %v, expected 200", resp.StatusCode)
	}
}

func TestSearch(t *testing.T) {
	query := "mage-rio de janeiro-brazil"
	expected := searchReturn{

		Id:      279745,
		Name:    "Mage",
		Region:  "Rio de Janeiro",
		Country: "Brazil",
		Lat:     -22.66,
		Lon:     -43.02,
		Url:     "mage-rio-de-janeiro-brazil",
	}

	result, err := search(query)
	if err != nil {
		t.Errorf("search() returned an error: %v", err)
	}

	if result != expected {
		t.Errorf("search() returned unexpected result: got %v want %v", result, expected)
	}
}

func TestCurrent(t *testing.T) {
	query := "mage-rio de janeiro-brazil"
	lang := "pt"

	result, err := CurrentWeather(query, lang)
	if err != nil {
		t.Errorf("Current() returned an error: %v", err)
	}

	if *result.Current == (Current{}) {
		t.Errorf("Current() returned an empty Current struct")
	}

	if *result.Current.Condition == (Condition{}) {
		t.Errorf("Current() returned an empty Current struct")
	}

	// testando se os principais campos estão de volta
	fields := []string{
		"WindMph", "WindKph", "WindDegree", "WindDir", "PressureMb", "PressureIn", "PrecipMm", "PrecipIn",
		"Humidity", "Cloud", "FeelslikeC", "FeelslikeF", "TempC", "TempF", "IsDay", "WindMph",
		"WindKph", "WindDegree", "WindDir", "Humidity", "Uv", "GustMph", "GustKph",
	}

	currentVal := reflect.ValueOf(*result.Current) // Dereference the pointer here
	for _, field := range fields {
		val := currentVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("Current() did not return a Current struct with the field %s", field)
		}
	}
}

func TestForecast(t *testing.T) {
	query := "mage-rio de janeiro-brazil"
	lang := "pt"
	days := 3

	result, err := forecast(query, lang, days)
	if err != nil {
		t.Errorf("forecast() returned an error: %v", err)
	}

	// validar se o slice/array forecast day está presente e possui o mesmo tamanho de day
	if len(*result.Forecast.ForecastDay) != 3 {
		t.Errorf("forecast() returned an empty ForecastDay slice")
	}

	// Validar a existencia dos principais campos
	fields := []string{
		"Date", "DateEpoch", "Day", "Astro", "Hour",
	}

	for _, day := range *result.Forecast.ForecastDay {
		dayVal := reflect.ValueOf(day)
		for _, field := range fields {
			val := dayVal.FieldByName(field)
			if !val.IsValid() {
				t.Errorf("forecast() did not return a ForecastDay struct with the field %s", field)
			}
		}
	}
}
func TestIP(t *testing.T) {
	ipaddress := "192.168.1.1"

	result, err := ip(ipaddress)
	if err != nil {
		t.Errorf("ip() returned an error: %v", err)
	}

	// Check if the fields are present
	if result == (IP{}) {
		t.Errorf("ip() returned an empty IP struct")
	}

	// Check if the specific fields in the IP struct are present
	fields := []string{
		"IP", "Type", "ContinentCode", "ContinentName", "CountryCode", "CountryName", "IsEU", "GeonameID", "City", "Region", "Lat", "Lon", "TzID", "LocaltimeEpoch", "Localtime",
	}

	ipVal := reflect.ValueOf(result) // No need to dereference the pointer here as the result is not a pointer
	for _, field := range fields {
		val := ipVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("ip() did not return an IP struct with the field %s", field)
		}
	}
}
func TestFuture(t *testing.T) {
	code := "mage-rio de janeiro-brazil"
	lang := "pt"

	date := time.Now().Format("2006-01-02") // error test

	_, err := future(code, lang, date)
	if err == nil {
		t.Errorf("future() did not return an error for an invalid date")
	}

	//valid test
	date = time.Now().AddDate(0, 0, 20).Format("2006-01-02")

	result, err := future(code, lang, date)
	if err != nil {
		t.Errorf("future() returned an error: %v", err)
	}

	// Check if the fields are present
	if *result.Location == (Location{}) {
		t.Errorf("future() returned an empty Location struct")
	}

	if *result.Forecast == (ForecastBase{}) {
		t.Errorf("future() returned an empty ForecastBase struct")
	}

	// Check if the specific fields in the Forecast struct are present
	fields := []string{
		"Location", "Current", "Forecast",
	}

	forecastVal := reflect.ValueOf(result) // No need to dereference the pointer here as the result is not a pointer
	for _, field := range fields {
		val := forecastVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("future() did not return a Forecast struct with the field %s", field)
		}
	}
}

func TestTimezone(t *testing.T) {
	code := "mage-rio de janeiro-brazil"

	result, err := timezone(code)
	if err != nil {
		t.Errorf("timezone() returned an error: %v", err)
	}

	if result == (TimeZone{}) {
		t.Errorf("timezone() returned an empty TimeZone struct")
	}
}

func TestAstronomy(t *testing.T) {
	query := "mage-rio de janeiro-brazil"
	date := "2024-01-01" // This date should be on or after 1st Jan, 2015

	result, err := astronomy(query, date)
	if err != nil {
		t.Errorf("astronomy() returned an error: %v", err)
	}

	// Check if the fields are present
	if *result.Location == (Location{}) {
		t.Errorf("astronomy() returned an empty Location struct")
	}

	if *result.Astronomy == (AstronomyBase{}) {
		t.Errorf("astronomy() returned an empty AstronomyBase struct")
	}

	// Check if the specific fields in the Astronomy struct are present
	fields := []string{
		"Location", "Astronomy",
	}

	astronomyVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := astronomyVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("astronomy() did not return an Astronomy struct with the field %s", field)
		}
	}
}

func TestMarine(t *testing.T) {
	query := "mage - rio de janeiro - brazil"
	days := 3

	result, err := marine(query, "pt", days, "", 0, 0)
	if err != nil {
		t.Errorf("marine() returned an error: %v", err)
	}

	// Check if the fields are present
	if *result.Location == (Location{}) {
		t.Errorf("marine() returned an empty Location struct")
	}

	if *result.Forecast == (ForecastBase{}) {
		t.Errorf("marine() returned an empty Forecast struct")
	}

	if len(*result.Forecast.ForecastDay) != 3 {
		t.Errorf("forecast() returned an ForecastDay slice with wrong length")
	}

	fields := []string{
		"Location", "Forecast",
	}

	marineVal := reflect.ValueOf(result)
	for _, field := range fields {
		val := marineVal.FieldByName(field)
		if !val.IsValid() {
			t.Errorf("marine() did not return a Marine struct with the field %s", field)
		}
	}
}
