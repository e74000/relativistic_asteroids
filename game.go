package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"time"
)

// Defining textures that are used in the game.
var (
	//go:embed starfield.kage
	starfieldData []byte

	//go:embed textures/ship.png
	shipData []byte

	//go:embed textures/ship_thrust.png
	shipThrustData []byte

	//go:embed textures/arrow.png
	arrowData []byte

	//go:embed textures/asteroid_big.png
	bigAsteroidData []byte

	//go:embed textures/asteroid_small.png
	smallAsteroidData []byte

	//go:embed textures/spectra.png
	spectraData []byte

	//go:embed textures/bullet.png
	bulletData []byte

	//go:embed textures/explosion.png
	explosionData []byte

	// The starfield shader
	starfield *ebiten.Shader

	// The ship sprite
	ship *ebiten.Image

	// The ship thrust sprite
	shipThrust *ebiten.Image

	// The arrow sprite
	arrow *ebiten.Image

	// The bigAsteroid sprite
	bigAsteroid *ebiten.Image

	// The smallAsteroid sprite
	smallAsteroid *ebiten.Image

	// The bullet sprite
	bullet *ebiten.Image

	// The explosion sprite
	explosion *ebiten.Image

	// The redshift data stored from the spectra sprite
	redshiftData [3][100]float64
)

// Defining fonts/colours that are used in the game.
var (
	guiFont font.Face

	colorDefault = colornames.White
	colorTitle   = colornames.Cyan
	colorHealth  = colornames.Red
	colorDamage  = colornames.Red
	colorSuccess = colornames.Limegreen
	colorFailure = colornames.Red
)

// Initialising random numbers for shaders
var (
	shaderSeed = rand.Float64()*10000 - 5000
)

// Defining the constants used in the game.
const (
	// the seconds per tick
	dt float64 = 1.0 / 60.0

	// The scale factor of the physics engine
	physScale float64 = 0.3
)

// Game is the main struct of the (relativistic) asteroids clone
type Game struct {
	// Particles
	ship      *Particle // The ship particle
	asteroids *Pool     // The asteroids pool
	bullets   *Pool     // The bullets pool
	explosion *Pool     // The explosion particles pool

	// Graphical elements
	thrusting    bool      // Whether the ship is thrusting
	health       int       // The health of the ship
	ammo         int       // The amount of ammo the ship has
	score        int       // The player's score
	highScore    int       // The high score of the player
	newHighScore bool      // Whether the high score of the player has changed
	mainMenu     bool      // Whether the main menu screen is shown
	gameOver     bool      // Whether the game over screen is shown
	bgScroll     float64   // The position of the starfield background in the main menu / game over screen
	screenStart  time.Time // The time at which the current screen became visible
	gameEnd      bool      // Whether the game has ended
	gameEndTime  time.Time // The time at which the game ends after the player has died

	// Important values
	screenWidth  int // The width of the screen
	screenHeight int // The height of the screen

	// Physical constants
	c float64 // The speed of light

	// Interpolation variables
	cLerpInitial float64       // The initial value of the speed of light when interpolating
	cLerpTarget  float64       // The target value of the speed of light when interpolating
	cLerpStart   time.Time     // The start time of interpolating the speed of light
	cLerpTime    time.Duration // How long to interpolate the speed of light
	cLerp        bool          // Whether the speed of light is interpolating

	invincibilityStart    time.Time     // The start time of the invincibility
	invincibilityDuration time.Duration // The duration of the invincibility
	invincibility         bool          // Whether the ship is invincible

	clock float64 // The time from the point of view of a stationary observer
}

func (g *Game) beginCLerp(c float64) {
	g.cLerpInitial = g.c
	g.cLerpTarget = c
	g.cLerpStart = time.Now()
	g.cLerp = true
}

// Init initializes the game object
func (g *Game) Init(screenWidth, screenHeight int) {
	log.Debug("initialising game object", "screenWidth", screenWidth, "screenHeight", screenHeight)
	g.screenWidth = screenWidth   // Initialize the width of the screen
	g.screenHeight = screenHeight // Initialize the height of the screen
	g.c = 299792458.0             // Initialize the speed of light

	log.Debug("loading shader")

	// Load the starfield shader
	var err error
	starfield, err = ebiten.NewShader(starfieldData)
	if err != nil {
		log.Fatal("failed to load starfield shader", "error", err)
		return
	}

	log.Debug("all shaders loaded, loading textures")

	log.Debug("loading ship texture")

	// Load the ship sprite
	img, err := png.Decode(bytes.NewReader(shipData))
	if err != nil {
		log.Fatal("failed to load ship texture", "error", err)
		return
	}

	ship = ebiten.NewImageFromImage(img)

	log.Debug("loading ship thrust texture")

	// Load the ship thrust sprite
	img, err = png.Decode(bytes.NewReader(shipThrustData))
	if err != nil {
		log.Fatal("failed to load ship thrust texture", "error", err)
		return
	}

	shipThrust = ebiten.NewImageFromImage(img)

	log.Debug("loading arrow texture")

	img, err = png.Decode(bytes.NewReader(arrowData))
	if err != nil {
		log.Fatal("failed to load arrow texture", "error", err)
		return
	}

	arrow = ebiten.NewImageFromImage(img)

	log.Debug("loading big_asteroid texture")

	// Load the bigAsteroid sprite
	img, err = png.Decode(bytes.NewReader(bigAsteroidData))
	if err != nil {
		log.Fatal("failed to load big_asteroid texture", "error", err)
		return
	}

	bigAsteroid = ebiten.NewImageFromImage(img)

	log.Debug("loading small_asteroid texture")

	// Load the smallAsteroid sprite
	img, err = png.Decode(bytes.NewReader(smallAsteroidData))
	if err != nil {
		log.Fatal("failed to load small_asteroid texture", "error", err)
		return
	}

	smallAsteroid = ebiten.NewImageFromImage(img)

	log.Debug("loading bullet texture")

	// Load the bullet sprite
	img, err = png.Decode(bytes.NewReader(bulletData))
	if err != nil {
		log.Fatal("failed to load bullet texture", "error", err)
		return
	}

	bullet = ebiten.NewImageFromImage(img)

	log.Debug("loading explosion texture")

	// Load the explosion sprite
	img, err = png.Decode(bytes.NewReader(explosionData))
	if err != nil {
		log.Fatal("failed to load explosion texture", "error", err)
		return
	}

	explosion = ebiten.NewImageFromImage(img)

	log.Debug("all textures loaded, loading redshift data")

	// Load redshift data
	redshiftData = getRedshift()

	log.Debug("redshift data loaded, loading fonts")

	// Load the GUI font
	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal("failed to parse font", "error", err)
		return
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal("failed to create font face", "error", err)
		return
	}

	guiFont = text.FaceWithLineHeight(face, 24)

	log.Debug("all fonts loaded, initializing values")

	g.highScore = 0
	g.mainMenu = true
	g.gameOver = false
	g.bgScroll = 0
	g.screenStart = time.Now()

	g.cLerpTime = time.Second * 2
	g.invincibilityDuration = time.Millisecond * 1000

	log.Debug("game initialised")
}

func (g *Game) startGame() {
	log.Debug("starting new game")
	// Initialise the speed of light
	g.c = 299792458.0

	// Initialise the score and health
	g.health = 3
	g.score = 0
	g.ammo = 10

	// Initialize the ship
	g.ship = new(Particle)
	g.ship.Mass = 1
	g.ship.Radius = 1

	log.Debug("initialising asteroids pool")

	// Initialize the asteroids
	g.asteroids = newAsteroidPool(64, 20).
		SetSpriteSheet(64, bigAsteroid, smallAsteroid) // Set the sprite for the asteroids

	log.Debug("initialising bullets pool")

	// Initialize the bullets
	g.bullets = NewPool(256).
		SetSpriteSheet(64, bullet). // Set the sprite for the bullets
		EnforceLifetime(time.Second * 10). // Enforce a lifetime of 10 seconds
		DisableCollision() // Disable collision between bullets

	log.Debug("initialising explosion pool")

	// Initialize the explosion particles
	g.explosion = NewPool(256).
		SetSpriteSheet(64, explosion). // Set the sprite for the explosions
		EnforceLifetime(time.Second * 1). // Enforce a lifetime of 1 second
		FadeOverLifetime(). // Fade out the explosion particles over the lifetime of the particle
		DisableCollision() // Disable collision between explosions

	log.Debug("all pools initialised, starting game")

	g.newHighScore = false
	g.mainMenu = false
	g.gameOver = false
	g.screenStart = time.Now()
}

// endGame ends the game and checks if the score is higher than the high score
func (g *Game) endGame() {
	log.Debug("ending game")

	if g.score > g.highScore {
		log.Debug("new high score", "score", g.score, "highScore", g.highScore)
		g.highScore = g.score
		g.newHighScore = true
	}

	g.gameEnd = true
	g.gameEndTime = time.Now()
}

// gameUpdate is called every physics update whenever the game is being played
func (g *Game) gameUpdate() error {
	// If the game has ended, wait 5 seconds before moving to the gameOver screen
	if time.Since(g.gameEndTime).Seconds() > 5 && g.gameEnd {
		log.Debug("moving to game over screen")
		g.gameEnd = false
		g.gameOver = true
		g.screenStart = time.Now()
		return nil
	} else if g.gameEnd {
		g.asteroids.Update(g.ship.Vel, g.c, dt)
		g.bullets.Update(g.ship.Vel, g.c, dt)
		g.explosion.Update(g.ship.Vel, g.c, dt)
		return nil
	}

	// Terminate program if escape key is pressed
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		log.Debug("received escape key")
		g.endGame()
	}

	// Initialise the thrust direction
	thrust := Vector{0, 0}

	// If the ship is thrusting add a force and enable thrusting flag
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		thrust.X = math.Sin(g.ship.AngPos)
		thrust.Y = -math.Cos(g.ship.AngPos)
		g.thrusting = true
	} else if g.thrusting {
		g.thrusting = false
	}

	// Rotate the ship if A or D keys are pressed
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.ship.AngPos -= 0.05
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.ship.AngPos += 0.05
	}

	// If space is pressed shoot a bullet
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.ammo > 0 {
		g.bullets.Activate(
			g.ship.Pos,
			g.ship.Rap.Add(Vector{
				X: 3 * math.Sin(g.ship.AngPos),
				Y: 3 * -math.Cos(g.ship.AngPos),
			}),
			0, 1,
			1, 1,
			0,
		)

		g.ammo--
	}

	//Update the ship
	g.ship.Update(
		thrust.Scl(10).Add(getDrag(g.ship.Vel, 0.1)),
		g.ship.Vel,
		g.c,
		dt,
	)

	if len(g.asteroids.Collisions(g.ship, g.ship.Vel)) > 0 && !g.invincibility {
		log.Debug("ship hit an asteroid")

		g.invincibility = true
		g.invincibilityStart = time.Now()
		g.health--

		g.ship.AngVel += float64(rand.Int63n(2)*2-1) * 10
	}

	if time.Since(g.invincibilityStart) > g.invincibilityDuration && g.invincibility {
		log.Debug("invincibility period ended")
		g.invincibility = false
		g.ship.AngVel = 0
	}

	if g.health <= 0 || g.ammo <= 0 {
		log.Debug("game over", "health", g.health, "ammo", g.ammo)
		explode(g.explosion, g.ship)
		g.endGame()
	}

	// Update the asteroids and solve for collisions with the ship
	g.asteroids.UpdateWith(g.ship.Vel, g.c, dt, g.ship)

	// Update the bullets
	g.bullets.Update(g.ship.Vel, g.c, dt)

	// Update the explosion particles
	g.explosion.Update(g.ship.Vel, g.c, dt)

	// Get all collisions between bullets and asteroids and for each:
	for _, ints := range g.bullets.PoolCollisions(g.asteroids, g.ship.Vel) {
		log.Debug("bullet hit asteroid", "bullet", ints[0], "asteroid", ints[1])
		g.bullets.Deactivate(ints[0]) // Remove the bullet from the game
		// If the asteroid is a big asteroid, add three small asteroids in its place
		if g.asteroids.sprites[ints[1]] == 0 {
			log.Debug("big asteroid hit, adding small asteroids")
			parent := g.asteroids.particles[ints[1]]

			for i := 0; i < 3; i++ {
				radius := rand.Float64()*0.5 + 0.5
				mass := math.Pi * math.Pow(radius, 2)

				g.asteroids.Activate(
					parent.Pos.Add(Vector{1, 0}.Rotate(float64(i)*math.Pi*2.0/3.0).Scl(0.2)),
					parent.Rap.Add(Vector{1, 0}.Rotate(float64(i)*math.Pi*2.0/3.0).Scl(0.1)),
					rand.Float64()*math.Pi*2,
					rand.Float64()*0.25-0.125,
					mass, radius,
					1,
				)
			}

			g.score += 3
			g.ammo += 3
		} else {
			g.score += 1
			g.ammo += 1
		}

		explode(g.explosion, g.asteroids.particles[ints[1]])
		g.asteroids.Deactivate(ints[1]) // Remove the bigAsteroid from the game 		// Increment the score by 1
		g.beginCLerp(g.c/8 + 10.0)      // Begin reducing the speed of light
	}

	if g.cLerp {
		g.c = mapRange(time.Since(g.cLerpStart).Seconds(), 0, g.cLerpTime.Seconds(), g.cLerpInitial, g.cLerpTarget)

		if time.Since(g.cLerpStart).Seconds() > g.cLerpTime.Seconds() {
			g.c = g.cLerpTarget
			g.cLerp = false
		}
	}

	g.clock += dt * Gamma(g.ship.Vel.Mag(), g.c)

	return nil
}

// mainMenuUpdate is called every physics update whenever the main menu is being displayed
func (g *Game) mainMenuUpdate() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		log.Debug("leaving main menu")
		g.startGame()
	}

	g.bgScroll -= dt * 10

	return nil
}

// gameOverUpdate is called every physics update whenever the game over screen is being displayed
func (g *Game) gameOverUpdate() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		log.Debug("moving to main menu")
		g.gameOver = false
		g.mainMenu = true
		g.screenStart = time.Now()
	}

	g.bgScroll -= dt * 10

	return nil
}

// Update is called every physics update
func (g *Game) Update() error {
	if g.mainMenu {
		return g.mainMenuUpdate()
	} else if g.gameOver {
		return g.gameOverUpdate()
	} else {
		return g.gameUpdate()
	}
}

// gameDraw is called every frame when the game is being played
func (g *Game) gameDraw(screen *ebiten.Image) {
	// Draw the starfield background with a rectangle shader
	screen.DrawRectShader(g.screenWidth, g.screenHeight, starfield, &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"C":             g.c,                    // Pass in the speed of light
			"RedshiftRed":   redshiftData[0],        // Pass in redshift color data for the red channel
			"RedshiftGreen": redshiftData[1],        // Pass in redshift color data for the green channel
			"RedshiftBlue":  redshiftData[2],        // Pass in redshift color data for the blue channel
			"Position":      g.ship.Pos.ToUniform(), // Pass in the position of the ship
			"Velocity":      g.ship.Vel.ToUniform(), // Pass in the velocity of the ship
			"Seed":          shaderSeed,             // Pass in the seed
		},
	})

	// Initialise the color scale as nil
	var colorScale *ebiten.ColorScale

	// If the ship is invincible adjust the color scale to show this effect
	if g.invincibility {
		colorScale = &ebiten.ColorScale{}
		colorScale.ScaleWithColor(colorMix(colorDamage, color.White, math.Abs(math.Sin(time.Since(g.invincibilityStart).Seconds()*4*math.Pi))))
	}

	// Draw the ship, if thrusting, use alt texture
	if g.thrusting && !g.gameEnd {
		g.ship.Draw(screen, shipThrust, 64, g.ship.Pos, g.ship.Vel, colorScale)
	} else if !g.gameEnd {
		g.ship.Draw(screen, ship, 64, g.ship.Pos, g.ship.Vel, colorScale)
	}

	// Draw the asteroids
	g.asteroids.Draw(screen, g.ship.Pos, g.ship.Vel)

	// Draw the bullets
	g.bullets.Draw(screen, g.ship.Pos, g.ship.Vel)

	// Draw the explosion particles
	g.explosion.Draw(screen, g.ship.Pos, g.ship.Vel)

	// Draw arrows to the closest bigAsteroid
	if !g.gameEnd {
		closestPos := g.asteroids.Closest(g.ship.Pos)
		drawArrow(screen, g.ship.Pos, closestPos)
	}

	// Draw the health of the ship, score and the speed of light
	text.Draw(screen, fmt.Sprintf("♥ %d", g.health), guiFont, 10, 24, colorHealth)
	text.Draw(screen, fmt.Sprintf("! %d", g.ammo), guiFont, 10, 48, colorDefault)
	text.Draw(screen, fmt.Sprintf("SCORE: %04d", g.score), guiFont, 610, 24, colorDefault)
	text.Draw(screen, fmt.Sprintf("v = %fc\nc = %sm/s", g.ship.Vel.Mag()/g.c, scientificNotation(g.c)), guiFont, 10, 572, colorDefault)
}

// mainMenuDraw is called every frame when the main menu is being displayed
func (g *Game) mainMenuDraw(screen *ebiten.Image) {
	screen.DrawRectShader(g.screenWidth, g.screenHeight, starfield, &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"C":             g.c,                               // Pass in the speed of light
			"RedshiftRed":   redshiftData[0],                   // Pass in redshift color data for the red channel
			"RedshiftGreen": redshiftData[1],                   // Pass in redshift color data for the green channel
			"RedshiftBlue":  redshiftData[2],                   // Pass in redshift color data for the blue channel
			"Position":      Vector{0, g.bgScroll}.ToUniform(), // Pass in the position of the ship
			"Velocity":      Vector{}.ToUniform(),              // Pass in the velocity of the ship
			"Seed":          shaderSeed + 10000,                // Pass in the seed
		},
	})

	if time.Since(g.screenStart).Seconds() > 1.0 {
		text.Draw(screen, "RELATIVISTIC ASTEROIDS", guiFont, 230, 24, colorTitle)
	}

	if time.Since(g.screenStart).Seconds() > 2.0 {
		text.Draw(screen, "Press space to start", guiFont, 250, 572, colorDefault)
	}
}

// gameOverDraw is called every frame when the game over screen is being displayed
func (g *Game) gameOverDraw(screen *ebiten.Image) {
	screen.DrawRectShader(g.screenWidth, g.screenHeight, starfield, &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"C":             g.c,                               // Pass in the speed of light
			"RedshiftRed":   redshiftData[0],                   // Pass in redshift color data for the red channel
			"RedshiftGreen": redshiftData[1],                   // Pass in redshift color data for the green channel
			"RedshiftBlue":  redshiftData[2],                   // Pass in redshift color data for the blue channel
			"Position":      Vector{0, g.bgScroll}.ToUniform(), // Pass in the position of the ship
			"Velocity":      Vector{}.ToUniform(),              // Pass in the velocity of the ship
			"Seed":          shaderSeed + 10000,                // Pass in the seed
		},
	})

	if time.Since(g.screenStart).Seconds() > 1.0 {
		if g.newHighScore {
			text.Draw(screen, "NEW HIGH SCORE", guiFont, 300, 24, colorSuccess)
		} else {
			text.Draw(screen, "GAME OVER", guiFont, 320, 24, colorFailure)
		}
	}

	if time.Since(g.screenStart).Seconds() > 1.5 {
		text.Draw(screen, fmt.Sprintf("  SCORE: %04d", g.score), guiFont, 280, 100, colorDefault)
	}

	if time.Since(g.screenStart).Seconds() > 2.0 {
		text.Draw(screen, fmt.Sprintf("HISCORE: %04d", g.highScore), guiFont, 280, 124, colorDefault)
	}

	if time.Since(g.screenStart).Seconds() > 3 {
		text.Draw(screen, "Press space to continue", guiFont, 230, 572, colorDefault)
	}
}

// Draw is called every frame
func (g *Game) Draw(screen *ebiten.Image) {
	if g.mainMenu {
		g.mainMenuDraw(screen)
	} else if g.gameOver {
		g.gameOverDraw(screen)
	} else {
		g.gameDraw(screen)
	}
}

// Layout is called every time the window is resized
func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return g.screenWidth, g.screenHeight
}
