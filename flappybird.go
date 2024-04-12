package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

const (
	screenWidth  = 320
	screenHeight = 480
	birdSize     = 20
	gravity      = 0.5
	jumpVelocity = -8
	pipeWidth    = 60
	pipeGap      = 100
)

type Pipe struct {
	X      float64
	GapY   float64
	Passed bool
}

type Game struct {
	birdY     float64
	velocity  float64
	pipes     []Pipe
	score     int
	highScore int
	gameState string
	frameCount int
}

func loadHighScore() int {
	content, err := os.ReadFile("highscore.txt")
	if err != nil {
		return 0 // Assume no high score if file doesn't exist or can't be read
	}
	highScore, err := strconv.Atoi(string(content))
	if err != nil {
		return 0 // Default to 0 if there's an error converting the content to an integer
	}
	return highScore
}

func saveHighScore(highScore int) {
	err := os.WriteFile("highscore.txt", []byte(strconv.Itoa(highScore)), 0644)
	if err != nil {
		log.Printf("Error saving high score: %v", err)
	}
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		birdY:      screenHeight / 2.0,
		gameState:  "start",
		highScore:  loadHighScore(),
		score:      0,
	}
}

func (g *Game) Update() error {
	g.frameCount++

	switch g.gameState {
	case "start":
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.gameState = "play"
		}
	case "play":
		g.velocity += gravity
		g.birdY += g.velocity

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.velocity = jumpVelocity
		}

		if g.birdY > screenHeight-birdSize {
			g.gameState = "gameover"
		}

		if g.birdY < 0 {
			g.birdY = 0
			g.velocity = 0
		}

		if g.frameCount%180 == 0 {
			g.pipes = append(g.pipes, Pipe{
				X:    screenWidth,
				GapY: screenHeight/2 + float64((rand.Intn(3)-1)*60),
			})
		}

		for i := range g.pipes {
			pipe := &g.pipes[i]
			pipe.X -= 2

			if !pipe.Passed && pipe.X < screenWidth/2-pipeWidth/2 {
				pipe.Passed = true
				g.score++
			}

			birdRect := image.Rect(
				int(screenWidth/2-birdSize/2), int(g.birdY-birdSize/2),
				int(screenWidth/2+birdSize/2), int(g.birdY+birdSize/2),
			)

			upperPipeRect := image.Rect(
				int(pipe.X), 0,
				int(pipe.X+pipeWidth), int(pipe.GapY-pipeGap/2),
			)

			lowerPipeRect := image.Rect(
				int(pipe.X), int(pipe.GapY+pipeGap/2),
				int(pipe.X+pipeWidth), screenHeight,
			)

			if birdRect.Overlaps(upperPipeRect) || birdRect.Overlaps(lowerPipeRect) {
				g.gameState = "gameover"
			}
		}

		g.pipes = filterPipes(g.pipes)

	case "gameover":
		if g.score > g.highScore {
			g.highScore = g.score
			saveHighScore(g.highScore)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			*g = *NewGame()
		}
	}
	return nil
}

func filterPipes(pipes []Pipe) []Pipe {
    var filtered []Pipe
    for _, p := range pipes {
        if p.X+pipeWidth > 0 { // Corrected from p.pipeWidth to pipeWidth
            filtered = append(filtered, p)
        }
    }
    return filtered
}


func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Skyblue)

	switch g.gameState {
	case "start":
		ebitenutil.DebugPrint(screen, fmt.Sprintf("High Score: %d\nPress Enter to start", g.highScore))
	case "play", "gameover":
		ebitenutil.DrawRect(screen, screenWidth/2-birdSize/2, g.birdY-birdSize/2, birdSize, birdSize, colornames.Yellow)

		for _, p := range g.pipes {
			upperPipeBottomY := p.GapY - pipeGap/2
			lowerPipeTopY := p.GapY + pipeGap/2
			ebitenutil.DrawRect(screen, p.X, 0, pipeWidth, upperPipeBottomY, colornames.Green)
			ebitenutil.DrawRect(screen, p.X, lowerPipeTopY, pipeWidth, screenHeight-lowerPipeTopY, colornames.Green)
		}

		if g.gameState == "gameover" {
			msg := fmt.Sprintf("Game Over! Score: %d\nPress Enter to restart", g.score)
			ebitenutil.DebugPrint(screen, msg)
		} else {
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d", g.score))
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Flappy Bird")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
