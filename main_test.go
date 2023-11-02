package main_test

import (
	_ "bytes"
	_ "encoding/json"
	"flag"

	"log"
	_ "net/http"
	_ "net/http/httptest"
	"os"
	_ "strconv"
	"testing"
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
	flag.StringVar(&DB_pass, "password", "password123", "Password for PSQL DB")
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
