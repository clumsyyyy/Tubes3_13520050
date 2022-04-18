package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

type Response struct {
	Time       string `json:"time"`
	Name       string `json:"name"`
	Disease    string `json:"disease"`
	Found      bool   `json:"found"`
	Similarity int    `json:"similarity"`
	Indexes    []int  `json:"indexes"`
}

func match(c echo.Context) error {
	// if data yg diminta gada, return bad request
	if c.FormValue("text") == "" || c.FormValue("disease") == ""|| c.FormValue("method") == "" {
		return c.JSON(http.StatusBadRequest, "text or disease or method is empty")
	}
	text := c.FormValue("text")
	disease := c.FormValue("disease")
	fmt.Println("[REQUEST IN]", text, disease)
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec("USE stima3;")
	if err != nil {
		panic(err)
	}

	var pattern string

	if err := db.QueryRow("SELECT rantai FROM penyakit WHERE nama = ?", disease).Scan(&pattern); err != nil {
		panic(err)
	}
	var result []int

	similarity := int(countSimilarities(text, pattern) * 100 / (len(pattern)))
	if c.FormValue("method") == "BM" {
		result = BMAlgo(text, pattern)
	} else {
		result = KMPAlgo(text, pattern)
	}

	res := &Response{
		Time:       time.Now().Local().Format("2006-01-02"),
		Name:       c.FormValue("username"),
		Disease:    c.FormValue("disease"),
		Found:      similarity >= 80,
		Similarity: similarity,
		Indexes:    result,
	}
	found := 0
	if res.Found {
		found = 1
	}
	_, err = db.Exec("INSERT INTO hasil_prediksi (tanggal, nama_pasien, penyakit_prediksi, status, kesamaan) VALUES (?, ?, ?, ?, ?);", res.Time, res.Name, res.Disease, found, res.Similarity)
	if err != nil {
		panic(err)
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
	c.Response().WriteHeader(http.StatusOK)
	return c.JSON(http.StatusOK, res)
}

func main() {
	// initialize database
	createDB()
	// initializing, open localhost:3000 to check
	e := echo.New()
	e.POST("/api/match", match)

	e.Logger.Fatal(e.Start(":1323"))
}
