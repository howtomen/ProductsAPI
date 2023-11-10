package main

import (
	"flag"
)

var (
	DB_username string
	DB_pass     string
	DB_name     string
)

func main() {
	//Get Database info from flags.
	flag.StringVar(&DB_username, "username", "postgres", "Username for PSQL DB")
	flag.StringVar(&DB_pass, "password", "randpasssecure123", "Password for PSQL DB")
	flag.StringVar(&DB_name, "db", "Products-1", "Name of the PSQL DB")
	flag.Parse()

	a := App{}
	a.Initialize(
		DB_username,
		DB_pass,
		DB_name)

	a.Run(":8010")
}
