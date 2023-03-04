package main

import (
	"log"
	"net"
	"strconv"
	"sync"
)

var w sync.WaitGroup
var lock sync.Mutex

func main() {
	// OUVERTURE DU SERVEUR
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	defer listener.Close()

	// ATTENTE DES 4 CLIENTS
	var liste_conn = make([]net.Conn, 4)
	for i := 0; i < 4; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		}
		defer conn.Close()
		liste_conn[i] = conn
		// ON ENVOIE AU JOUEURS PRESENTS LE NOMBRE DE JOUEURS CONNECTES
		for j := 0; j <= i; j++ {
			_, err = liste_conn[j].Write([]byte(strconv.Itoa(i + 1)))
			if err != nil {
				log.Println("write error:", err)
				return
			}
		}
	}
	var continu = true
	for continu {

		var chann = make(chan int, 1)
		chann <- 0
		// ON ATTEND QUE TOUS LES JOUEURS AIENT CHOISIS, EN COMPTANT LE NOMBRES DE JOUEURS PRETS
		w.Add(4)
		go recoit_message_compte(liste_conn[0], chann, liste_conn)
		go recoit_message_compte(liste_conn[1], chann, liste_conn)
		go recoit_message_compte(liste_conn[2], chann, liste_conn)
		go recoit_message_compte(liste_conn[3], chann, liste_conn)
		w.Wait()

		// CHANNEL POUR LES TEMPS
		var ct1 = make(chan string)
		var ct2 = make(chan string)
		var ct3 = make(chan string)
		var ct4 = make(chan string)

		// ON ATTEND DE RECEVOIR LES 4 TEMPS
		w.Add(4)
		go lis_temps(liste_conn[0], ct1)
		go lis_temps(liste_conn[1], ct2)
		go lis_temps(liste_conn[2], ct3)
		go lis_temps(liste_conn[3], ct4)
		w.Wait()

		// ENVOIE DU TEMPS A TOUS LES JOUEURS
		string_temps := <-ct1 + ";" + <-ct2 + ";" + <-ct3 + ";" + <-ct4
		log.Print(string_temps)
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

// FONCTION QUI ATTEND DE RECEVOIR UN MESSAGE D'UN JOUEUR ET QUI COMPTE LE NOMBRE DE JOUEURS AYANT DEJA ENVOYE UN MESSAGE
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
		log.Print(x + 1)
		_, err := liste_conn[i].Write([]byte(strconv.Itoa(x + 1)))
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
	lock.Unlock()
	w.Done()
}

// FONCTION QUI ATTEND DE RECEVOIR UN TEMPS D'UN JOUEUR
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
