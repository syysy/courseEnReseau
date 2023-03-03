/*
// Implementation of a main function setting a few characteristics of
// the game window, creating a game, and launching it
*/

package main

import (
	"flag"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800 // Width of the game window (in pixels)
	screenHeight = 160 // Height of the game window (in pixels)
)

func main() {

	var getTPS bool
	var ip string
	flag.BoolVar(&getTPS, "tps", false, "Afficher le nombre d'appel à Update par seconde")
	flag.StringVar(&ip, "ip", "127.0.0.1", "Ip du serveur")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Vroom vroom va plus vite")

	g := InitGame()
	g.getTPS = getTPS
	g.ip = ip
	g.myNum = -1

	canal_ecriture := make(chan int, 1)
	canal_lecture := make(chan int, 1)
	canal_temp := make(chan int, 1)
	canal_string := make(chan string, 1)
	canal_pos := make(chan float64, 1)
	g.channel_ecriture = canal_ecriture
	g.channel_lecture = canal_lecture
	g.channel_temp = canal_temp
	g.channel_string = canal_string
	g.channel_pos = canal_pos
	go connexion(g.ip, canal_lecture, canal_ecriture, canal_temp, canal_string, canal_pos)
	err := ebiten.RunGame(&g)
	log.Print(err)

}
