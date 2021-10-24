package core

import (
	"bot/messaging"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type StockResponse struct {
	Symbol string `json:"symbol"`
	Date   string `json:"date"`
	Time   string `json:"time"`
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
}

func GetStockQuote(code string) (string, error) {
	if code == "" {
		return "", errors.New("invalid code")
	}

	data, err := readCSVFromUrl(messaging.STOCK_URL + code)
	if err != nil {
		return "", errors.New("error parsing CSV from URL")
	}

	dataFieldRows := data[1]
	stooqResponse := &StockResponse{
		Symbol: dataFieldRows[0],
		Close:  dataFieldRows[6],
	}

	var msg string
	if stooqResponse.Close != "N/D" {
		msg = fmt.Sprintf("%s quote is %v per share.", stooqResponse.Symbol, stooqResponse.Close)
	} else {
		msg = "Could not get stock quote."
	}

	return msg, nil
}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rows [][]string
	for _, e := range data {
		rows = append(rows, strings.Split(e[0], ","))
	}

	return rows, nil
}
