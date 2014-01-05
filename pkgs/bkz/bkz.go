package bkz

import (
	
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

)

type Book struct {

	Title string
	Author string
	ISBN string
	Genre string
	Id string

}

// Creates an account and adds it to the Database
func CreateBook(book *Book) bool {

	session, err := mgo.Dial("127.0.0.1:27017/")
	if err != nil {
		fmt.Println(err)
		return false
	}

	c := session.DB("library").C("users")
	result := Book{}
	err = c.Find(bson.M{"id": book.Id}).One(&result)
	if result.Id != "" {
		// return true because book is present in the database
		// and we can say, "it's been added" without causing errors
		return true
	}
	fmt.Println(book)

	err = c.Insert(*book)

	if err != nil {
		return false
	}
	return true
}
