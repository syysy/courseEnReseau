package main

import (
	"log"
	"net"
	"strconv"
	"sync"
	"strings"
	"unicode"
)

var w sync.WaitGroup
var lock sync.Mutex

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	defer listener.Close()

	var liste_conn = make([]net.Conn, 4)
	for i := 0; i < 4; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		}
		defer conn.Close()
		liste_conn[i] = conn
		for j := 0; j <= i; j++ {
			_, err = liste_conn[j].Write([]byte(strconv.Itoa(i + 1)))
			if err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}

	// On doit receptionner les index du choix des runners et renvoyer a tous

	var continu = true
	for continu {

		var chann = make(chan int, 1)
		chann <- 0
		log.Print("je wait")
		w.Add(4)
		go recoit_message_compte(liste_conn[0], chann, liste_conn)
		go recoit_message_compte(liste_conn[1], chann, liste_conn)
		go recoit_message_compte(liste_conn[2], chann, liste_conn)
		go recoit_message_compte(liste_conn[3], chann, liste_conn)
		w.Wait()

		
		// channel pour la pos des joueurs
		cp0 := make(chan string, 1)
		cp1 := make(chan string, 1)
		cp2 := make(chan string, 1)
		cp3 := make(chan string, 1)

		cstop := make(chan bool, 1)

		go envoi_pos(liste_conn, []chan string{cp0, cp1, cp2, cp3}, cstop)
		w.Add(4)
		go lis_pos(liste_conn[0], cp0)
		go lis_pos(liste_conn[1], cp1)
		go lis_pos(liste_conn[2], cp2)
		go lis_pos(liste_conn[3], cp3)
		w.Wait()

		log.Print("tout le monde a fini")
		cstop <- true

		for i := 0; i < 4; i++ {
			_, err = liste_conn[i].Write([]byte("8"))
			if err != nil {
				log.Println("write error:", err)
				return
			}
		}

		var ct0 = make(chan string)
		var ct1 = make(chan string)
		var ct2 = make(chan string)
		var ct3 = make(chan string)

		w.Add(4)
		go lis_temps(liste_conn[0], ct0)
		go lis_temps(liste_conn[1], ct1)
		go lis_temps(liste_conn[2], ct2)
		go lis_temps(liste_conn[3], ct3)
		w.Wait()

		string_temps := <-ct0 + ";" + <-ct1 + ";" + <-ct2 + ";" + <-ct3
		for i := 0; i < 4; i++ {
			_, err = liste_conn[i].Write([]byte(string_temps))
			if err != nil {
				log.Println("write error:", err)
				return
			}
		}
		continu = true

	}

}

func recoit_message_compte(c net.Conn, chann chan int, liste_conn []net.Conn) {
	var tab = make([]byte, 1)
	_, err := c.Read(tab)
	if err != nil {
		log.Println("read error:", err)
		return
	}
	lock.Lock()
	x := <-chann
	chann <- x + 1
	for i := 0; i < 4; i++ {
		_, err := liste_conn[i].Write([]byte(strconv.Itoa(x + 1)))
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
	lock.Unlock()
	w.Done()
}

func lis_pos(c net.Conn, cp chan string) {
	fin := true
	for fin {
		var tab = make([]byte, 50)
		_, err := c.Read(tab)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		a, err := strconv.ParseFloat(cleanString(string(tab)), 64)
		if err != nil {
			log.Println("read error:", err)
			return
		}
		if a == 8.0 {
			log.Print("un jour a fini")
			w.Done()
			return
		}

		cp <- string(tab)
		
		
	}
}

func envoi_pos(liste_conn []net.Conn, liste_chan []chan string, cstop chan bool) {
	xpos0 := "50.0"
	xpos1 := "50.0"
	xpos2 := "50.0"
	xpos3 := "50.0"
	string_pos := "50.0;50.0;50.0;50.0"

	for {
		select {
		case <-cstop:
			return
		default:
			select {
			case x0 := <-liste_chan[0]:
				xpos0 = x0
			}
			select {
			case x1 := <-liste_chan[1]:
				xpos1 = x1
			}
			select {
			case x2 := <-liste_chan[2]:
				xpos2 = x2
			}
			select {
			case x3 := <-liste_chan[3]:
				xpos3 = x3
			}
			string_pos = xpos0 + ";" + xpos1 + ";" + xpos2 + ";" + xpos3
			for i := 0; i < 4; i++ {
				_, err := liste_conn[i].Write([]byte(string_pos))
				if err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}
	}
}

func lis_temps(c net.Conn, ct chan string) {
	var tab = make([]byte, 100)
	_, err := c.Read(tab)
	if err != nil {
		log.Println("read error:", err)
		return
	}
	w.Done()
	ct <- string(tab)
}


func cleanString(strtemp string) string {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, strtemp)
	return clean
}
