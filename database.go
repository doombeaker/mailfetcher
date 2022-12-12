package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
)

func createDatabaseIfNeeded(DATABASE_FILE string) {
	if _, err := os.Stat(DATABASE_FILE); errors.Is(err, os.ErrNotExist) {
		log.Println(DATABASE_FILE, "CREATED")
		file, err := os.Create(DATABASE_FILE)
		if err != nil {
			log.Fatal("os.Create:", err)
		}
		file.Close()

		db, err := sql.Open("sqlite3", DATABASE_FILE)
		defer db.Close()
		if err != nil {
			log.Fatal("sql.Open:", err)
		}

		//time.Sleep(time.Millisecond * 500)
		sql_to_exec := `CREATE TABLE violate (
			clsname CHAR (40)     NOT NULL,
			stuname VARCHAR (40)  NOT NULL,
			others  VARCHAR (200),
			date    DATETIME (40) NOT NULL,
			vid     INTEGER       PRIMARY KEY AUTOINCREMENT
		);		
		`
		stmt, err := db.Prepare(sql_to_exec)
		if err != nil {
			log.Fatal("db.Prepare:", err)
		} else {
			_, err := stmt.Exec()
			if err != nil {
				log.Fatal("stmt.Exec(): ", err)
			}
		}

		return
	}
	log.Println(DATABASE_FILE, "exists, USE IT")
}

//SaveViolates2DB Saves violated records
func saveViolates2Sqlite(DATABASE_FILE string) {

	db, err := sql.Open("sqlite3", DATABASE_FILE)
	defer db.Close()

	if err != nil {
		log.Println(err)
	}

	for _, item := range MailFetchConfig.VIOLATELIST {
		stmt, err := db.Prepare(`INSERT INTO violate (clsname, stuname, date) VALUES (?, ?, datetime('now', 'localtime'))`)
		if err != nil {
			log.Println(err)
		} else {
			_, err := stmt.Exec(MailFetchConfig.className, item)
			if err != nil {
				log.Println(err)
			}
		}

	}
}
