package main

import (
//	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
)

type Page struct {
	Title string
	Body  template.HTML
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p, err := loadPage(title)

	if err != nil {
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}

