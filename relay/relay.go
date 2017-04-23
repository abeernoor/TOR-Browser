package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const entryRelay = "EN"
const middleRelay = "I"
const exitRelay = "EX"

// Node is struct that stores the relay info
type Node struct {
	RelayType string
	IPandPort string
	PubKey    string
}

type torData struct {
	MidRelay  string
	ExitRelay string
	URL       string
	PageBody  string
}

var clientConnection net.Conn

func sendAliveMessages(conn net.Conn, sig <-chan os.Signal) {
	d, _ := time.ParseDuration("2s")
	for {
		select {
		case <-sig:
			log.Println("Closing...")
			conn.Close()
			clientConnection.Close()
			return
		case <-time.After(d):
			conn.Write([]byte("ALIVE\n"))
		}
	}
}

func clientHandler(rw http.ResponseWriter, r *http.Request) {
	//Get list from directory server
	//Generate a random path of relays
	//Send to entry relay: address of mid and exit relay and URL
	//Wait for response
	log.Println("Getting list")
	clientConnection.Write([]byte("GET_LIST\n"))
	buffer := make([]byte, 2048, 10240)
	n, err := clientConnection.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Recieved List: ", buffer)
	if buffer == nil || len(buffer) == 0 {
		rw.Write([]byte("<html><h2>ERROR! Not enough relays in TOR Network</h2></html>"))
		return
	}
	relayList := make([]Node, 0, 10)
	json.Unmarshal(buffer[:n], &relayList)
	fmt.Println(relayList)
	entryRelayList := make([]Node, 0, 10)
	middleRelayList := make([]Node, 0, 10)
	exitRelayList := make([]Node, 0, 10)
	for _, relay := range relayList {
		if relay.RelayType == entryRelay {
			entryRelayList = append(entryRelayList, relay)
		} else if relay.RelayType == middleRelay {
			middleRelayList = append(middleRelayList, relay)
		} else {
			exitRelayList = append(exitRelayList, relay)
		}
	}

	chEntryRelay := entryRelayList[rand.Intn(len(entryRelayList))]
	chMidRelay := middleRelayList[rand.Intn(len(middleRelayList))]
	chExitRelay := exitRelayList[rand.Intn(len(exitRelayList))]
	log.Println("Entry Relay chosen: ", chEntryRelay)
	log.Println("Middle Relay chosen: ", chMidRelay)
	log.Println("Exit Relay chosen: ", chExitRelay)

	tData := torData{MidRelay: chMidRelay.IPandPort, ExitRelay: chExitRelay.IPandPort, URL: "http://" + r.URL.Path[len("/fastor/"):], PageBody: ""}
	conn, _ := net.Dial("tcp", ":"+chEntryRelay.IPandPort)
	log.Println("Connected with Entry Relay")
	buffer, _ = json.Marshal(tData)
	//fmt.Println(buffer)
	conn.Write(buffer)
	log.Println("Sent data to Entry Relay")
	// newBuffer := make([]byte, 1024, 1024)
	// n, err = conn.Read(newBuffer)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// var size string
	// json.Unmarshal(newBuffer[:n], &size)
	// intsize, _ := strconv.Atoi(size)
	newBuffer := make([]byte, 102400, 102400)
	n, err = conn.Read(newBuffer)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Read webpage from entry relay")
	newTData := torData{}
	json.Unmarshal(newBuffer[:n], &newTData)
	//fmt.Println("Printing the fetched Page")
	//fmt.Println(newTData.PageBody)
	rw.Write([]byte(newTData.PageBody))
}

func listenAsARelay(relayType string) {
	defer clientConnection.Close()
	port := strconv.Itoa(rand.Intn(3000) + 4000)
	line, err := net.Listen("tcp", ":"+port)
	defer line.Close()
	for err != nil {
		port = strconv.Itoa(rand.Intn(3000) + 4000)
		line, err = net.Listen("tcp", ":"+port)
	}
	clientConnection.Write([]byte(port + "\n"))
	log.Println("Port sent to DS on which this relay is listening for other relay requests")

	for {
		con, _ := line.Accept()
		log.Println("Got a request from relay: ", con.RemoteAddr())
		go handleRelayConnection(con, relayType)

	}
}

func handleRelayConnection(conn net.Conn, relayType string) {
	//Forward to next relay the info
	//or get page from server if you are exit relay
	buffer := make([]byte, 1024, 10240)
	n, _ := conn.Read(buffer)
	log.Println("Read request from relay: ", conn.RemoteAddr())
	defer conn.Close()
	tData := torData{}
	json.Unmarshal(buffer[:n], &tData)
	n = 0
	if relayType == entryRelay {
		midRelayConn, _ := net.Dial("tcp", ":"+tData.MidRelay)
		tData.MidRelay = ""
		newBuffer, _ := json.Marshal(tData)
		log.Println("Forwarding request to Middle Relay")
		midRelayConn.Write(newBuffer)
		// clientreader := bufio.NewReader(midRelayConn)
		// nbuffer, _, _ := clientreader.ReadLine()
		// clientwriter := bufio.NewWriter(conn)
		// clientwriter.Write(nbuffer)
		// // conn.Write(nbuffer)
		// fmt.Println(len(nbuffer))
		// size := ""
		// json.Unmarshal(nbuffer, &size)
		// intsize, err := strconv.Atoi(strings.TrimSpace(size))
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println("incoming size", intsize)
		buffer = make([]byte, 102400, 102400)
		n, _ = midRelayConn.Read(buffer)
		log.Println("Recieved response from Middle Relay")
		//fmt.Println(buffer[:n])
	} else if relayType == middleRelay {
		exitRelayConn, _ := net.Dial("tcp", ":"+tData.ExitRelay)
		tData.ExitRelay = ""
		newBuffer, err := json.Marshal(tData)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Forwarding request to Exit Relay")
		exitRelayConn.Write(newBuffer)
		// clientreader := bufio.NewReader(exitRelayConn)
		// nbuffer, _, _ := clientreader.ReadLine()
		// conn.Write(nbuffer)
		// fmt.Println(len(nbuffer))
		// size := ""
		// json.Unmarshal(nbuffer, &size)
		// intsize, _ := strconv.Atoi(strings.TrimSpace(size))
		// clientwriter := bufio.NewWriter(conn)
		// clientwriter.Write(nbuffer)
		// // conn.Write(nbuffer)
		// // if err != nil {
		// // 	fmt.Println(err)
		// // }
		// fmt.Println("incoming size", intsize)
		buffer = make([]byte, 102400, 102400)
		n, _ = exitRelayConn.Read(buffer)
		log.Println("Recieved response from Exit Relay")
		//fmt.Println(buffer[:n])

	} else {
		log.Println("Fetching webpage", tData.URL)
		response, err := http.Get(tData.URL)
		if err != nil {
			log.Fatal(err)
		}
		responseBody, er := ioutil.ReadAll(response.Body)
		if er != nil {
			log.Fatal(er)
		}
		// fmt.Println(string(responseBody))
		response.Body.Close()
		tData.PageBody = string(responseBody)
		buffer, _ = json.Marshal(tData)
		//size := len(buffer)
		//str := strconv.Itoa(size) + "\n"
		//fmt.Println("size string ", str)
		//buffersize, _ := json.Marshal(str)
		// fmt.Println("before : ", len(buffersize), cap(buffersize))
		// buffersize = append([]byte(nil), buffersize[:len(buffersize)]...)
		// fmt.Println("after : ", len(buffersize), cap(buffersize))
		// clientwriter := bufio.NewWriter(conn)
		// clientwriter.Write(buffersize)
		// _, err = conn.Write(buffersize)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		//fmt.Println(size, buffersize)
	}
	if n != 0 {
		conn.Write(buffer[:n])
	} else {
		conn.Write(buffer)

	}
	log.Println("Writing back")

}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	rand.Seed(int64(time.Now().Nanosecond()))
	log.SetFlags(log.LstdFlags)
	fmt.Println("Do you want to participate as a relay?")
	//Take input
	consoleReader := bufio.NewReader(os.Stdin)
	input, err := consoleReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	input = strings.TrimSpace(input)
	msg := ""
	if strings.ToLower(input) == "yes" || strings.ToLower(input) == "y" {
		fmt.Print("1. Entry Relay\n2. Middle Relay\n3. Exit Relay\nEnter your choice: ")
		choice, _ := consoleReader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			msg = entryRelay
			log.Println("Joined as Entry Relay")
			break
		case "2":
			msg = middleRelay
			log.Println("Joined as Middle Relay")
			break
		case "3":
			msg = exitRelay
			log.Println("Joined as Exit Relay")
			break
		default:
			break
		}
	}
	clientConnection, err = net.Dial("tcp", "localhost:6000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Directory Server!")
	if msg != "" {
		clientConnection.Write([]byte(msg + "\n"))
		clientConnection.Write([]byte("KEY\n"))
		log.Println("Relay Type and Public Key sent to Directory Server")
		go listenAsARelay(msg)
		log.Println("Started listening for other relay requests")
		time.Sleep(20000000)
		go sendAliveMessages(clientConnection, sig)
		log.Println("Started sending heartbeat")
	} else {
		clientConnection.Write([]byte("N\n"))
		clientConnection.Write([]byte("KEY\n"))
		log.Println("Relay Type and Public Key send to Directory Server")
		go sendAliveMessages(clientConnection, sig)
		log.Println("Started sending heartbeat")
	}
	defer clientConnection.Close()
	http.HandleFunc("/fastor/", clientHandler)
	port := rand.Intn(3000) + 3000
	if msg == "" {
		log.Println("Listening for client requests at port: ", port)
	}
	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	for err != nil {
		port = rand.Intn(3000) + 3000
		if msg == "" {
			log.Println("Listening for client requests at port: ", port)
		}
		err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	}

}
