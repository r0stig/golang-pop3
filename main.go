package main

import (
	"net"
	"os"
	"fmt"
	"bufio"
	"strings"
	"strconv"
	"errors"
)

const (
	host = "experimental.zapto.org"
)
const (
	STATE_UNAUTHORIZED = 1
	STATE_TRANSACTION = 2
	STATE_UPDATE = 3
)

func main() {
	connectDatabase()
	// Server IP and PORT
	service := "192.168.1.4:3110"
	tcpAddr, err := net.ResolveTCPAddr("ip4", service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error.. %s", err.Error())
	}
	
	listener, err := net.ListenTCP("tcp", tcpAddr)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error.. %s", err.Error())
	}

	fmt.Fprintf(os.Stdout, "Server listening, host: %s\n", tcpAddr.String() );
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// run as goroutine
		go handleClient(conn)
	}
	
	fmt.Println(tcpAddr)
	fmt.Println("hejsan")
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	
	var (
		tmp_user = ""
		eol = "\r\n"
		state = 1
	)
	reader := bufio.NewReader(conn)
	//writer := bufio.NewWriter(conn)
	
	
	// State
	// 1 = Unauthorized
	// 2 = Transaction mode
	// 3 = update mode
	
	// First welcome the new connection
	fmt.Fprintf(conn, "+OK simple POP3 server %s powered by Go" + eol, host)
	//nr, _ := writer.WriteString("+OK simple POP3 server (" + host + ") powered by Go")

	//writer.Flush()	// Måste använda flush eftersom inte buffern  fylls med ett meddelande...

	for {
		// Reads a line from the client
		raw_line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error!!" + err.Error())
			return
		}
		
		// Parses the command
		cmd, args := getCommand(raw_line)
		
		fmt.Println(">" + cmd + "<")
		arg, _ := getSafeArg(args, 0)
		fmt.Println("line>" + arg + "<")
		fmt.Println(raw_line)
		
		if cmd == "USER" && state == STATE_UNAUTHORIZED {
			tmp_user, _ = getSafeArg(args, 0)
			if userExists(tmp_user) {
				fmt.Fprintf(conn, "+OK" + eol)
				fmt.Println(">+OK" + eol)
			} else {
				fmt.Fprintf(conn, "-ERR The user %s doesn't belong here!" + eol, tmp_user)
				fmt.Println(">-ERR The user " + tmp_user + " doesn't belong here!" + eol)
			}
		} else if cmd == "PASS" && state == STATE_UNAUTHORIZED {
			pass, _ := getSafeArg(args, 0)
			if authUser(tmp_user, pass) {
				fmt.Fprintf(conn, "+OK User signed in" + eol)
				fmt.Println(">+OK User signed in" + eol)
				
				state = 2
				
			} else {
				fmt.Fprintf(conn, "-ERR Password incorrect!" + eol)
				fmt.Println(">-ERR Password incorrect!" + eol)
			}
		} else if cmd == "STAT" && state == STATE_TRANSACTION {
			nr_messages, size_messages := getStat(tmp_user)
			fmt.Fprintf(conn, "+OK " + strconv.Itoa(nr_messages) + " " + strconv.Itoa(size_messages) + eol)
			fmt.Println(">+OK " + strconv.Itoa(nr_messages) + " " + strconv.Itoa(size_messages) + eol)
		
		} else if cmd == "LIST" && state == STATE_TRANSACTION {
			fmt.Println("List accepted")
			nr, tot_size, Message_head := getList(tmp_user)
			fmt.Fprintf(conn, "+OK %d messages (%d octets)\r\n", nr, tot_size)
			// Print all messages
			for _, val := range Message_head {
				fmt.Fprintf(conn, "%d %d\r\n", val.Id, val.Size)
			}
			// Ending
			fmt.Fprintf(conn, ".\r\n")
		
		} else if cmd == "UIDL" && state == STATE_TRANSACTION  {
		
			// Retreive one message but don't delete it from the server..
			//message, size, _ := getMessage(tmp_user, 1)
			//fmt.Fprintf(conn, "+OK " + strconv.Itoa(size) + " octets" + eol)
			//fmt.Fprintf(conn, message.message + eol + "." + eol)
			fmt.Fprintf(conn, "-ERR Command not implemented" + eol)
			
		} else if cmd == "TOP" && state == STATE_TRANSACTION {
			arg, _ := getSafeArg(args, 0)
			nr, _ := strconv.Atoi(arg)
			headers := getTop(tmp_user, nr)

			fmt.Fprintf(conn, "+OK Top message followes" + eol)
			fmt.Fprintf(conn, headers + eol + eol + "." + eol)
			
		} else if cmd == "RETR" && state == STATE_TRANSACTION {
			arg, _ := getSafeArg(args, 0)
			nr, _ := strconv.Atoi(arg)
			message, size, _ := getMessage(tmp_user, nr)
			fmt.Fprintf(conn, "+OK " + strconv.Itoa(size) + " octets" + eol)
			fmt.Fprintf(conn, message.Headers + eol + eol + message.Message + eol + "." + eol)
			
			fmt.Println(">+OK " + strconv.Itoa(size) + " octets")
			fmt.Println(message.Message + eol + "." + eol)
			
		} else if cmd == "DELE" && state == STATE_TRANSACTION {
			arg, _ := getSafeArg(args, 0)
			nr, _ := strconv.Atoi(arg)
			deleteMessage(tmp_user, nr)
			fmt.Fprintf(conn, "+OK" + eol)
		} else if cmd == "QUIT" {
			return
		}
	}
}
// cuts the line into command and arguments
func getCommand(line string) (string, []string) {
	line = strings.Trim(line, "\r \n")
	cmd := strings.Split(line, " ")
	return cmd[0], cmd[1:]
}
func getSafeArg(args []string, nr int) (string, error) {
	if nr < len(args) {
		return args[nr], nil
	} 
	return "", errors.New("Out of range")
}