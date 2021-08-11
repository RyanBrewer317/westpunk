package main

import (
	"database/sql"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/mattn/go-sqlite3"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

var db *sql.DB
var player_img *ebiten.Image
var ground_img *ebiten.Image
var playerDrawOptions *ebiten.DrawImageOptions
var groundDrawOptions *ebiten.DrawImageOptions

const screen_height float64 = 540
const screen_width float64 = 810

const pixel_yard_ratio float64 = 50

var player_x float64 = 70
var player_y float64 = 6
var player_x_velocity float64 = 0
var player_y_velocity float64 = 0

const place_width float64 = 256
const place_height float64 = 128
const player_height float64 = 2
const player_width float64 = 0.5
const ground_height float64 = screen_height / pixel_yard_ratio
const ground_y float64 = 0

func init() {
	var err error
	player_img, _, err = ebitenutil.NewImageFromFile("gopher.png")
	if err != nil {
		log.Fatal(err)
	}
	playerDrawOptions = &ebiten.DrawImageOptions{}
	ground_img, _, err = ebitenutil.NewImageFromFile("ground.png")
	if err != nil {
		log.Fatal(err)
	}
	groundDrawOptions = &ebiten.DrawImageOptions{}
}

type viewport struct {
	x float64
	y float64
	w float64
	h float64
}

type Game struct {
	viewport viewport
}

func (g *Game) Update() error {
	if player_y-player_height > ground_y {
		player_y_velocity -= 0.03
	} else if player_y_velocity < 0 {
		player_y_velocity = 0
		player_y = player_height
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && player_y-player_height == ground_y {
		player_y_velocity += 0.5
	}
	player_x_velocity *= 0.8
	player_y_velocity *= 0.8
	player_x += player_x_velocity
	player_y += player_y_velocity
	g.viewport.w = screen_width
	g.viewport.h = screen_height
	g.viewport.x = player_x*pixel_yard_ratio - (g.viewport.w / 2)
	g.viewport.y = player_y*pixel_yard_ratio - (g.viewport.h / 2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	draw_player(screen, player_x*pixel_yard_ratio, player_y*pixel_yard_ratio, g.viewport)
	draw_ground(screen, g.viewport)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(screen_width), int(screen_height)
}

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ebiten.SetWindowSize(int(screen_width), int(screen_height))
	ebiten.SetWindowTitle("Westpunk")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

func dbrun(sqlstuff string) {
	_, err := db.Exec(sqlstuff)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlstuff)
		return
	}
}

func draw_player(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	var player_width_int, player_height_int int = player_img.Size()
	playerDrawOptions.GeoM.Scale(player_width*pixel_yard_ratio/float64(player_width_int), player_height*pixel_yard_ratio/float64(player_height_int))
	playerDrawOptions.GeoM.Translate(x-vp.x, getpxy(y/pixel_yard_ratio)+vp.y)
	screen.DrawImage(player_img, playerDrawOptions)
}

func draw_ground(screen *ebiten.Image, vp viewport) {
	groundDrawOptions.GeoM.Reset()
	var ground_width_int, ground_height_int int = ground_img.Size()
	groundDrawOptions.GeoM.Scale(place_width*pixel_yard_ratio/float64(ground_width_int), ground_height*pixel_yard_ratio/float64(ground_height_int))
	groundDrawOptions.GeoM.Translate(-vp.x, getpxy(ground_y)+vp.y)
	screen.DrawImage(ground_img, groundDrawOptions)
}

func getpxy(y float64) float64 {
	return screen_height - (y * pixel_yard_ratio)
}
