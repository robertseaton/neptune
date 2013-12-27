package user

import(
	"os"
	"fmt"
)

// TODO add error checking

func CreateUserFile(usrName string){

	file, err := os.Create("accounts/" + usrName + ".txt")			// creates a file with that usrName
	if err != nil {	fmt.Printf("error createUserFile FIX")  }

	s := "You've successfully managed to log in! Enjoy your stay on Neptune. The current temperature outside the shuttle is -356°F. <br> <br>"

	file.WriteString(s)

}

func ReadUserFile(usrName string)(file *os.File){

	file, err := os.Open(usrName)
	if err != nil {	fmt.Printf("error readUserFile FIX")  }

	return file

}

func AppendUserFile(usrName string){

	file, err:= os.OpenFile(usrName, os.O_WRONLY, 0666)
	if err != nil {	fmt.Printf("error appendUserFile FIX")  }

	s := "nothing"

	file.WriteString(s)
	

}
