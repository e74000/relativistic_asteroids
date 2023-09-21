package main

import (
	"github.com/charmbracelet/log"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Create a new Game object
	g := new(Game)
	g.Init(800, 600)

	// Initialise the window
	ebiten.SetWindowTitle("Relativistic Asteroids")
	ebiten.SetWindowSize(g.screenWidth, g.screenHeight)

	log.Debug("Starting the game loop")

	// Run the game
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal("error running game", "error", err)
	}
}
