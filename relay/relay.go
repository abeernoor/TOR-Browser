package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const httpPort = ":7000"
const otherPort = ":7001"
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
<<<<<<< HEAD
		d, _ := time.ParseDuration("2s")
		time.Sleep(d)
		conn.Write([]byte("ALIVE\n"))
=======
		select {
		case <-sig:
			fmt.Println("signal got getting out", conn.Close())
			clientConnection.Close()
			return
		case <-time.After(d):
			conn.Write([]byte("ALIVE\n"))
		}
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
	}
}

func clientHandler(rw http.ResponseWriter, r *http.Request) {
	//Get list from directory server
	//Generate a random path of relays
	//Send to entry relay: address of mid and exit relay and URL
	//Wait for response
<<<<<<< HEAD
	clientConnection.Write([]byte("GET_LIST\n"))
	buffer := make([]byte, 0, 1024)
	n, _ := clientConnection.Read(buffer)
=======
	fmt.Println("GETTING LIST")
	clientConnection.Write([]byte("GET_LIST\n"))
	fmt.Println("got list")
	buffer := make([]byte, 2048, 10240)
	n, err := clientConnection.Read(buffer)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Recieved Buffer", buffer)
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
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

	tData := torData{MidRelay: chMidRelay.IPandPort, ExitRelay: chExitRelay.IPandPort, URL: "http://" + r.URL.Path[len("/fastor/"):], PageBody: ""}
	fmt.Println(tData)
	conn, _ := net.Dial("tcp", ":"+chEntryRelay.IPandPort)
	buffer, _ = json.Marshal(tData)
	fmt.Println(buffer)
	conn.Write(buffer)
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
		fmt.Println(err)
	}
	newTData := torData{}
	json.Unmarshal(newBuffer[:n], &newTData)
	fmt.Println("Printing the fetched Page")
	// fmt.Println(newTData.PageBody)
	rw.Write([]byte(newTData.PageBody))
}

func listenAsARelay(relayType string) {
<<<<<<< HEAD

	line, err := net.Listen("tcp", ":"+string(rand.Intn(3000)+4000))
	for err != nil {
		line, err = net.Listen("tcp", ":"+string(rand.Intn(3000)+4000))
	}
=======
	defer clientConnection.Close()
	port := strconv.Itoa(rand.Intn(3000) + 4000)
	line, err := net.Listen("tcp", ":"+port)
	defer line.Close()
	for err != nil {
		port = strconv.Itoa(rand.Intn(3000) + 4000)
		line, err = net.Listen("tcp", ":"+port)
	}
	clientConnection.Write([]byte(port + "\n"))

>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
	for {
		con, _ := line.Accept()
		go handleRelayConnection(con, relayType)

	}
}

func handleRelayConnection(conn net.Conn, relayType string) {
	//Forward to next relay the info
	//or get page from server if you are exit relay
	buffer := make([]byte, 1024, 10240)
	n, _ := conn.Read(buffer)
	defer conn.Close()
	tData := torData{}
	json.Unmarshal(buffer[:n], &tData)
	n = 0
	if relayType == entryRelay {
		midRelayConn, _ := net.Dial("tcp", ":"+tData.MidRelay)
		tData.MidRelay = ""
		newBuffer, _ := json.Marshal(tData)
		fmt.Println(newBuffer)
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
		fmt.Println("printing recieved from middle")
		fmt.Println(buffer[:n])
	} else if relayType == middleRelay {
		exitRelayConn, _ := net.Dial("tcp", ":"+tData.ExitRelay)
		tData.ExitRelay = ""
		newBuffer, err := json.Marshal(tData)
		if err != nil {
			fmt.Println(err)
		}
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
		fmt.Println("printing recieved from exit")
		fmt.Println(buffer[:n])

	} else {
		fmt.Println(tData.URL)
		response, err := http.Get(tData.URL)
		if err != nil {
			fmt.Println(err)
		}
		responseBody, er := ioutil.ReadAll(response.Body)
		if er != nil {
			fmt.Println(er)
		}
		// fmt.Println(string(responseBody))
		response.Body.Close()
		tData.PageBody = string(responseBody)
		buffer, _ = json.Marshal(tData)
		size := len(buffer)
		str := strconv.Itoa(size) + "\n"
		fmt.Println("size string ", str)
		buffersize, _ := json.Marshal(str)
		// fmt.Println("before : ", len(buffersize), cap(buffersize))
		// buffersize = append([]byte(nil), buffersize[:len(buffersize)]...)
		// fmt.Println("after : ", len(buffersize), cap(buffersize))
		// clientwriter := bufio.NewWriter(conn)
		// clientwriter.Write(buffersize)
		// _, err = conn.Write(buffersize)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		fmt.Println(size, buffersize)
	}
	if n != 0 {
		conn.Write(buffer[:n])
	} else {
		conn.Write(buffer)

	}
	fmt.Println("writing back")

}

func main() {
<<<<<<< HEAD
=======
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
	rand.Seed(int64(time.Now().Nanosecond()))
	fmt.Println("Do you want to participate as a relay?")
	//Take input
	consoleReader := bufio.NewReader(os.Stdin)
	input, _ := consoleReader.ReadString('\n')
	input = strings.TrimSpace(input)
	msg := ""
	if strings.ToLower(input) == "yes" || strings.ToLower(input) == "y" {
		fmt.Println("1")
		fmt.Print("1. Entry Relay\n2. Middle Relay\n3. Exit Relay\nEnter your choice: ")
		choice, _ := consoleReader.ReadString('\n')
		choice = strings.TrimSpace(choice)
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
	}
	clientConnection, _ = net.Dial("tcp", "localhost:6000")
	if msg != "" {
		fmt.Println("1")
		clientConnection.Write([]byte(msg + "\n"))
		clientConnection.Write([]byte("KEY\n"))
<<<<<<< HEAD
		go sendAliveMessages(clientConnection)
=======
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
		go listenAsARelay(msg)
		time.Sleep(20000000)
		go sendAliveMessages(clientConnection, sig)

	} else {
		clientConnection.Write([]byte("N\n"))
<<<<<<< HEAD
=======
		clientConnection.Write([]byte("KEY\n"))
		go sendAliveMessages(clientConnection, sig)
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
	}
	defer clientConnection.Close()
	http.HandleFunc("/fastor/", clientHandler)
<<<<<<< HEAD
	err := http.ListenAndServe(":"+string(rand.Intn(3000)+3000), nil)
	for err != nil {
		err = http.ListenAndServe(":"+string(rand.Intn(3000)+3000), nil)
=======
	fmt.Println("assigning port")
	port := rand.Intn(3000) + 3000
	fmt.Println("Might be Listening at port : ", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	fmt.Println(err)
	for err != nil {
		port = rand.Intn(3000) + 3000
		fmt.Println("Might be Listening at port : ", port)

		fmt.Println(":" + strconv.Itoa(port))
		err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
		fmt.Println(err)
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
	}

}
