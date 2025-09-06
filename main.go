package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var addr = "localhost:8080"

type MovieDB struct {
	db *gorm.DB
	ctx context.Context
}

type Settings struct {
  gorm.Model
  LibPath  string
}

type Movies struct {
  gorm.Model
  Title    string
  Year     int
  Cover    string
}

func (m MovieDB) scanLibrary(w http.ResponseWriter, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	
	for _, e := range entries {
		fmt.Println(e.Name())
		err = gorm.G[Movies](m.db).Create(m.ctx, &Movies{Title: e.Name()})
		if err != nil {
			fmt.Fprintln(w, err)
		}
	}
}

func main()  {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	ctx := context.Background()
	
	movieDB := MovieDB{db: db, ctx: ctx}

	db.AutoMigrate(&Settings{}, &Movies{})

	http.HandleFunc("/lib", func(w http.ResponseWriter, r *http.Request) {
		movieDB.scanLibrary(w, "./")
	})

	s := &http.Server{
		Addr:           addr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("Starting server on %s\n", addr)
	log.Fatal(s.ListenAndServe())
}