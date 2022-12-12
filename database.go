package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
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

func saveViolates2Txt() {
	//打印违纪名单
	outputTemplate := `
	<class>    <date>
	应交:%d		实交:%d


	班级名单:
	%s


	违纪名单:
	%s
	`
	outputTemplate = strings.Replace(outputTemplate, "<class>", MailFetchConfig.className, 1)
	outputTemplate = strings.Replace(outputTemplate, "<date>", time.Now().Format(time.RFC1123Z), 1)
	strAll := strings.Join(MailFetchConfig.stuLists, "    ")
	strViolate := strings.Join(MailFetchConfig.VIOLATELIST, "    ")

	outputText := fmt.Sprintf(outputTemplate, len(MailFetchConfig.stuLists),
		len(MailFetchConfig.stuLists)-len(MailFetchConfig.VIOLATELIST),
		strAll, strViolate)
	fmt.Print(outputText)

	file, _ := os.Create(path.Join(MailFetchConfig.rootPath, "违纪统计.txt"))
	defer file.Close()

	io.WriteString(file, outputText)
}

func recordLogs() {
	database_file := "./data.db"
	createDatabaseIfNeeded(database_file)
	saveViolates2Txt()
	saveViolates2Sqlite(database_file)
}

//Remove name from VIOLATELIST
func removeName(stuName string) {
	for i, item := range MailFetchConfig.VIOLATELIST {
		if item == stuName {
			MailFetchConfig.VIOLATELIST = append(MailFetchConfig.VIOLATELIST[:i],
				MailFetchConfig.VIOLATELIST[i+1:]...)
			return
		}
	}
}
