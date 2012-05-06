package main

import (
	"fmt"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"

)
type Message struct {
	Headers string
	Message string
	Subject string
}
type User struct {
	Username string
	Password string
	Messages []Message
}

type Message_head struct {
	Id int
	Size int
}

var (
	session *mgo.Session
	collection *mgo.Collection
)
var (
	eol = "\r\n"
)

// Connect to the database
func connectDatabase() {
	sess, err := mgo.Dial("192.168.1.4")
	if err != nil {
		fmt.Println("Couldn't connect to the MongoDB")
	}
	
	collection = sess.DB("robmail").C("users")
	session = sess
	
	fmt.Println( userExists("test4") )
	fmt.Println( authUser("test4", "test4") )
	//fmt.Println(getStat("test4"))
	//fmt.Println(getList("test4"))
	//fmt.Println(getMessage("test4", 1))
	//fmt.Println(getTop("test4", 2))
}
// Disconnect from the database
func disconnectDatabase() {
	session.Close()
}

// Checks if a user exists or not
func userExists(user string) bool {
	cnt, _ := collection.Find(bson.M{"username": user }).Count()
	return cnt == 1
}
// Checks if the login info is correct
func authUser(user string, pass string) bool {
	cnt, _ := collection.Find(bson.M{"username": user , "password": pass}).Count()
	return cnt == 1
}
// Returns the number of messages in the maildrop and
// the size in bytes of the maildrop
func getStat(user string) (int, int) {
	/*test := User{ Username: "test1223", Password: "pass1223" }
	errore := collection.Insert(&test)
	if errore != nil {
		fmt.Println(errore.Error())
	}*/
	result := User{}
	err := collection.Find(bson.M{"username": user}).One(&result)
	if err != nil {
		fmt.Println("Error:" + err.Error())
		// If the document cannot be found, error occurs here
	}

	fmt.Println( result )
	fmt.Println( len(result.Messages) )
	// To count the total octets..
	var (
		sum = 0
	)
	// Count how many letters there are in all the headers and messages
	for _, v := range result.Messages {
		sum = sum + len(v.Headers) + len(v.Message) // headers_cnt+message_cnt 
	}
	
	// return the count and the size in octets (bytes)
	return len(result.Messages), sum*8
}

// Returns all message heads in maildrop if no argument
// or the message head for the mail id
// in format:
// 
func getList(user string) (int, int,  []Message_head) {
	result := User{}
	err := collection.Find(bson.M{"username": user}).One(&result)
	if err != nil {
		fmt.Println("Error:" + err.Error())
		// If the document cannot be found, error occurs here
	}

	fmt.Println( result )
	fmt.Println( len(result.Messages) )
	// To count the total octets..
	var (
		sum = 0
		messages []Message_head
	)
	// Add all messages into a header struct
	for i, v := range result.Messages {
		size := (len(v.Headers) + len(v.Message))*8
		m := Message_head{ Id: i+1, 
									Size:  size}
		messages = append(messages, m)
		sum = sum + len(v.Headers) + len(v.Message) // headers_cnt+message_cnt 
	}
	fmt.Println(messages)
	return len(messages), sum*8, messages
	/*
	m := Message_head{
		Id: 1, 
		Size: 180}
	messages := []Message_head{ m }
	return 1, 180, messages*/
}
//func getListN(
// Returns the message of the id
func getMessage(user string, id int) (Message, int, error) {
	
	result := User{}
	err := collection.Find(bson.M{"username": user}).One(&result)
	if err != nil {
		fmt.Println("Error:" + err.Error())
		// If the document cannot be found, error occurs here
	}

	fmt.Println( result )
	fmt.Println( len(result.Messages) )

	// Get the specified message
	i := id-1

	size := len(result.Messages[i].Message)*8
	return result.Messages[i], size, nil

		
	
	/*
	message := Message{
		Headers: "From: test@test.se"+ eol + 
"Subject: This is the subject"+ eol + 
"To: pr_125@hotmail.com"+ eol + 
"X-PHP-Originating-Script: 0:smtp-server-test.php",
		Subject: "This is the subject",
		Message: "This is the message"}
	return message, len(message.Message)*8, nil
	*/
}

//						TO-DO:	int arg1, int arg2
func getTop(user string, id int) string {
	result := User{}
	err := collection.Find(bson.M{"username": user}).One(&result)
	if err != nil {
		fmt.Println("Error:" + err.Error())
		// If the document cannot be found, error occurs here
	}

	// Get the specified message
	return result.Messages[id-1].Headers
	/*
	m := Message {
		Headers: "From: test@test.se"+ eol + 
"Subject: This is the subject"+ eol + 
"To: pr_125@hotmail.com"+ eol + 
"X-PHP-Originating-Script: 0:smtp-server-test.php",
		Message: "This is the message"}
	// Depending on the arg1 and arg2 it returns different
	// 1 0 means return header but no body
	return m.Headers*/
}
// Marks a message for deletion, the message is removed
// in the UPDATE-state
func deleteMessage(user string, id int) {
	i := id-1
	err := collection.Update(bson.M{ "username": user}, bson.M{ "$unset" : bson.M{ "messages" : i } })
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}