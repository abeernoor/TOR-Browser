package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const httpPort = ":7000"
const otherPort = ":7001"
const entryRelay = "EN"
const middleRelay = "I"
const exitRelay = "EX"

type torData struct {
	MidRelay  string
	ExitRelay string
	URL       string
	pageBody  string
}

func sendAliveMessages(conn net.Conn) {
	for {
		d, _ := time.ParseDuration("2s")
		time.Sleep(d)
		conn.Write([]byte("ALIVE"))
	}
}

func clientHandler(rw http.ResponseWriter, r *http.Request) {
	//Get list from directory server
	//Generate a random path of relays
	//Send to entry relay: address of mid and exit relay and URL
	//Wait for response

}

func listenAsARelay(relayType string) {
	line, _ := net.Listen("tcp", otherPort)
	for {
		con, _ := line.Accept()
		go handleRelayConnection(con, relayType)
	}
}

func handleRelayConnection(conn net.Conn, relayType string) {
	//Forward to next relay the info
	//or get page from server if you are exit relay
	buffer := make([]byte, 500, 1024)
	conn.Read(buffer)
	tData := torData{}
	json.Unmarshal(buffer, &tData)
	if relayType == entryRelay {
		midRelayConn, _ := net.Dial("tcp", tData.MidRelay)
		tData.MidRelay = ""
		newBuffer, _ := json.Marshal(tData)
		midRelayConn.Write(newBuffer)
		buffer = make([]byte, 500, 1024)
		midRelayConn.Read(buffer)
	} else if relayType == middleRelay {
		exitRelayConn, _ := net.Dial("tcp", tData.ExitRelay)
		tData.ExitRelay = ""
		newBuffer, _ := json.Marshal(tData)
		exitRelayConn.Write(newBuffer)
		buffer = make([]byte, 500, 1024)
		exitRelayConn.Read(buffer)
	} else {
		response, _ := http.Get(tData.URL)
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		tData.pageBody = string(responseBody)
		buffer, _ = json.Marshal(tData)
	}
	conn.Write(buffer)
	conn.Close()
}

func main() {
	fmt.Println("Do you want to participate as a relay?")
	//Take input
	consoleReader := bufio.NewReader(os.Stdin)
	input, _ := consoleReader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.ToLower(input) == "yes" || strings.ToLower(input) == "y" {
		fmt.Print("1. Entry Relay\n2. Middle Relay\n3. Exit Relay\nEnter your choice: ")
		choice, _ := consoleReader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		msg := ""
		switch choice {
		case "1":
			msg = entryRelay
			break
		case "2":
			msg = middleRelay
			break
		case "3":
			msg = exitRelay
			break
		default:
			break
		}
		if msg != "" {
			conn, _ := net.Dial("tcp", "localhost:6000")
			fmt.Fprintf(conn, "%s", msg)
			go sendAliveMessages(conn)
			go listenAsARelay(msg)
		}
	}
	http.HandleFunc("/fastor/", clientHandler)
	http.ListenAndServe(httpPort, nil)
}
