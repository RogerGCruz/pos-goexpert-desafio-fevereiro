package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	urlAwesomeApi  = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dataSourceName = "../storage/cotation.db"
)

type Cotation struct {
	Bid string `json:"bid"`
}

type ResponseAwesomeApi struct {
	USDBRL UsdBrl `json:"USDBRL"`
}

type UsdBrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	TimeStamp  string `json:"timeStamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	initDb()
	initServer()
}

func initDb() {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotation (id INTEGER PRIMARY KEY autoincrement, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database created")
}

func initServer() {
	// Endpoint cotacao na porta 8080
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
	log.Println("Listening on port 8080")
}

func insertCotation(responseAwesomeApi ResponseAwesomeApi) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctxDb, cancelDb := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancelDb()

	if err != nil {
		log.Println("Error inserting AwesomeApi", err)
		return
	}

	_, err = db.ExecContext(ctxDb, "INSERT INTO cotation (bid) VALUES (?)", responseAwesomeApi.USDBRL.Bid)
}

func handler(w http.ResponseWriter, r *http.Request) {

	// Contexto com timeout 200ms para AwesomeApi
	ctxAwesomeApi, cancelAwesomeApi := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelAwesomeApi()

	// Contextos retornam erro nos logs se timeout insuficiente
	req, err := http.NewRequestWithContext(ctxAwesomeApi, "GET", urlAwesomeApi, nil)
	if err != nil {
		log.Println("Error creating request for AwesomeApi", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Consome API https://economia.awesomeapi.com.br/json/last/USD-BRL
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error making request for AwesomeApi", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var responseAwesomeApi ResponseAwesomeApi
	err = json.NewDecoder(resp.Body).Decode(&responseAwesomeApi)
	if err != nil {
		log.Println("Error decoding AwesomeApi", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	insertCotation(responseAwesomeApi)

	// Retorna JSON para cliente com o bid
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Cotation{Bid: responseAwesomeApi.USDBRL.Bid})
}
