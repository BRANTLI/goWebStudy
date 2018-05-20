package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = ":3306"
	DBUser  = "root"
	DBPass  = ""
	DBDbase = "cms"
	PORT    = ":8080"
)

var database *sql.DB

type Page struct {
	Title   string
	Content string
	Date    string
	GUID    string
	ID 		string
}

func (p Page) TruncatedText() string {
	chars := 0
	for i, _ := range p.Content {
		chars++
		if chars > 150 {
			return p.Content[:i] + ` ...`
		}
	}
	return p.Content
}

func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}
	pages, err := database.Query("SELECT title,content,date,id FROM pages ORDER BY ? DESC", "date")
	if err != nil {
		fmt.Fprintln(w, err.Error)
	}
	defer pages.Close()

	for pages.Next() {
		thisPage := Page{}
		err=pages.Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date,&thisPage.ID)
		if err != nil {
			log.Fatal(err)
		}
		thisPage.Content	= thisPage.TruncatedText()
		Pages = append(Pages, thisPage)
		log.Println(thisPage.ID, thisPage.Content)

	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT title,content,date FROM pages WHERE id=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return
	}
	// html := `<html><head><title>` + thisPage.Title + `</title></head><body><h1>` + thisPage.Title + `</h1><div>` + thisPage.Content + `</div></body></html>`

	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}

func main() {
	//dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, DBPass, DBDbase)
	//fmt.Println(dbConn)
	//db, err := sql.Open("mysql", dbConn)
	//if err != nil {
	//	log.Println("Couldn't connect!")
	//	log.Println(err.Error)
	//}
	db,err:=sql.Open("sqlite3","./cms/cms")
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err)
	}
	defer db.Close()
	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
	routes.HandleFunc("/", RedirIndex)
	routes.HandleFunc("/home", ServeIndex)
	http.Handle("/", routes)
	http.ListenAndServe(PORT, nil)

}
