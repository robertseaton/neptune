package main

import (
	"crypto/sha256"
	"encoding/hex"

	"io/ioutil"
	"net/http"
	"html/template"
	"labix.org/v2/mgo"
	 "labix.org/v2/mgo/bson"
	"fmt"
)

type Page struct {
	Title string
	Body  template.HTML
}

func SHA(str string)(string){

	var bytes []byte
	//var n int32
    //binary.Read(rand.Reader, binary.LittleEndian, &n)

	copy(bytes[:], str)								// convert string to bytes
    h := sha256.New()								// new sha256 object
    h.Write(bytes)										// data is now converted to hex
	code := h.Sum(nil)								// code is now the hex sum
	codestr := hex.EncodeToString(code)	// converts hex to string
    return codestr
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: template.HTML(body)}, nil
}


type User struct {
	Email string
	Password string
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
	pass := r.FormValue("pwd")
	usr.Password = pass
	
	if len(usr.Password) > 0 {
		usr.Password = SHA(pass)
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
