package main

import (
	"embed"
	"image/color"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed assets/*
var assets embed.FS
var streamer beep.StreamSeekCloser
var upPressed bool
var format beep.Format

type Game struct{}

func loadMusic(music string) beep.StreamSeekCloser {
	f, err := os.Open(music)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	streamer1, format1, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	speaker.Init(format1.SampleRate, format1.SampleRate.N(time.Second/10))
	log.Println("Samples per second = ", format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(streamer1, beep.Callback(func() {
		log.Println("Music finished")
		streamer1.Close()
	})))
	format = format1
	return streamer1
}

func (g *Game) Update() error {
	upPressed = ebiten.IsKeyPressed(ebiten.KeyUp)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	if upPressed {
		ebitenutil.DebugPrintAt(screen, "Up pressed", 0, 100)
		c := color.RGBA{0xff, 0x00, 0x00, 0xff}
		if onBeat(streamer.Position()) {
			c = color.RGBA{0x00, 0xff, 0x00, 0xff}
		}
		screen.Fill(c)
	} else {
		screen.Fill(color.Black)
		ebitenutil.DebugPrintAt(screen, "No key pressed", 0, 100)
	}
	ebitenutil.DebugPrint(screen, strconv.Itoa(streamer.Position()))
	ebitenutil.DebugPrintAt(screen, strconv.FormatBool(onBeat(streamer.Position())), 0, 110)
	ebitenutil.DebugPrintAt(screen, strconv.Itoa(streamer.Len()), 0, 120)
}

func onBeat(position int) bool {
	const bpm = 170 // chase and status (burning)
	const firstBeat = 500
	const beatInterval = 60000 / bpm
	const tolerance = 100

	// this doesn't work because position is in bytes, not milliseconds
	var sampleRatePerMillisecond = format.SampleRate.N(time.Millisecond)
	if sampleRatePerMillisecond <= 0 {
		log.Println("Samples too low")
		return false
	}
	positionInMilliseconds := position / sampleRatePerMillisecond
	offSet := positionInMilliseconds - firstBeat
	if offSet < 0 {
		log.Println("Offset below 0")
		return false
	}

	distanceToBeat := offSet % beatInterval
	return distanceToBeat < tolerance || distanceToBeat > beatInterval-tolerance
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	g := &Game{}
	streamer = loadMusic("./assets/music/12026292_Burning_(Original Mix).mp3")
	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
