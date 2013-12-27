package main

import (
<<<<<<< HEAD
//	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
=======
	"io/ioutil"
	"net/http"
	"html/template"
	"labix.org/v2/mgo"
	 "labix.org/v2/mgo/bson"
	"fmt"
>>>>>>> a8517194256faf748bf0b54b91b3c928c3c6ee19
)

type Page struct {
	Title string
<<<<<<< HEAD
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

=======
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

>>>>>>> a8517194256faf748bf0b54b91b3c928c3c6ee19
func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p, err := loadPage(title)

	if err != nil {
<<<<<<< HEAD
		http.Redirect(w, r, "/index", http.StatusFound)
=======
		http.Redirect(w, r, "/home", http.StatusFound)
>>>>>>> a8517194256faf748bf0b54b91b3c928c3c6ee19
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
<<<<<<< HEAD
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
=======
>>>>>>> a8517194256faf748bf0b54b91b3c928c3c6ee19
}

func createAccount(usr *User) (bool) {
	session, err := mgo.Dial("127.0.0.1:27017/")
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

// Checks if an account exists in the userbase.
func doesAccountExist(email string) (bool) {
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		panic(err)
	}

	result := User{}
	c := session.DB("test").C("users")
	
	err = c.Find(bson.M{"email": email}).One(&result)
	if err != nil {
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
		if doesAccountExist(usr.Email) { 
			http.Redirect(w, r, "/account-exists", http.StatusFound)
		} else {
			ok := createAccount(usr)
			if ok {
				http.Redirect(w, r, "/success", http.StatusFound)
			} else {
				http.Redirect(w, r, "/failed", http.StatusFound)
			}
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
