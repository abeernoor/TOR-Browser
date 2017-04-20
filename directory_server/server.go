package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	read <- readinput(buf.ReadLine())
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
				responselistchan <- relays
			} else {
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
	fmt.Println("Connected new user")
	clientreader := bufio.NewReader(c)
	Option, _, _ := clientreader.ReadLine()
	fmt.Println("Read input")
	fmt.Println(c.RemoteAddr())
	var n Node
	fmt.Println("Entered Option : ", string(Option))
	//storing info of the connected relay
	n.RelayType = string(Option)
	key, _, _ := clientreader.ReadLine()
	n.IPandPort = c.RemoteAddr().String()
	n.PubKey = string(key)

	if n.RelayType != "N" {
		clientchan <- n
	}

	for {
		breakout := false
		readchan := make(chan []byte)
		go nonBlockingReader(clientreader, readchan)
		var output []byte
		fmt.Println("going for select")
		select {
		case <-time.After(15000000000):
			breakout = true
		case output = <-readchan:
		}
		fmt.Println("Read from chan", string(output))
		if breakout == false {
			if string(output) == "GET_LIST" {
				requestlistchan <- n
				res := <-responselistchan
				for _, element := range res {
					fmt.Println(element)
				}
				v, _ := json.Marshal(&res)

				c.Write(v)
			} else if string(output) == "EXIT" {
				breakout = true
			}
		}
		if breakout == true {
			if n.RelayType != "N" {
				deleteRelay <- n
			}
			fmt.Println("EXITING")
			c.Close()
			break
		}
	}
	fmt.Println("Out of loop")
}

func main() {
	ln, err := net.Listen("tcp", ":6000")
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
		if err != nil {
			fmt.Println(err)
		}
		go handleClient(conn, addRelay, removeRelay, responsechan, requestchan)
	}
}
