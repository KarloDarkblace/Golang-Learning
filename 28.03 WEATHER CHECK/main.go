package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type WeatherData struct {
	City string `json:"name"`
	Main struct {
		Temperature float64 `json:"temp"`
		Humidity    float64 `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
}

const (
	VLD_CITY    = "Vladikavkaz"
	BSL_CITY    = "Beslan"
	API         = "f7860d6d03b23e9e51194de9d77a44e5"
	CSVFileName = "weather_data.csv"
	API_URL     = "http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s"
)

func GetWeatherData(city, apiKey string) (*WeatherData, error) {
	url := fmt.Sprintf(API_URL, city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERROR | API запрос выдал ошибку %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var weatherData WeatherData
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, err
	}

	return &weatherData, nil
}

func SaveWeatherDataToCSV(data *WeatherData) error {
	file, err := os.OpenFile(CSVFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("ERROR | Не удалось открыть файл CSV: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, weather := range data.Weather {
		record := []string{
			time.Now().Format("2006-01-02 15:04:05"),
			data.City,
			weather.Main,
			weather.Description,
			fmt.Sprintf("%.2f", data.Main.Temperature-273.15),
			fmt.Sprintf("%v", data.Main.Humidity),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("ERROR | Ошибка при записи в файл CSV: %w", err)
		}
	}

	return nil
}

func FetchAndSaveWeatherData(cities []string, apiKey string) {
	for _, city := range cities {
		data, err := GetWeatherData(city, apiKey)

		if err != nil {
			log.Printf("ERROR | Ошибка при получении данных о погоде для города %s: %v\n", city, err)
			continue
		}

		if err := SaveWeatherDataToCSV(data); err != nil {
			log.Printf("ERROR | Ошибка при сохранении данных о погоде в CSV: %v\n", err)
			continue
		}

		fmt.Printf("Данные о погоде для %s сохранены.\n", city)
	}
}

func main() {
	cities := []string{VLD_CITY, BSL_CITY}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	FetchAndSaveWeatherData(cities, API)

	for {
		select {
		case <-ticker.C:
			FetchAndSaveWeatherData(cities, API)
		case <-quit:
			log.Println("Программа завершена.")
			return
		}
	}
}
