package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	urlCotationApi = "http://localhost:8080/cotacao"
	pathFile       = "./cotation.txt"
)

type Cotation struct {
	Bid string `json:"bid"`
}

func main() {
	// Timeout 300ms para receber do server
	// Contexto retorna erro nos logs se timeout insuficiente
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlCotationApi, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Realiza requisição HTTP na API em server.go solicitando cotação
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making request (timeout?):", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body: ", err)
	}

	// Recebe apenas campo "bid" definido na struct Cotation
	var cotation Cotation
	err = json.Unmarshal(body, &cotation)
	if err != nil {
		log.Fatal("Error parsing response JSON body:", err)
	}

	file, err := os.Create(pathFile)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	// Salva em "cotacao.txt" no formato "Dólar: {valor}"
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", cotation.Bid))
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}
	log.Println("Cotation as saved in file:", file.Name())
}
