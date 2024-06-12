package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type searchReturn struct {
	Id      int32   `json:"id"`
	Name    string  `json:"name"`
	Region  string  `json:"region"`
	Country string  `json:"country"`
	Lat     float32 `json:"lat"`
	Lon     float32 `json:"lon"`
	Url     string  `json:"url"`
}

type Location struct {
	Name           string  `json:"name"`
	Region         string  `json:"region"`
	Country        string  `json:"country"`
	Lat            float32 `json:"lat"`
	Lon            float32 `json:"lon"`
	TzID           string  `json:"tz_id"`
	LocaltimeEpoch int32   `json:"localtime_epoch"`
	Localtime      string  `json:"localtime"`
}

type Condition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
	Code int32  `json:"code"`
}

type AirQuality struct {
	Co           float32 `json:"co"`
	No2          float32 `json:"no2"`
	O3           float32 `json:"o3"`
	So2          float32 `json:"so2"`
	Pm2_5        float32 `json:"pm2_5"`
	Pm10         float32 `json:"pm10"`
	UsEpaIndex   float32 `json:"us-epa-index"`
	GbDefraIndex float32 `json:"gb-defra-index"`
}

type CurrentModel struct {
	Location *Location `json:"location"`
	Current  *Current  `json:"current"`
}

type Current struct {
	LastUpdatedEpoch int32       `json:"last_updated_epoch"`
	LastUpdated      string      `json:"last_updated"`
	TempC            float32     `json:"temp_c"`
	TempF            float32     `json:"temp_f"`
	IsDay            float32     `json:"is_day"`
	Condition        *Condition  `json:"condition"`
	WindMph          float32     `json:"wind_mph"`
	WindKph          float32     `json:"wind_kph"`
	WindDegree       float32     `json:"wind_degree"`
	WindDir          string      `json:"wind_dir"`
	PressureMb       float32     `json:"pressure_mb"`
	PressureIn       float32     `json:"pressure_in"`
	PrecipMm         float32     `json:"precip_mm"`
	PrecipIn         float32     `json:"precip_in"`
	Humidity         float32     `json:"humidity"`
	Cloud            float32     `json:"cloud"`
	FeelslikeC       float32     `json:"feelslike_c"`
	FeelslikeF       float32     `json:"feelslike_f"`
	VisKm            float32     `json:"vis_km"`
	VisMiles         float32     `json:"vis_miles"`
	Uv               float32     `json:"uv"`
	GustMph          float32     `json:"gust_mph"`
	GustKph          float32     `json:"gust_kph"`
	AirQuality       *AirQuality `json:"air_quality"`
}

type IP struct {
	IP             string  `json:"ip"`
	Type           string  `json:"type"`
	ContinentCode  string  `json:"continent_code"`
	ContinentName  string  `json:"continent_name"`
	CountryCode    string  `json:"country_code"`
	CountryName    string  `json:"country_name"`
	IsEU           string  `json:"is_eu"`
	GeonameID      int32   `json:"geoname_id"`
	City           string  `json:"city"`
	Region         string  `json:"region"`
	Lat            float32 `json:"lat"`
	Lon            float32 `json:"lon"`
	TzID           string  `json:"tz_id"`
	LocaltimeEpoch float32 `json:"localtime_epoch"`
	Localtime      string  `json:"localtime"`
}

type Day struct {
	MaxtempC          float32    `json:"maxtemp_c"`
	MaxtempF          float32    `json:"maxtemp_f"`
	MintempC          float32    `json:"mintemp_c"`
	MintempF          float32    `json:"mintemp_f"`
	AvgtempC          float32    `json:"avgtemp_c"`
	AvgtempF          float32    `json:"avgtemp_f"`
	MaxwindMph        float32    `json:"maxwind_mph"`
	MaxwindKph        float32    `json:"maxwind_kph"`
	TotalprecipMm     float32    `json:"totalprecip_mm"`
	TotalprecipIn     float32    `json:"totalprecip_in"`
	AvgvisKm          float32    `json:"avgvis_km"`
	AvgvisMiles       float32    `json:"avgvis_miles"`
	Avghumidity       float32    `json:"avghumidity"`
	DailyWillItRain   float32    `json:"daily_will_it_rain"`
	DailyChanceOfRain float32    `json:"daily_chance_of_rain"`
	DailyWillItSnow   float32    `json:"daily_will_it_snow"`
	DailyChanceOfSnow float32    `json:"daily_chance_of_snow"`
	Condition         *Condition `json:"condition"`
	Uv                float32    `json:"uv"`
}

type Astro struct {
	Sunrise          string   `json:"sunrise"`
	Sunset           string   `json:"sunset"`
	Moonrise         string   `json:"moonrise"`
	Moonset          string   `json:"moonset"`
	MoonPhase        string   `json:"moon_phase"`
	MoonIllumination float64  `json:"moon_illumination"`
	IsMoonUp         *float32 `json:"is_moon_up"`
	IsSunUp          *float32 `json:"is_sun_up"`
}

type Hour struct {
	TimeEpoch    float32    `json:"time_epoch"`
	Time         string     `json:"time"`
	TempC        float32    `json:"temp_c"`
	TempF        float32    `json:"temp_f"`
	IsDay        float32    `json:"is_day"`
	Condition    *Condition `json:"condition"`
	WindMph      float32    `json:"wind_mph"`
	WindKph      float32    `json:"wind_kph"`
	WindDegree   float32    `json:"wind_degree"`
	WindDir      string     `json:"wind_dir"`
	PressureMb   float32    `json:"pressure_mb"`
	PressureIn   float32    `json:"pressure_in"`
	PrecipMm     float32    `json:"precip_mm"`
	PrecipIn     float32    `json:"precip_in"`
	Humidity     float32    `json:"humidity"`
	Cloud        float32    `json:"cloud"`
	FeelslikeC   float32    `json:"feelslike_c"`
	FeelslikeF   float32    `json:"feelslike_f"`
	WindchillC   float32    `json:"windchill_c"`
	WindchillF   float32    `json:"windchill_f"`
	HeatindexC   float32    `json:"heatindex_c"`
	HeatindexF   float32    `json:"heatindex_f"`
	DewpointC    float32    `json:"dewpoint_c"`
	DewpointF    float32    `json:"dewpoint_f"`
	WillItRain   float32    `json:"will_it_rain"`
	ChanceOfRain float32    `json:"chance_of_rain"`
	WillItSnow   float32    `json:"will_it_snow"`
	ChanceOfSnow float32    `json:"chance_of_snow"`
	VisKm        float32    `json:"vis_km"`
	VisMiles     float32    `json:"vis_miles"`
	GustMph      float32    `json:"gust_mph"`
	GustKph      float32    `json:"gust_kph"`
	Uv           float32    `json:"uv"`
}

type Marine struct {
	Location *Location
	Forecast *ForecastBase
}

type AstronomyBase struct {
	Astro *Astro `json:"astro"`
}

type Astronomy struct {
	Location  *Location      `json:"location"`
	Astronomy *AstronomyBase `json:"astronomy"`
}

type ForecastDay struct {
	Date      string  `json:"date"`
	DateEpoch int32   `json:"date_epoch"`
	Day       *Day    `json:"day"`
	Astro     *Astro  `json:"astro"`
	Hour      *[]Hour `json:"hour"`
}

type Forecast struct {
	Location *Location     `json:"location"`
	Current  *Current      `json:"current"`
	Forecast *ForecastBase `json:"forecast"`
}
type ForecastBase struct {
	ForecastDay *[]ForecastDay `json:"forecastday"`
}

type TimeZone struct {
	Location *Location `json:"location"`
}

const apiKey = "cf7ea6d61fd247d78e9171401240206"
const baseUrl = "https://api.weatherapi.com/v1"

const weatherRequestExpirationTime = 60 * time.Second

func doRequest(method string, path string, params map[string]string) (*http.Response, error) {
	path = strings.ReplaceAll(path, "/", "")
	method = strings.ToUpper(method)
	//ctx := context.Background()
	//ctx, cancel := context.WithTimeout(ctx, weatherRequestExpirationTime)
	//defer cancel() // de alguma forma nosso contexto será cancelado

	u, err := url.Parse(baseUrl + "/" + path)
	if err != nil {
		return nil, err
	}

	// add api key
	params["key"] = apiKey

	//parseando e adicionando do map
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	//req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	req, err := http.NewRequest(method, u.String(), nil)

	if err != nil {
		return nil, err
	}

	// faz a request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	//if ctx.Err() == context.DeadlineExceeded {
	//	fmt.Println("Api fetch timeout exceeed.")
	//	return nil, errors.New("Api fetch timeout exceeed.")
	//}

	return resp, nil
}

func CurrentWeather(query string, lang string) (CurrentModel, error) {
	// Define the parameters for the request
	params := map[string]string{
		"q":    query,
		"lang": lang,
	}

	// Make the request
	resp, err := doRequest("GET", "current.json", params)
	if err != nil {
		return CurrentModel{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CurrentModel{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into a Current struct
	var current CurrentModel
	err = json.Unmarshal(body, &current)
	if err != nil {
		return CurrentModel{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the current weather data
	return current, nil
}

func forecast(query string, lang string, days int) (Forecast, error) {
	// Define the parameters for the request
	params := map[string]string{
		"q":    query,
		"lang": lang,
		"days": strconv.Itoa(days),
	}

	// Make the request
	resp, err := doRequest("GET", "forecast.json", params)
	if err != nil {
		return Forecast{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Forecast{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into a Forecast struct
	var forecast Forecast
	err = json.Unmarshal(body, &forecast)
	if err != nil {
		return Forecast{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the forecast data
	return forecast, nil
}

func ip(ipaddress string) (IP, error) {
	// Remove all dots from the IP address
	ipaddress = strings.ReplaceAll(ipaddress, ".", "")

	// Convert the IP address to an integer
	ipInt, err := strconv.ParseInt(ipaddress, 10, 32)
	if err != nil {
		return IP{}, fmt.Errorf("converting IP address to integer: %v", err)
	}

	// Define the parameters for the request
	params := map[string]string{
		"q": strconv.FormatInt(ipInt, 10),
	}

	// Make the request
	resp, err := doRequest("GET", "ip.json", params)
	if err != nil {
		return IP{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return IP{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into an IP struct
	var ip IP
	err = json.Unmarshal(body, &ip)
	if err != nil {
		return IP{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the IP data
	return ip, nil
}

func search(toSearch string) (searchReturn, error) {
	param := map[string]string{"q": toSearch}
	resp, err := doRequest("GET", "search.json", param)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return searchReturn{}, fmt.Errorf("reading response body: %v", err)
	}
	var results []searchReturn
	err = json.Unmarshal(body, &results)
	if err != nil {
		return searchReturn{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	if len(results) == 0 {
		return searchReturn{}, fmt.Errorf("no results returned")
	}

	return results[0], nil
}

func future(query string, lang string, date string) (Forecast, error) {
	// Parse the date string into a time.Time value
	dt, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Forecast{}, fmt.Errorf("parsing date: %v", err)
	}

	// Get the current date
	now := time.Now()

	// Calculate the difference between the current date and the provided date
	diff := dt.Sub(now).Hours() / 24

	// Check if the date is between 14 and 300 days from today
	if diff < 14 || diff > 300 {
		return Forecast{}, errors.New("date should be between 14 and 300 days from today")
	}

	// Define the parameters for the request
	params := map[string]string{
		"q":    query,
		"dt":   date,
		"lang": lang,
	}

	// Make the request
	resp, err := doRequest("GET", "forecast.json", params)
	if err != nil {
		return Forecast{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Forecast{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into a Forecast struct
	var forecast Forecast
	err = json.Unmarshal(body, &forecast)
	if err != nil {
		return Forecast{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the forecast data
	return forecast, nil
}

func timezone(query string) (TimeZone, error) {
	params := map[string]string{
		"q": query,
	}
	response, err := doRequest("GET", "timezone.json", params)
	if err != nil {
		return TimeZone{}, err
	}

	dataJson, err := io.ReadAll(response.Body)
	if err != nil {
		return TimeZone{}, errors.New("Error reading response body")

	}
	location := TimeZone{}
	err = json.Unmarshal(dataJson, &location)
	if err != nil {
		return TimeZone{}, errors.New("Error unmarshalling response body")
	}

	return location, nil
}

func astronomy(query string, date string) (Astronomy, error) {
	// Define the parameters for the request
	params := map[string]string{
		"q":  query,
		"dt": date,
	}

	// Make the request
	resp, err := doRequest("GET", "astronomy.json", params)
	if err != nil {
		return Astronomy{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Astronomy{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into an Astronomy struct
	var astronomy Astronomy
	err = json.Unmarshal(body, &astronomy)
	if err != nil {
		return Astronomy{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the astronomy data
	return astronomy, nil
}

func marine(query string, lang string, days int, date string, unixdt int, hour int) (Marine, error) {
	params := map[string]string{
		"q":    query,
		"days": strconv.Itoa(days),
	}

	// Add optional parameters if they are not their zero values
	if date != "" {
		params["dt"] = date
	}
	if unixdt != 0 {
		params["unixdt"] = strconv.Itoa(unixdt)
	}
	if hour != 0 {
		params["hour"] = strconv.Itoa(hour)
	}
	if lang != "" {
		params["lang"] = lang
	}

	// Make the request
	resp, err := doRequest("GET", "marine.json", params)
	if err != nil {
		return Marine{}, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Marine{}, fmt.Errorf("reading response body: %v", err)
	}

	// Unmarshal the JSON response into a Marine struct
	var marine Marine
	err = json.Unmarshal(body, &marine)
	if err != nil {
		return Marine{}, fmt.Errorf("unmarshalling response body: %v", err)
	}

	// Return the marine data
	return marine, nil
}
