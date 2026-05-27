package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	path := "data/lifesim.db"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Println("open:", err)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, type, status FROM jobs ORDER BY rowid DESC LIMIT 8`)
	if err != nil {
		fmt.Println("query:", err)
		return
	}
	defer rows.Close()
	fmt.Println("jobs in", path)
	for rows.Next() {
		var id, t, s string
		_ = rows.Scan(&id, &t, &s)
		fmt.Println(" ", id, t, s)
	}
	var c int
	_ = db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE id = ?`, "e4b9ad97-bc32-4a71-b420-e8e29f513f6a").Scan(&c)
	fmt.Println("target job count:", c)
}
