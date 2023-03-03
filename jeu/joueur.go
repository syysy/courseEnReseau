package main

import (
	"log"
	"net"
	"strconv"
)

func connexion(ip string, canal_ecriture, canal_lecture, canal_temp chan int, canal_string chan string, canal_pos chan float64) {

	//ETABLIR LA CONNEXION
	conn, err := net.Dial("tcp", ip+":8080")
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close()

	log.Println("je suis connecté")

	var tab = make([]byte, 1)
	var msg = "0"
	//EN ATTENTE DES 4 JOUEURS
	for msg != "4" {
		_, err = conn.Read(tab)

		if err != nil {
			log.Println("read error:", err)
			return
		}

		msg = string(tab)
		msgI, err := strconv.Atoi(msg)
		if err != nil {
			log.Println("read error:", err)
			return
		}
		canal_ecriture <- msgI
	}
	//CHOISI LE PERSO

	for <-canal_lecture != 2 {

	}

	var continu = true
	for continu {
		_, err = conn.Write([]byte("2"))
		if err != nil {
			log.Println("write error:", err)
			return
		}

		//ATTEND QUE TOUT LE MONDE AI CHOISI
		msg = "0"
		for msg != "4" {
			_, err = conn.Read(tab)

			if err != nil {
				log.Println("read error:", err)
				return
			}

			msg = string(tab)
			msgI, err := strconv.Atoi(msg)
			if err != nil {
				log.Println("read error:", err)
				return
			}
			canal_ecriture <- msgI
		}

		//LANCEMENT DE LA COURSE
		canal_ecriture <- 3
		log.Print("vroum vroum")
		t := true
		for t {
			select {
			case <-canal_lecture:
				canal_ecriture <- 3
			case <-canal_temp:
				t = false
			}
		}

		// ENVOIE EN BOUCLE SON TEMPS JUSQUA CE QU'IL AIT FINI
		var temps int
		cours := true

		 //// ALLO CACA LE TABLEAU IL EST TROP GRAND ET CA MULTIPLIE LES VALEURS DEDANS ALED
		for cours {
			select {
			case temps = <-canal_lecture:
				cours = false
			case pos := <-canal_pos:
				_, err = conn.Write([]byte(strconv.FormatFloat(pos, 'f', 1, 64)))
				if err != nil {
					log.Println("write error:", err)
					return
				}
				var tab_pos = make([]byte, 10000)
				_, err = conn.Read(tab_pos)
				if err != nil {
					log.Println("read error:", err)
					return
				}
				canal_string <- string(tab_pos)
			}
		}

		// ATTEND QUE TOUT LE MONDE A FINI POUR ENVOYER SON TEMPS
		_, err = conn.Read(tab)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		// ENVOIE LE TEMPS AU SERVEUR
		_, err = conn.Write([]byte(strconv.Itoa(temps)))
		if err != nil {
			log.Println("write error:", err)
			return
		}

		// ATTEND LE RETOUR DE TOUS LES TEMPS
		var tab_temps = make([]byte, 10000)
		_, err = conn.Read(tab_temps)
		if err != nil {
			log.Println("read error:", err)
			return
		}

		canal_ecriture <- 5
		canal_string <- string(tab_temps)

		for <-canal_lecture != 6 {

		}
	}
}
