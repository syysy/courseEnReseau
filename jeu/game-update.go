/*
//  Implementation of the Update method for the Game structure
//  This method is called once at every frame (60 frames per second)
//  by ebiten, juste before calling the Draw method (game-draw.go).
//  Provided with a few utilitary methods:
//    - CheckArrival
//    - ChooseRunners
//    - HandleLaunchRun
//    - HandleResults
//    - HandleWelcomeScreen
//    - Reset
//    - UpdateAnimation
//    - UpdateRunners
*/

package main

import (
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HandleWelcomeScreen waits for the player to push SPACE in order to
// start the game
func (g *Game) HandleWelcomeScreen() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

// ChooseRunners loops over all the runners to check which sprite each
// of them selected
func (g *Game) ChooseRunners() (done bool) {
	done = true
	for i := range g.runners {
		if i == g.myNum {
			done = g.runners[i].ManualChoose() && done
		}
	}
	return done
}

// HandleLaunchRun countdowns to the start of a run
func (g *Game) HandleLaunchRun() bool {
	if time.Since(g.f.chrono).Milliseconds() > 1000 {
		g.launchStep++
		g.f.chrono = time.Now()
	}
	if g.launchStep >= 5 {
		g.launchStep = 0
		return true
	}
	return false
}

// UpdateRunners loops over all the runners to update each of them
func (g *Game) UpdateRunners() {
	for i := range g.runners {
		if i == g.myNum {
			g.runners[i].ManualUpdate()
		}
	}
}

// CheckArrival loops over all the runners to check which ones are arrived
func (g *Game) CheckArrival() (finished bool) {
	finished = true
	// Lorsque le joueur a fini il envoie un message au serveur avec son temps
	// Une fois que le serveur a reçu les 4 il renvoie à chacun des joueurs le classement trié
	
	for i := 0; i < 4; i++ {
		g.runners[i].CheckArrival(&g.f)
		if !g.runners[i].arrived {
			finished = false
		}
	}
	return finished
}

// Reset resets all the runners and the field in order to start a new run
func (g *Game) Reset() {
	for i := range g.runners {
		g.runners[i].Reset(&g.f)
	}
	g.f.Reset()
}

// UpdateAnimation loops over all the runners to update their sprite
func (g *Game) UpdateAnimation() {
	for i := range g.runners {
		g.runners[i].UpdateAnimation(g.runnerImage)
	}
}

// HandleResults computes the resuls of a run and prepare them for
// being displayed
func (g *Game) HandleResults() bool {
	if time.Since(g.f.chrono).Milliseconds() > 1000 || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.resultStep++
		g.f.chrono = time.Now()
	}
	if g.resultStep >= 4 && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.resultStep = 0
		return true
	}
	return false
}

// Update is the main update function of the game. It is called by ebiten
// at each frame (60 times per second) just before calling Draw (game-draw.go)
// Depending of the current state of the game it calls the above utilitary
// function and then it may update the state of the game
func (g *Game) Update() error {
	switch g.state {
	case StateWelcomeScreen:
		done := g.HandleWelcomeScreen()
		select {
		case x := <-g.channel_lecture:
			g.nbJoueurs = x
			if g.myNum == -1 {
				g.myNum = g.nbJoueurs - 1
			}
		default:
			if done && g.nbJoueurs == 4 {
				g.nbJoueurs = 0
				g.state++
			}
		}
	case StateChooseRunner:
		done := g.ChooseRunners()
		if done {
			g.channel_ecriture <- 2
			g.UpdateAnimation()
			g.state++
		}

	case StateLaunchRun:
		if g.nbJoueurs == 4 {
			if <-g.channel_lecture == 3 {
				done := g.HandleLaunchRun()
				if done {
					g.channel_temp <- 4
					g.state++
					log.Print("je lance ma course")
				} else {
					g.channel_ecriture <- 4
				}
			}
		} else {
			g.nbJoueurs = <-g.channel_lecture
		}

	case StateRun:
		g.UpdateRunners()
		g.channel_pos <- g.runners[g.myNum].xpos
		finished := g.CheckArrival()
		var tpos = strings.Split(<-g.channel_string, ";")
		p0, err := strconv.ParseFloat(cleanString(tpos[0]), 64)
		if err != nil {
			return err
		}
		p1, err := strconv.ParseFloat(cleanString(tpos[1]), 64)
		if err != nil {
			return err
		}
		p2, err := strconv.ParseFloat(cleanString(tpos[2]), 64)
		if err != nil {
			return err
		}
		p3, err := strconv.ParseFloat(cleanString(tpos[3]), 64)
		if err != nil {
			return err
		}

		tabp := []float64{p0, p1, p2, p3}
		for i := range g.runners{
			if i != g.myNum{
				g.runners[i].xpos = tabp[i]
			}
		}

		g.UpdateAnimation()
		if g.runners[g.myNum].arrived{
			g.channel_ecriture <- int(g.runners[g.myNum].runTime.Milliseconds())
		}

		if finished {
			if <-g.channel_lecture == 5 {
				var ttemp = strings.Split(<-g.channel_string, ";")
				t0, err := strconv.Atoi(cleanString(ttemp[0]))
				if err != nil {
					return err
				}
				t1, err := strconv.Atoi(cleanString(ttemp[1]))
				if err != nil {
					return err
				}
				t2, err := strconv.Atoi(cleanString(ttemp[2]))
				if err != nil {
					return err
				}
				t3, err := strconv.Atoi(cleanString(ttemp[3]))
				if err != nil {
					return err
				}

				g.runners[0].runTime = time.Duration(t0) * time.Millisecond
				g.runners[1].runTime = time.Duration(t1) * time.Millisecond
				g.runners[2].runTime = time.Duration(t2) * time.Millisecond
				g.runners[3].runTime = time.Duration(t3) * time.Millisecond

				g.state++
			}

		}

	case StateResult:
		done := g.HandleResults()
		if done {
			g.Reset()
			g.channel_ecriture <- 6
			g.state = StateLaunchRun
			log.Print("reset de run")
			g.nbJoueurs = 0
		}
	}
	return nil
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
