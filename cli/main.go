package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

var apiKey string = "e6dd2bb259387cc30149060c"
var lang string

type ExchangeRateResponse struct {
	BaseCode           string             `json:"base_code"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

func convertCurrency(from, to string, amount float64) (float64, ExchangeRateResponse, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	rate, response, err := getExchangeRate(from, to)
	if err != nil {
		return 0, response, err
	}
	return amount * rate, response, nil
}

func main() {
	var from, to string
	var amount float64

	var rootCmd = &cobra.Command{
		Use:   "convert",
		Short: "Currency Converter CLI",
		Run: func(cmd *cobra.Command, args []string) {
			if from == "" || to == "" || amount <= 0 {
				printError("Invalid input. Ensure 'from', 'to' currencies and a positive 'amount' are provided.")
				return
			}

			result, response, err := convertCurrency(from, to, amount)
			if err != nil {
				printError("Error: " + err.Error())
				return
			}

			printResult(response, amount, from, to, result)
		},
	}

	rootCmd.Flags().StringVarP(&from, "from", "f", "", "Currency to convert from (e.g., USD)")
	rootCmd.Flags().StringVarP(&to, "to", "t", "", "Currency to convert to (e.g., EUR)")
	rootCmd.Flags().Float64VarP(&amount, "amount", "a", 0, "Amount to convert")
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "en", "Language for output (en, ru, uz)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func getExchangeRate(from, to string) (float64, ExchangeRateResponse, error) {
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", apiKey, from)

	resp, err := http.Get(url)
	if err != nil {
		return 0, ExchangeRateResponse{}, fmt.Errorf("could not fetch exchange rate: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, ExchangeRateResponse{}, fmt.Errorf("could not read response body: %v", err)
	}

	var result ExchangeRateResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, ExchangeRateResponse{}, fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	rate, found := result.ConversionRates[to]
	if !found {
		return 0, result, fmt.Errorf("rate not found for %s", to)
	}

	return rate, result, nil
}

func printResult(response ExchangeRateResponse, amount float64, from, to string, result float64) {
	tashkentLocation, _ := time.LoadLocation("Asia/Tashkent")
	lastUpdateTashkent := time.Unix(response.TimeLastUpdateUnix, 0).In(tashkentLocation).Format("Mon, 02 Jan 2006 15:04:05")

	switch lang {
	case "uz":
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+" -> "+ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+"\n", amount, from, result, to)
		fmt.Printf("Asosiy valyuta: "+ColorYellow+"%s"+ColorReset+"\n", response.BaseCode)
		fmt.Printf("So'nggi yangilanish (Toshkent vaqti): %s\n", lastUpdateTashkent)
		fmt.Println("---------------------------------------------------------")
	case "ru":
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+" -> "+ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+"\n", amount, from, result, to)
		fmt.Printf("Основная валюта: "+ColorYellow+"%s"+ColorReset+"\n", response.BaseCode)
		fmt.Printf("Последнее обновление (Ташкентское время): %s\n", lastUpdateTashkent)
		fmt.Println("---------------------------------------------------------")
	default:
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+" -> "+ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+"\n", amount, from, result, to)
		fmt.Printf("Base Currency: "+ColorYellow+"%s"+ColorReset+"\n", response.BaseCode)
		fmt.Printf("Last Update (Tashkent Time): %s\n", lastUpdateTashkent)
		fmt.Println("---------------------------------------------------------")
	}
}

func printError(message string) {
	switch lang {
	case "uz":
		fmt.Println("Xato:", message)
	case "ru":
		fmt.Println("Ошибка:", message)
	default:
		fmt.Println("Error:", message)
	}
}
