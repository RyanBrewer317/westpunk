package main

import (
	"database/sql"
	"fmt"
	_ "image/png"
	"log"
	"math"
	"math/rand"

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
var oak_img *ebiten.Image
var playerDrawOptions *ebiten.DrawImageOptions
var groundDrawOptions *ebiten.DrawImageOptions

const screen_height float64 = 540
const screen_width float64 = 810

const pixel_yard_ratio float64 = 50

var player_x float64 = 70
var player_y float64 = 6
var player_x_velocity float64 = 0
var player_y_velocity float64 = 0
var grid map[vertex][]thing

const place_width float64 = 256
const place_height float64 = 128
const player_height float64 = 2
const player_width float64 = 0.5
const ground_height float64 = screen_height / pixel_yard_ratio
const ground_y float64 = 0
const oak_height float64 = 5
const oak_width float64 = 2

type thing int

const (
	oak thing = iota + 1
)

func init() {
	var err error
	player_img, _, err = ebitenutil.NewImageFromFile("gopher.png")
	playerDrawOptions = &ebiten.DrawImageOptions{}
	ground_img, _, err = ebitenutil.NewImageFromFile("ground.png")
	groundDrawOptions = &ebiten.DrawImageOptions{}
	oak_img, _, err = ebitenutil.NewImageFromFile("tree.png")
	if err != nil {
		log.Fatal(err)
	}
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

type vertex struct {
	x, y int
}

func (g *Game) Update() error {
	if player_y-player_height > ground_y {
		player_y_velocity -= 0.03
	} else if player_y_velocity < 0 {
		player_y_velocity = 0
		player_y = player_height
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && player_y-player_height == ground_y {
		player_y_velocity += 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) && player_x < place_width-0.5*screen_width/pixel_yard_ratio {
		player_x += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && player_x > 0.5*screen_width/pixel_yard_ratio {
		player_x -= 0.1
	}
	player_x_velocity *= 0.8
	player_y_velocity *= 0.8
	player_x += player_x_velocity
	player_y += player_y_velocity
	g.viewport.w = screen_width
	g.viewport.h = screen_height
	g.viewport.x = player_x*pixel_yard_ratio - (g.viewport.w / 2) + (player_width * pixel_yard_ratio / 2)
	g.viewport.y = player_y*pixel_yard_ratio - (g.viewport.h / 2) - (player_height * pixel_yard_ratio / 2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	oak_counter := 0
	draw_ground(screen, g.viewport)
	var chunk_start_y, chunk_start_x int = 0, 0
	chunk_ends_y := int(math.Floor(place_height))
	chunk_ends_x := int(math.Floor(place_width))
	if math.Floor(player_y)-math.Floor(0.75*screen_height/pixel_yard_ratio) > 0 {
		chunk_start_y = int(math.Floor(player_y) - math.Floor(0.75*screen_height/pixel_yard_ratio))
	}
	if math.Floor(player_y+1)+math.Floor(0.75*screen_height/pixel_yard_ratio) < math.Floor(place_height) {
		chunk_ends_y = int(math.Floor(player_y+1) + math.Floor(0.75*screen_height/pixel_yard_ratio))
	}
	if math.Floor(player_x)-math.Floor(0.75*screen_width/pixel_yard_ratio) > 0 {
		chunk_start_x = int(math.Floor(player_x) - math.Floor(0.75*screen_width/pixel_yard_ratio))
	}
	if math.Floor(player_x+1)+math.Floor(0.75*screen_width/pixel_yard_ratio) < math.Floor(place_width) {
		chunk_ends_x = int(math.Floor(player_x+1) + math.Floor(0.75*screen_width/pixel_yard_ratio))
	}
	for i := chunk_start_y; i < chunk_ends_y; i++ {
		for j := chunk_start_x; j < chunk_ends_x; j++ {
			chunklet := grid[vertex{j, i}]
			// fmt.Println(j, i, chunklet)
			for k := 0; k < len(chunklet); k++ {
				switch t := chunklet[k]; t {
				case oak:
					oak_counter++
					draw_oak(screen, float64(j)*pixel_yard_ratio, float64(i)*pixel_yard_ratio+oak_height*pixel_yard_ratio, g.viewport)
				}
			}
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("trees loaded: %d\n", oak_counter))
	draw_player(screen, player_x*pixel_yard_ratio, player_y*pixel_yard_ratio, g.viewport)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(screen_width), int(screen_height)
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
	var player_width_int, player_height_int int = player_img.Size() // optimizable by moving scaling somewhere else that's not called every tick
	playerDrawOptions.GeoM.Scale(player_width*pixel_yard_ratio/float64(player_width_int), player_height*pixel_yard_ratio/float64(player_height_int))
	playerDrawOptions.GeoM.Translate(x-vp.x, getpxy(y/pixel_yard_ratio)+vp.y)
	screen.DrawImage(player_img, playerDrawOptions)
}

func draw_ground(screen *ebiten.Image, vp viewport) {
	groundDrawOptions.GeoM.Reset()
	ground_width_int, ground_height_int := ground_img.Size() // optimizable by moving scaling somewhere else that's not called every tick
	groundDrawOptions.GeoM.Scale(place_width*pixel_yard_ratio/float64(ground_width_int), ground_height*pixel_yard_ratio/float64(ground_height_int))
	groundDrawOptions.GeoM.Translate(-vp.x, getpxy(ground_y)+vp.y)
	screen.DrawImage(ground_img, groundDrawOptions)
}

func draw_oak(screen *ebiten.Image, x float64, y float64, vp viewport) {
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	oak_width_int, oak_height_int := oak_img.Size() // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Scale(oak_width*pixel_yard_ratio/float64(oak_width_int), oak_height*pixel_yard_ratio/float64(oak_height_int))
	oakDrawOptions.GeoM.Translate(x-vp.x, getpxy(y/pixel_yard_ratio)+vp.y)
	screen.DrawImage(oak_img, oakDrawOptions)
}

func getpxy(y float64) float64 {
	return screen_height - (y * pixel_yard_ratio)
}

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	grid = make(map[vertex][]thing)
	tree_xs := [int(place_width)]int{}
	for i := 0; i < int(place_width); i++ {
		tree_xs[i] = i
	}
	rand.Shuffle(int(place_width), func(i, j int) { tree_xs[i], tree_xs[j] = tree_xs[j], tree_xs[i] })
	for i := 0; i < 20; i++ {
		grid[vertex{tree_xs[i], 0}] = append(grid[vertex{tree_xs[i], 0}], oak)
	}

	ebiten.SetWindowSize(int(screen_width), int(screen_height))
	ebiten.SetWindowTitle("Westpunk")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
