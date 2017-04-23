package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Node is struct that stores the relay info
type Node struct {
	RelayType string
	IPandPort string
	PubKey    string
}

func readinput(line []byte, isPrefix bool, err error) []byte {
	return line
}

func nonBlockingReader(buf *bufio.Reader, read chan<- []byte) {
	line, _, err := buf.ReadLine()
	if err != nil {
		read <- []byte("EXIT")
	} else {
		read <- line
	}
}

// HandlingRelayList maintains the list of active relays
func HandlingRelayList(relays []Node, newRelay <-chan Node, deleteRelay <-chan Node, requestlistchan <-chan Node, responselistchan chan<- []Node) {
	for {
		select {
		case newRelay := <-newRelay:
			relays = append(relays, newRelay)
		case <-requestlistchan:
			en := false
			ex := false
			I := false
			for _, element := range relays {
				if element.RelayType == "EN" {
					en = true
				} else if element.RelayType == "I" {
					I = true
				} else if element.RelayType == "EX" {
					ex = true
				}
			}
			if en == true && I == true && ex == true {
				fmt.Println("all present")
				responselistchan <- relays
			} else {
				fmt.Println("Not complete", en, I, ex)
				responselistchan <- nil
			}
		case req := <-deleteRelay:
			for i, element := range relays {
				if element == req {
					relays = append(relays[:i], relays[i+1:]...)
					break
				}
			}
		}
	}
}

func handleClient(c net.Conn, clientchan chan<- Node, deleteRelay chan<- Node, requestlistchan chan<- Node, responselistchan <-chan []Node) {
	defer c.Close()
	fmt.Println("Connected new user")
	clientreader := bufio.NewReader(c)
	Option, _, _ := clientreader.ReadLine()
	fmt.Println("Read input")
	fmt.Println(c.RemoteAddr())
	var n Node
	fmt.Println("Entered Option : ", string(Option))
	//storing info of the connected relay
	n.RelayType = strings.TrimSpace(string(Option))
	key, _, _ := clientreader.ReadLine()
	n.PubKey = strings.TrimSpace(string(key[:len(key)-1]))
	fmt.Println(n)
	if n.RelayType != "N" {
		port, _, _ := clientreader.ReadLine()
		n.IPandPort = strings.TrimSpace(string(port))
		clientchan <- n
	}

	for {
		breakout := false
		readchan := make(chan []byte)
		go nonBlockingReader(clientreader, readchan)
		var output []byte
<<<<<<< HEAD
		//fmt.Println("going for select")
=======
		// fmt.Println("going for select")
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
		select {
		case <-time.After(5000000000):
			breakout = true
		case output = <-readchan:
		}
<<<<<<< HEAD
		//fmt.Println("Read from chan", string(output))
		if breakout == false && len(output) > 0 {
			if string(output) == "GET_LIST" {
=======
		// fmt.Println("Read from chan", string(output))
		if breakout == false && len(output) > 0 {
			str := strings.TrimSpace(string(output))
			if str == "GET_LIST" {
				fmt.Println("sending list")
>>>>>>> c5c103c7720377e097fd2a7e230f7a55bccdd22a
				requestlistchan <- n
				res := <-responselistchan
				for _, element := range res {
					fmt.Println(element)
				}
				v, _ := json.Marshal(&res)
				fmt.Println("Sent list", v)
				_, err := c.Write(v)
				if err != nil {
					fmt.Println(err)
				}
			} else if str == "EXIT" {
				breakout = true
			}
		}
		if breakout == true {
			if n.RelayType != "N" {
				deleteRelay <- n
			}
			break
		}
	}
	fmt.Println("Exiting : ", n)
}

func main() {
	ln, err := net.Listen("tcp", ":6000")
	defer ln.Close()
	if err != nil {
		log.Fatal(err)
	}
	mylist := make([]Node, 0, 5)
	addRelay := make(chan Node)
	removeRelay := make(chan Node)
	requestchan := make(chan []Node)
	responsechan := make(chan Node)
	go HandlingRelayList(mylist, addRelay, removeRelay, responsechan, requestchan)
	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println(err)
		}
		go handleClient(conn, addRelay, removeRelay, responsechan, requestchan)
	}

}
