package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	log.Println("Server running on port 8080")

	http.HandleFunc("/cotacao", handleServer)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}

}

func handleServer(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	//add to the response of this handler the body of the implemented request
	w.Write(body)

	var data cotacao
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	dbCtx, dbCancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer dbCancel()

	err = insertCotacao(dbCtx, &data)
	if err != nil {
		panic(err)
	}

}

func insertCotacao(ctx context.Context, cotacao *cotacao) error {

	//add to the response of this get to the database

	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO cotacao (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cotacao.USDBRL.Code, cotacao.USDBRL.Codein, cotacao.USDBRL.Name, cotacao.USDBRL.High, cotacao.USDBRL.Low, cotacao.USDBRL.VarBid, cotacao.USDBRL.PctChange, cotacao.USDBRL.Bid, cotacao.USDBRL.Ask, cotacao.USDBRL.Timestamp, cotacao.USDBRL.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
