package main

import (
	"bufio"
	"container/list"
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

//InputRead stores the info got from ReadLine Function
type InputRead struct {
	line     []byte
	isPrefix bool
	err      error
}

func (ir *InputRead) readinput(line []byte, isPrefix bool, err error) []byte {
	ir.err = err
	ir.isPrefix = isPrefix
	ir.line = line
	return line
}

func nonBlockingReader(buf *bufio.Reader, read chan<- []byte) {
	input := InputRead{}
	select {
	case read <- input.readinput(buf.ReadLine()):
	}

}

// HandlingRelayList maintains the list of active relays
func HandlingRelayList(relays *list.List, newRelay <-chan Node, deleteRelay <-chan Node, requestlistchan <-chan Node, responselistchan chan<- *list.List) {
	for {
		select {
		case newRelay := <-newRelay:
			relays.PushBack(newRelay)
			// list = append(list, newRelay)
		case <-requestlistchan:
			responselistchan <- relays
		case req := <-deleteRelay:
			for element := relays.Front(); element != nil; element = element.Next() {
				if req == element.Value {
					relays.Remove(element)
					break
				}
			}
		}
	}
}

func handleClient(c net.Conn, clientchan chan<- Node, deleteRelay chan<- Node, requestlistchan chan<- Node, responselistchan <-chan *list.List) {
	fmt.Println("Connected new user")
	//c.Write([]byte("To become Entry Relay (EN) Intermediate relay (I), Exit relay (EX) or not to participate (N)\n"))
	clientreader := bufio.NewReader(c)
	Option, _, _ := clientreader.ReadLine()
	fmt.Println("Read input")
	fmt.Println(c.RemoteAddr())
	var n Node
	addr := c.RemoteAddr()
	fmt.Println("Entered Option : ", string(Option))
	n.RelayType = string(Option)
	key, _, _ := clientreader.ReadLine()
	n.IPandPort = addr.String()
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
		case <-time.After(5000000000):
			breakout = true
		case output = <-readchan:
		}
		fmt.Println("Read from chan", string(output))
		if breakout == false {
			if string(output) == "GET_LIST" {
				requestlistchan <- n
				res := <-responselistchan
				for i := res.Front(); i != nil; i = i.Next() {
					fmt.Println(i.Value)
				}
				v, _ := json.Marshal(res)
				fmt.Println(v)
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
	mylist := list.New()
	addRelay := make(chan Node)
	removeRelay := make(chan Node)
	requestchan := make(chan *list.List)
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
