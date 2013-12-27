package main

import (
	"io/ioutil"
	"net/http"
	"html/template"
	"labix.org/v2/mgo"
	"fmt"
)

type Page struct {
	Title string
	Body template.HTML
}

type User struct {
	Email string
	Password string
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
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

func createAccount(usr *User) (bool) {
	session, err := mgo.Dial("localhost:27017/")
	if err != nil {
		fmt.Println(err)
		return false
	}

	c := session.DB("test").C("users")
	err = c.Insert(usr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	const minPasswordLength = 4

	usr := new(User)
	usr.Email = r.FormValue("email")
	usr.Password = r.FormValue("pwd")

	if len(usr.Password) > 0 {
		ok := createAccount(usr)
		if ok {
			http.Redirect(w, r, "/success", http.StatusFound)
		} else {
			http.Redirect(w, r, "/failed", http.StatusFound)
		}
		
	} else {
		viewHandler(w, r)
	}
}

func main() {
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
