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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/*
var assets embed.FS
var streamer beep.StreamSeekCloser
var upPressed bool
var format beep.Format
var beats []int

var score int

const bpm = 170       // Burning - chase and status
const firstBeat = 500 // approximate - need to find a way to calculate this
const beatInterval = 60000 / bpm

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
	format = format1

	speaker.Init(format1.SampleRate, format1.SampleRate.N(time.Second/10))
	log.Println("Samples per second = ", format1.SampleRate.N(time.Second/10))
	log.Println("Len = ", streamer1.Len())
	return streamer1
}

func calculateBeats(streamer beep.StreamSeekCloser) []int {

	//positionPerSecond := (streamer.Len() / sampleRatePerMillisecond) * 1000

	// 500 is the first beat
	// 60000 / bpm is the interval between beats
	// the number of beats is the length of the song in seconds
	// so the number of beats is the length of the song in seconds * 60000 / bpm

	// song is 60*4 seconds long
	// ~170*4= 680 beats

	var beats []int
	for i := 0; i < 675; i++ {
		beats = append(beats, firstBeat+(beatInterval*i))
	}
	return beats

}

func playMusic(streamer beep.StreamSeekCloser) {
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		log.Println("Music finished")
		streamer.Close()
	})))
}

func (g *Game) Update() error {
	upPressed = ebiten.IsKeyPressed(ebiten.KeySpace)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	const pixelsPerSecond = 100 // adjust this value as needed
	//deltaTime := 1.0 / 60.0  // assuming a frame rate of 60 FPS

	if upPressed {
		ebitenutil.DebugPrintAt(screen, "Up pressed", 0, 100)
		var c color.Color
		if onBeat(streamer.Position()) {
			c = color.RGBA{0x00, 0xff, 0x00, 0xff}
			score++
		} else {
			c = color.RGBA{0xff, 0x00, 0x00, 0xff}
			score--
		}
		screen.Fill(c)
	} else {
		screen.Fill(color.Black)
		ebitenutil.DebugPrintAt(screen, "No key pressed", 0, 100)
	}
	currentSongPositionSeconds := streamer.Position() / int(format.SampleRate.N(time.Second))
	totalSongLengthSeconds := streamer.Len() / int(format.SampleRate.N(time.Second))
	ebitenutil.DebugPrint(screen, "Songtime (calculated): "+strconv.Itoa(currentSongPositionSeconds)+"s of "+strconv.Itoa(totalSongLengthSeconds)+"s")
	ebitenutil.DebugPrintAt(screen, strconv.FormatBool(onBeat(streamer.Position())), 0, 110)
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(score), screen.Bounds().Dx()-100, 0)

	//vector.StrokeLine(screen, 0, 300, 0, 500, 1, color.White, false)
	//ebitenutil.DrawLine(screen, 10, 10, 500, 10, color.White)
	totalLength := screen.Bounds().Dx() * 10
	totalRemaining := totalSongLengthSeconds - currentSongPositionSeconds
	startPosition := (screen.Bounds().Dx() / 2) - (totalLength * (currentSongPositionSeconds / totalSongLengthSeconds)) + 100

	log.Println("Total length = ", totalLength)
	log.Println("Total song length = ", totalSongLengthSeconds)
	log.Println("Total remaining = ", totalRemaining)
	log.Println("Start position = ", startPosition)

	// Calculate the start and end times of the visible part of the song
	startSongPositionSeconds := float32(currentSongPositionSeconds)
	endSongPositionSeconds := startSongPositionSeconds + float32(totalLength)/float32(pixelsPerSecond)

	// middle of the screen - ratio of elapsed length to total song length
	for _, beat := range beats {

		// Calculate the beat in seconds
		beatSeconds := float32(beat) / 1000.0

		// Skip this beat if it's not currently visible
		if beatSeconds < startSongPositionSeconds || beatSeconds > endSongPositionSeconds {
			continue
		}

		// Calculate the position of the beat relative to the start of the song
		beatPosition := (float32(beatSeconds) / float32(totalSongLengthSeconds)) * float32(totalLength)

		log.Println("Beat = ", beat)
		// Calculate the x-coordinate of the beat
		x := float32(startPosition) - beatPosition - float32(currentSongPositionSeconds)
		log.Println("Beat x = ", x)
		vector.StrokeLine(screen, x, 305, x, 315, 1, color.White, false)
	}
	// panic("die")
	vector.StrokeRect(screen, float32(screen.Bounds().Size().X/2), float32(screen.Bounds().Size().Y/2), 10, 10, 1, color.White, false)

}

func onBeat(position int) bool {
	const tolerance = 45

	var sampleRatePerMillisecond = format.SampleRate.N(time.Millisecond)

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
	beats = calculateBeats(streamer)
	playMusic(streamer)

	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}
}
