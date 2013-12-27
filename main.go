package main

import (
	"crypto/sha256"
	"encoding/hex"

	"fmt"
	"html/template"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"

	"time"

	"neptune/user"
)

type Page struct {
	Title string
	Body  template.HTML
}

type User struct {
	Email    string
	Password string
	cookie http.Cookie
}

func SHA(str string) string {

	var bytes []byte
	copy(bytes[:], str)                 // convert string to bytes
	h := sha256.New()                   // new sha256 object
	h.Write(bytes)                      // data is now converted to hex
	code := h.Sum(nil)                  // code is now the hex sum
	codestr := hex.EncodeToString(code) // converts hex to string
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

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	// check user status loged-in/not
	p, err := loadPage(title)

	if err != nil {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

func createAccount(usr *User) bool {
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
func doesAccountExist(email string) bool {
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

func checkCredentials(email string, password string) bool {
	password = SHA(password)
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

	if result.Password == password && result.Email == email {
		return true
	}

	return false
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	usr := new(User)
	usr.Email = r.FormValue("email")
	usr.Password = r.FormValue("pwd")

	if len(usr.Password) > 0 {
		ok := checkCredentials(usr.Email, usr.Password)
		if ok {
			user.CreateUserFile(usr.Email)
			s := "/" + usr.Email
			fmt.Println(s)
			http.Redirect(w, r, s, http.StatusFound)
		} else {
			http.Redirect(w, r, "/login-failed", http.StatusFound)
		}
	} else {
		viewHandler(w, r)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	const minPasswordLength = 4

	usr := new(User)
	usr.Email = r.FormValue("email")
	pass := r.FormValue("pwd")

	if len(pass) > 0 {
		usr.Password = SHA(pass)
		if doesAccountExist(usr.Email) {
			http.Redirect(w, r, "/account-exists", http.StatusFound)
		} else {
			ok := createAccount(usr)
			if ok {
				expire := time.Now().AddDate(0, 0, 1)
				// TODO actually make a cookie!!!!
				cookie := http.Cookie{"test", "tcookie", "/", "http://localhost:8080/", expire, expire.Format(time.UnixDate), 86400, true, true, "test=tcookie", []string{"test=tcookie"}}
				http.SetCookie(w, &cookie)
				usr.cookie = cookie
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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
