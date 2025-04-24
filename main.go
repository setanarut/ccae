package main

import (
	_ "embed"
	"image"
	"image/color"
	"log"
	"math/rand/v2"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

//go:embed cca.kage
var ccaShaderProgram []byte

//ebitengine:shaderfile cca.kage

var imgOffsetX, imgOffsetY float64 = 280, 0.0
var (
	windowW    int = 900
	windowH    int = 480
	OffscreenW int = windowW - int(imgOffsetX)
	OffScreenH int = 512
)

const (
	Moore   = 0
	Neumann = 1
)

var game = &Game{}
var shader *ebiten.Shader
var rectShaderOptions = &ebiten.DrawRectShaderOptions{}
var offScreen *ebiten.Image

var dop = ebiten.DrawImageOptions{}
var ui debugui.DebugUI
var accumulator float64

// CCA Rules
var (
	neighborhoodRange int = 1
	threshold         int = 1
	states            int = 24
	neighborhood      int = Moore
)

var BrushRadius float64 = 32
var BrushValue float64 = 1.0
var BrushNoise int = 1
var SpeedSlider float64 = 60
var NeighborhoodCheck bool = true
var isMoore bool = true
var isNeumann bool = false
var isNoiseBrush bool = true
var isValueBrush bool = false
var cx, cy int

type Game struct{}

func init() {
	dop.GeoM.Translate(imgOffsetX, imgOffsetY)
	rectShaderOptions.Images[0] = ebiten.NewImage(OffscreenW, OffScreenH)
	offScreen = ebiten.NewImage(OffscreenW, OffScreenH)
	// rectShaderOptions.GeoM.Translate(130, 0)
	FillRandom(rectShaderOptions.Images[0])
	var err error
	shader, err = ebiten.NewShader(ccaShaderProgram)
	if err != nil {
		log.Fatal(err)
	}
}
func (g *Game) Update() error {

	cx, cy = ebiten.CursorPosition()

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		BrushRadius = max(BrushRadius-0.6, 1.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		BrushRadius = min(BrushRadius+0.6, 300.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		BrushValue = min(BrushValue+0.03, 1.0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		BrushValue = max(BrushValue-0.03, 0.0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		FillRandom(rectShaderOptions.Images[0])
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		rectShaderOptions.Images[0].Clear()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		rectShaderOptions.Images[0].Fill(color.White)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		BrushNoise = 1 - BrushNoise
		isNoiseBrush = !isNoiseBrush
		isValueBrush = !isValueBrush
	}

	rectShaderOptions.Uniforms["BrushRadius"] = BrushRadius
	rectShaderOptions.Uniforms["BrushValue"] = BrushValue
	rectShaderOptions.Uniforms["BrushNoise"] = BrushNoise
	rectShaderOptions.Uniforms["NeighborhoodRange"] = neighborhoodRange
	rectShaderOptions.Uniforms["Threshold"] = float64(threshold)
	rectShaderOptions.Uniforms["States"] = float64(states)
	rectShaderOptions.Uniforms["Neighborhood"] = neighborhood
	rectShaderOptions.Uniforms["Cursor"] = []float64{float64(cx) - imgOffsetX, float64(cy) - imgOffsetY}
	rectShaderOptions.Uniforms["Tick"] = ebiten.Tick()
	rectShaderOptions.Uniforms["Time"] = float64(ebiten.Tick()) / float64(ebiten.TPS())

	if _, err := ui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Rule", image.Rect(0, 0, 140, windowH), func(layout debugui.ContainerLayout) {
			ctx.Text("Range")
			ctx.Slider(&neighborhoodRange, 1, 5, 1)
			ctx.Text("Threshold")
			ctx.Slider(&threshold, 1, 32, 1)
			ctx.Text("States")
			ctx.Slider(&states, 1, 32, 1)
			ctx.Text("Neighborhood")

			ctx.Checkbox(&isMoore, "Moore").On(func() {

				isNeumann = !isNeumann
				neighborhood = 0

			})
			ctx.Checkbox(&isNeumann, "Neumann").On(func() {

				neighborhood = 1
				isMoore = !isMoore

			})

			ctx.Text("Rule Presets")

			ctx.Button("1/1/14/N").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 1 && threshold == 1 && states == 14 && neighborhood == 1 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 1, 1
				rectShaderOptions.Uniforms["Threshold"], threshold = 1, 1
				rectShaderOptions.Uniforms["States"], states = 14, 14
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 1, 1
				isMoore = false
				isNeumann = true
				SpeedSlider = 55
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("1/1/15/N").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 1 && threshold == 1 && states == 15 && neighborhood == 1 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 1, 1
				rectShaderOptions.Uniforms["Threshold"], threshold = 1, 1
				rectShaderOptions.Uniforms["States"], states = 15, 15
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 1, 1
				isMoore = false
				isNeumann = true
				SpeedSlider = 55
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("1/1/24/M").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 1 && threshold == 1 && states == 24 && neighborhood == 0 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 1, 1
				rectShaderOptions.Uniforms["Threshold"], threshold = 1.0, 1
				rectShaderOptions.Uniforms["States"], states = 24.0, 24
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 0, 0
				isMoore = true
				isNeumann = false
				SpeedSlider = 50
				FillRandom(rectShaderOptions.Images[0])
			})

			ctx.Button("1/3/3/M").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 1 && threshold == 3 && states == 3 && neighborhood == 0 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 1, 1
				rectShaderOptions.Uniforms["Threshold"], threshold = 3, 3
				rectShaderOptions.Uniforms["States"], states = 3, 3
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 0, 0
				isMoore = true
				isNeumann = false
				SpeedSlider = 45
				FillRandom(rectShaderOptions.Images[0])
			})

			ctx.Button("3/4/5/N").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 3 && threshold == 4 && states == 5 && neighborhood == 1 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 3, 3
				rectShaderOptions.Uniforms["Threshold"], threshold = 4, 4
				rectShaderOptions.Uniforms["States"], states = 5, 5
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 1, 1
				isMoore = false
				isNeumann = true
				SpeedSlider = 50
				FillRandom(rectShaderOptions.Images[0])
			})

			ctx.Button("2/2/6/N").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 2 && threshold == 2 && states == 6 && neighborhood == 1 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 2, 2
				rectShaderOptions.Uniforms["Threshold"], threshold = 2, 2
				rectShaderOptions.Uniforms["States"], states = 6, 6
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 1, 1
				isMoore = false
				isNeumann = true
				SpeedSlider = 22
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("3/5/8/M").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 3 && threshold == 5 && states == 8 && neighborhood == 0 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 3, 3
				rectShaderOptions.Uniforms["Threshold"], threshold = 5, 5
				rectShaderOptions.Uniforms["States"], states = 8, 8
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 0, 0
				isMoore = true
				isNeumann = false
				SpeedSlider = 22
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("2/3/5/N").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 2 && threshold == 3 && states == 5 && neighborhood == 1 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 2, 2
				rectShaderOptions.Uniforms["Threshold"], threshold = 3, 3
				rectShaderOptions.Uniforms["States"], states = 5, 5
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 1, 1
				isMoore = false
				isNeumann = true
				SpeedSlider = 50
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("1/3/4/M").On(func() {
				accumulator = 0
				// Mevcut değerler zaten aynıysa işlem yapmayalım
				if neighborhoodRange == 1 && threshold == 3 && states == 4 && neighborhood == 0 {
					return
				}
				rectShaderOptions.Uniforms["NeighborhoodRange"], neighborhoodRange = 1, 1
				rectShaderOptions.Uniforms["Threshold"], threshold = 3, 3
				rectShaderOptions.Uniforms["States"], states = 4, 4
				rectShaderOptions.Uniforms["Neighborhood"], neighborhood = 0, 0
				isMoore = true
				isNeumann = false
				SpeedSlider = 50
				FillRandom(rectShaderOptions.Images[0])
			})

		})

		ctx.Window("Controls", image.Rect(140, 0, 280, windowH), func(layout debugui.ContainerLayout) {

			ctx.Text("Speed")
			ctx.SliderF(&SpeedSlider, 1.0, 60.0, 0.1, 1)

			ctx.Text("Brush radius (W/S)")
			ctx.SliderF(&BrushRadius, 1.0, 300, 1, 0)

			ctx.Text("Brush Value (A/D)")
			ctx.SliderF(&BrushValue, 0, 1, 0.003, 2)

			ctx.Checkbox(&isNoiseBrush, "Noise Brush (F)").On(func() {

				isValueBrush = !isValueBrush
				BrushNoise = 1

			})
			ctx.Checkbox(&isValueBrush, "Value Brush (F)").On(func() {

				BrushNoise = 0
				isNoiseBrush = !isNoiseBrush

			})

			ctx.Button("Fill Noise (Q)").On(func() {
				FillRandom(rectShaderOptions.Images[0])
			})
			ctx.Button("Clear (E)").On(func() {
				rectShaderOptions.Images[0].Clear()
			})

		})
		return nil
	}); err != nil {
		return err
	}

	// Birikimli değeri güncelle
	incrementAmount := SpeedSlider / 60.0
	accumulator += incrementAmount
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.Gray{100})

	if accumulator >= 1.0 {
		offScreen.DrawRectShader(offScreen.Bounds().Dx(), offScreen.Bounds().Dy(), shader, rectShaderOptions)
		// Copy the result from the screen to the buffer0
		rectShaderOptions.Images[0].DrawImage(offScreen, nil)
		accumulator -= float64(int(accumulator))
	}

	screen.DrawImage(offScreen, &dop)
	ui.Draw(screen)

	vector.DrawFilledRect(screen, float32(cx-2), float32(cy-2), 4, 4, colornames.Yellow, false)
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return windowW, windowH
}

func FillRandom(img *ebiten.Image) {
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			img.Set(x, y, color.Gray{uint8(rand.IntN(254))})
		}
	}
}

func main() {
	// ebiten.SetTPS(15)
	rectShaderOptions.Uniforms = map[string]any{}
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	// ebiten.SetCursorShape(ebiten.CursorShapeCrosshair)
	op := ebiten.RunGameOptions{
		DisableHiDPI: true,
	}
	ebiten.SetScreenClearedEveryFrame(false)
	// ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowTitle("CCA")
	if err := ebiten.RunGameWithOptions(game, &op); err != nil {
		log.Fatal(err)
	}
}

func MapRange(v, a, b, c, d float32) float32 {
	return (v-a)/(b-a)*(d-c) + c
}
