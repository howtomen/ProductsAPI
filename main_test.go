package main_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/howtomen/productsapi"
)

var (
	a           main.App
	DB_username string
	DB_pass     string
	DB_name     string
)

func TestMain(m *testing.M) {
	//Get Database info from flags.
	flag.StringVar(&DB_username, "username", "postgres", "Username for PSQL DB")
	flag.StringVar(&DB_pass, "password", "randpasssecure123", "Password for PSQL DB")
	flag.StringVar(&DB_name, "db", "Products-1", "Name of the PSQL DB")
	flag.Parse()

	a.Initialize(
		DB_username,
		DB_pass,
		DB_name)

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func TestEmptyTable (t *testing.T) {
	clearTable()
	req,_ := http.NewRequest("GET","/products",nil)
	response := executeRequest(req)
	checkResponseCode(t,http.StatusOK,response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected Empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr,req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected responsecode %d. Got %d\n", expected, actual)
	}
}

func TestNonexistentProduct(t *testing.T) {
	clearTable()

	req,_ := http.NewRequest("GET","/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t,http.StatusNotFound,response.Code)

	m := map[string]string{}

	json.Unmarshal(response.Body.Bytes(), &m)

	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])	
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	jsonStr := []byte(`{"name":"test product", "price": 11.22}`)
	req,_ := http.NewRequest("POST","/product",bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type","application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected Product name to be 'test product'. Got %v", m["name"])
	}

	if m["price"] != 11.22 {
        t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
    }

	//JSON unmarshalling converts ints to floats so we compare to 1.0
	if m["id"] != 1.0 {
        t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
    }
}
func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req,_ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
        a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
    }
}

func TestUpdateProduct(t *testing.T) {

	clearTable()
	addProducts(1)

	req,_ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var ogProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &ogProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req,_ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != ogProduct["id"] {
		t.Errorf("Expected the id to stay the same (%v). Got %v", ogProduct["id"], m["id"])
	}

	if m["name"] == ogProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to 'test product -updated name'. Got %v", ogProduct["name"], m["name"])
	}

	if m["price"] == ogProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '11.22'. Got %v", ogProduct["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req,_ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req,_ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}