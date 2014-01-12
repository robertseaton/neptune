package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/robertseaton/neptune/pkgs/bkz"
	"github.com/robertseaton/neptune/pkgs/codify"
	"github.com/robertseaton/neptune/pkgs/cookies"
	"github.com/robertseaton/neptune/pkgs/user"
)

type Page struct {
	Title    string
	Body     template.HTML
	UserData template.HTML
}

// Loads a page for use
func loadPage(title string, r *http.Request) (*Page, error) {

	filename, option, usr := user.LoadUserInfo(title, r)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &Page{Title: title, Body: template.HTML(body), UserData: (template.HTML(usr) + template.HTML(option))}, nil
}

// Shows a particular page
func viewHandler(w http.ResponseWriter, r *http.Request) {

	title := r.URL.Path[len("/"):]
	p, err := loadPage(title, r)

	if err != nil && !cookies.IsLoggedIn(r) {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	} else if err != nil {
		http.Redirect(w, r, "/login-succeeded", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("view.html")
	t.Execute(w, p)
}

// Handles the users loggin and gives them a cookie for doing so
func loginHandler(w http.ResponseWriter, r *http.Request) {

	usr := new(user.User)
	usr.Email = r.FormValue("email")
	pass := r.FormValue("pwd")

	if len(pass) > 0 {
		usr.Password = codify.SHA(pass)
		ok := user.CheckCredentials(usr.Email, usr.Password)
		if ok {
			user.CreateUserFile(usr.Email)
			cookie := cookies.LoginCookie(usr.Email)
			http.SetCookie(w, &cookie)
			usr.SessionID = cookie.Value
			_ = user.UpdateUser(usr)
			http.Redirect(w, r, "/login-succeeded", http.StatusFound)
		} else {
			http.Redirect(w, r, "/login-failed", http.StatusFound)
		}
	} else {
		viewHandler(w, r)
	}
}

// Registers the new user
func registerHandler(w http.ResponseWriter, r *http.Request) {

	usr := new(user.User)
	usr.Email = r.FormValue("email")
	pass := r.FormValue("pwd")

	if len(pass) > 0 {
		usr.Password = codify.SHA(pass)
		if user.DoesAccountExist(usr.Email) {
			http.Redirect(w, r, "/account-exists", http.StatusFound)
		} else {
			ok := user.CreateAccount(usr)
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

// Logs out the user, removes their cookie from the database
func logoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("SessionID")
	if err != nil {
		fmt.Println(err)
		return
	}

	result := new(user.User)
	sessionID := cookie.Value
	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		return
	}

	c := session.DB("test").C("users")
	c.Find(bson.M{"sessionid": sessionID}).One(&result)
	result.SessionID = result.Email + ":" + codify.SHA(result.SessionID+strconv.Itoa(rand.Intn(100000000)))
	err = c.Update(bson.M{"email": result.Email}, result)

	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/home", http.StatusFound)

}

// Adds a new book to the database/user
func bookHandler(w http.ResponseWriter, r *http.Request) {

	//	cookie, err := r.Cookie("SessionID")
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}

	//	sessionID := cookie.Value

	//	z := strings.Split(sessionID, ":")
	//	email := z[0]

	book := new(bkz.Book)
	book.Title = r.FormValue("book")
	book.Author = r.FormValue("author")
	book.ISBN = r.FormValue("isbn")
	book.Genre = r.FormValue("genre")

	fmt.Println(book)

	if len(book.Title) > 0 {
		ok := bkz.CreateBook(book)
		if ok {
			// TODO add a better succuess page
			http.Redirect(w, r, "/success", http.StatusFound)
		} else {
			// TODO add another error page
			http.Redirect(w, r, "/failed", http.StatusFound)
		}
	} else {
		viewHandler(w, r)
	}
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/add-book", bookHandler)
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
