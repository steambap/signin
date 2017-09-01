package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var router *gin.Engine

func TestGetDailyLog(t *testing.T) {
	res, err := runRequest(http.MethodGet, "/log?date=2017-6-21", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check to see if the response was what you expected
	if res.Code != http.StatusOK {
		t.Fatalf("Expected status %d ; got %d\n", http.StatusOK, res.Code)
	}
}

func TestWrongDate(t *testing.T) {
	res, err := runRequest(http.MethodGet, "/log?date=2017", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check to see if the response was what you expected
	if res.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d ; got %d\n", http.StatusBadRequest, res.Code)
	}
}

func TestWrongLoc(t *testing.T) {
	res, err := runRequest(http.MethodGet, "/log?date=2017-6-21&loc=1010", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check to see if the response was what you expected
	if res.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d ; got %d\n", http.StatusBadRequest, res.Code)
	}
}

func TestPostDailyLog(t *testing.T) {
	body := Body{Names: []string{"111"}, Tags: []string{"222"}}
	jsonValue, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	res, err := runRequest(http.MethodPost, "/log?date=2017-6-21", bytes.NewBuffer(jsonValue))
	if err != nil {
		t.Fatal(err)
	}

	// Check to see if the response was what you expected
	if res.Code != http.StatusOK {
		t.Fatalf("Expected status %d ; got %d\n", http.StatusOK, res.Code)
	}
}

func TestGetYear(t *testing.T) {
	res, err := runRequest(http.MethodGet, "/log/year/2017?loc=11", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check to see if the response was what you expected
	if res.Code != http.StatusOK {
		t.Fatalf("Expected status %d ; got %d\n", http.StatusOK, res.Code)
	}
}

func TestScanBucket(t *testing.T) {
	res, err := runRequest(http.MethodGet, "/loc/11", nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Expected status %d ; got %d \n", http.StatusOK, res.Code)
	}
}

func TestMain(m *testing.M) {
	// Switch to test mode so you don't get such noisy output
	gin.SetMode(gin.TestMode)
	db := startBoltDb("test.db")
	defer db.Close()

	env := &Env{db: db}

	router = gin.Default()

	router.GET("/log", dateFilter, locationFilter, env.getDaily)
	router.POST("/log", dateFilter, locationFilter, marshalBody, env.putDaily)
	router.GET("/log/year/:num", yearFilter, locationFilter, env.getYear)

	router.GET("/loc/:loc", locationParamFilter, env.scanBucket)

	os.Exit(m.Run())
}

func runRequest(method string, urlStr string, body io.Reader) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder so you can inspect the response
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	return w, nil
}
