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
var lastFrameOnBeat bool

var score int

const bpm = 170       // Burning - chase and status
const firstBeat = 450 // approximate - need to find a way to calculate this
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
	if upPressed {
		lastFrameOnBeat = onBeat(streamer.Position())
		if lastFrameOnBeat {
			score++
		} else {
			score--
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	drawBeats(screen)
	drawPlayer(screen)
	drawHud(screen)
}

func drawPlayer(screen *ebiten.Image) {
	var c color.Color
	if lastFrameOnBeat {
		c = color.RGBA{0x00, 0xff, 0x00, 0xff}
	} else {
		c = color.RGBA{0xff, 0x00, 0x00, 0xff}
	}
	var viewBoxHeight = 250
	var viewBoxWidth = 30
	vector.StrokeRect(screen, float32(screen.Bounds().Size().X/2)-float32(viewBoxWidth)/2, float32(screen.Bounds().Size().Y/2)-float32(viewBoxHeight/2), float32(viewBoxWidth), float32(viewBoxHeight), 3, c, false)
}

func drawHud(screen *ebiten.Image) {
	currentSongPositionSeconds := streamer.Position() / int(format.SampleRate.N(time.Second))
	totalSongLengthSeconds := streamer.Len() / int(format.SampleRate.N(time.Second))
	ebitenutil.DebugPrint(screen, "Songtime (calculated): "+strconv.Itoa(currentSongPositionSeconds)+"s of "+strconv.Itoa(totalSongLengthSeconds)+"s")
	ebitenutil.DebugPrintAt(screen, "Score: "+strconv.Itoa(score), screen.Bounds().Dx()-100, 0)
}

func drawBeats(screen *ebiten.Image) {
	const pixelsPerMillisecond = 0.1

	currentSongPositionMilliseconds := float32(streamer.Position() / int(format.SampleRate.N(time.Millisecond)))

	// Calculate the time it takes for a beat to move from the right edge of the screen to the left edge
	beatTravelTime := float32(screen.Bounds().Dx()) / float32(pixelsPerMillisecond)

	endSongPositionMilliseconds := float32(currentSongPositionMilliseconds) + beatTravelTime

	for idx, beat := range beats {

		if float32(beat) < currentSongPositionMilliseconds || float32(beat) > endSongPositionMilliseconds {
			continue
		}

		beat1 := (idx == 0 || idx%4 == 0)

		relativeBeatSeconds := float32(beat) - currentSongPositionMilliseconds
		beatPosition := float32(relativeBeatSeconds/beatTravelTime) * float32(screen.Bounds().Dx())
		x := beatPosition
		beatLineStart := float32(screen.Bounds().Size().Y/2) - 100
		beatLineEnd := float32(screen.Bounds().Size().Y/2) + 100
		var beatLineWidth float32
		var beatLineColor color.Color
		if beat1 {
			beatLineWidth = 3
			beatLineColor = color.White
		} else {
			beatLineWidth = 1
			beatLineColor = color.RGBA{0xff, 0x00, 0x00, 0xff}
		}
		vector.StrokeLine(screen, x, beatLineStart, x, beatLineEnd, beatLineWidth, beatLineColor, false)
	}
}

func onBeat(position int) bool {
	const tolerance = 100

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
