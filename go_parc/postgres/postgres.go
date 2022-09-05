package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	DB_USER     = "fedorshmidt"
	DB_PASSWORD = "postgres"
	DB_NAME     = "chickslist"
)

func OpenDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	return db
}

func CloseDB(db *sql.DB) {
	db.Close()
}

func ChlistToDB(Nick, Date string, db *sql.DB) {

	fmt.Println("# Inserting values")

	var lastInsertId int
	err := db.QueryRow("INSERT INTO chicks(Nickname, Parsedate) VALUES($1,$2) returning UIN;", Nick, Date).Scan(&lastInsertId)
	checkErr(err)
	fmt.Println("last inserted id =", lastInsertId)

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
