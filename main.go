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

var from, to, lang string
var apiKey string = "e6dd2bb259387cc30149060c"

type ExchangeRateResponse struct {
	BaseCode           string             `json:"base_code"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

var translations = map[string]map[string]string{
	"en": {
		"enter_amount":    "Enter amount (or 'exit' to quit): ",
		"invalid_amount":  "Invalid amount. Please enter a positive number.",
		"enter_from":      "Enter currency to convert from (e.g., USD): ",
		"enter_to":        "Enter currency to convert to (e.g., EUR): ",
		"missing_currencies": "Both 'from' and 'to' currencies must be specified.",
		"base_currency":   "Base Currency: ",
		"last_update":     "Last Update (Tashkent Time): ",
	},
	"ru": {
		"enter_amount":    "Введите сумму (или 'exit' для выхода): ",
		"invalid_amount":  "Недопустимая сумма. Пожалуйста, введите положительное число.",
		"enter_from":      "Введите валюту для конвертации (например, USD): ",
		"enter_to":        "Введите валюту для конвертации в (например, EUR): ",
		"missing_currencies": "Необходимо указать валюты 'from' и 'to'.",
		"base_currency":   "Базовая валюта: ",
		"last_update":     "Последнее обновление (Ташкентское время): ",
	},
	"uz": {
		"enter_amount":    "Summani kiriting (yoki chiqish uchun 'exit' kiriting): ",
		"invalid_amount":  "Noto'g'ri summa. Iltimos, musbat son kiriting.",
		"enter_from":      "O'zgarish uchun valyutani kiriting (masalan, USD): ",
		"enter_to":        "O'zgarish uchun valyutani kiriting (masalan, EUR): ",
		"missing_currencies": "'from' va 'to' valyutalari ko'rsatilishi kerak.",
		"base_currency":   "Asosiy valyuta: ",
		"last_update":     "Oxirgi yangilanish (Toshkent vaqti): ",
	},
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
	fmt.Println("Choose language (en, ru, uz):")
	fmt.Scanln(&lang)

	if lang != "en" && lang != "ru" && lang != "uz" {
		fmt.Println("Unsupported language, defaulting to English (en).")
		lang = "en"
	}

	var rootCmd = &cobra.Command{
		Use:   "convert",
		Short: translations[lang]["enter_amount"],
		Run: func(cmd *cobra.Command, args []string) {
			for {
				// Ask for the amount
				fmt.Print(translations[lang]["enter_amount"])
				var inputAmount string
				fmt.Scanln(&inputAmount)

				if strings.ToLower(inputAmount) == "exit" {
					break
				}

				var amount float64
				_, err := fmt.Sscanf(inputAmount, "%f", &amount)
				if err != nil || amount <= 0 {
					fmt.Println(translations[lang]["invalid_amount"])
					continue
				}

				fmt.Print(translations[lang]["enter_from"])
				fmt.Scanln(&from)
				from = strings.ToUpper(from)

				fmt.Print(translations[lang]["enter_to"])
				fmt.Scanln(&to)
				to = strings.ToUpper(to)

				if from == "" || to == "" {
					fmt.Println(translations[lang]["missing_currencies"])
					continue
				}

				result, response, err := convertCurrency(from, to, amount)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}

				tashkentLocation, _ := time.LoadLocation("Asia/Tashkent")
				lastUpdateTashkent := time.Unix(response.TimeLastUpdateUnix, 0).In(tashkentLocation).Format("Mon, 02 Jan 2006 15:04:05")

				fmt.Println("---------------------------------------------------------")
				fmt.Printf(ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+" is "+ColorGreen+"%.2f"+ColorReset+" "+ColorYellow+"%s"+ColorReset+"\n", amount, from, result, to)
				fmt.Printf(translations[lang]["base_currency"] + ColorYellow + "%s" + ColorReset + "\n", response.BaseCode)
				fmt.Printf(translations[lang]["last_update"]+"%s\n", lastUpdateTashkent)
				fmt.Println("---------------------------------------------------------")
			}
		},
	}

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
