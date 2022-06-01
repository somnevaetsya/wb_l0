package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Data struct {
	Id   uint   `gorm:"primaryKey;"`
	Json string `gorm:"not null;"`
}

func initDB() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open("host=localhost user=user password=userwb dbname=db_wb port=5432"))
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Data{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func fillCash(db *gorm.DB) error {
	var currentData []Data
	result := db.Find(&currentData)
	for _, item := range currentData {
		Cash[item.Id] = item.Json
	}
	return result.Error
}

func saveData(toSave Data, db *gorm.DB, currentData *map[uint]string) error {
	fmt.Println(toSave)
	err := db.Create(&toSave).Error
	fmt.Println(toSave)
	if err != nil {
		return err
	}
	(*currentData)[toSave.Id] = toSave.Json
	return nil
}

func sendData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method: ", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("template.gtpl")
		err := t.Execute(w, nil)
		if err != nil {
			return
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return
		}
		t, _ := template.ParseFiles("template.gtpl")
		err = t.Execute(w, nil)
		if err != nil {
			return
		}
		currentId, err := strconv.ParseUint(strings.Join(r.Form["id"], ""), 10, 32)
		fmt.Println("data", currentId)
		_, err = w.Write([]byte(Cash[uint(currentId)]))
		if err != nil {
			return
		}
	}
}

var Cash = map[uint]string{}

func main() {
	db, err := initDB()
	if err != nil {
		fmt.Println("error in initDB")
		return
	}
	err = fillCash(db)
	if err != nil {
		fmt.Println("error in filling cash")
		return
	}

	sc, err := stan.Connect("test-cluster", "client")
	if err != nil {
		fmt.Println(err)
		return
	}
	sub, _ := sc.Subscribe("data", func(m *stan.Msg) {
		err := saveData(Data{Json: string(m.Data[:])}, db, &Cash)
		if err != nil {
			fmt.Println("error in saveData", err)
		}
	})

	http.HandleFunc("/data", sendData)
	fmt.Println("starting server at: 8080")
	http.ListenAndServe(":8080", nil)

	sub.Unsubscribe()
	sc.Close()
}
