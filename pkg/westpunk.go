package main

import (
	"database/sql"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"

	// "math/rand"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/mattn/go-sqlite3"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

var (
	player_img        *ebiten.Image
	ground_img        *ebiten.Image
	oak_img           *ebiten.Image
	playerDrawOptions ebiten.DrawImageOptions
	groundDrawOptions ebiten.DrawImageOptions
)

const screen_height float64 = 540
const screen_width float64 = 810

const pixel_yard_ratio float64 = 100

var player_x float64 = 70
var player_y float64 = 6
var player_x_velocity float64 = 0
var player_y_velocity float64 = 0
var grid map[vertex][]thing

var player_height float64 = 1

const (
	place_width   float64 = 256
	place_height  float64 = 128
	player_width  float64 = 0.5
	ground_height float64 = screen_height / pixel_yard_ratio
	ground_y      float64 = 0
	oak_height    float64 = 5
	oak_width     float64 = 2
)
const (
	torso_width     float64 = 0.25
	torso_height    float64 = 0.5
	head_width      float64 = 0.25
	head_height     float64 = 0.3
	upperarm_height float64 = 0.35
	upperarm_width  float64 = 0.1
	lowerarm_height float64 = 0.35
	lowerarm_width  float64 = 0.1
	thigh_width     float64 = 0.1
	thigh_height    float64 = 0.35
	shin_width      float64 = 0.1
	shin_height     float64 = 0.35
)

type thing int

const (
	oak thing = iota + 1
)

var (
	rest_pose     stance = stance{0.01, 0.01, 5 * math.Pi / 6, -1, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01}
	player_stance stance = rest_pose
)

type stance struct {
	head          float64
	torso         float64
	rightupperarm float64
	leftupperarm  float64
	rightforearm  float64
	leftforearm   float64
	rightthigh    float64
	leftthigh     float64
	rightshin     float64
	leftshin      float64
	rightfoot     float64
	leftfoot      float64
	weapon        float64
}

func init() {
	var err error
	player_img, _, err = ebitenutil.NewImageFromFile("player.png")
	playerDrawOptions = ebiten.DrawImageOptions{}
	ground_img, _, err = ebitenutil.NewImageFromFile("ground.png")
	groundDrawOptions = ebiten.DrawImageOptions{}
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

var ind float64 = 0
var one = 0
var two = 0

func (g *Game) Update() error {
	ind += math.Pi / 180
	player_stance.leftthigh = ind
	player_stance.leftshin = -ind
	player_stance.rightthigh = -ind
	player_stance.rightshin = ind
	rightlegheight := thigh_height*math.Cos(player_stance.torso+player_stance.rightthigh) + shin_height*math.Cos(player_stance.torso+player_stance.rightthigh+player_stance.rightshin)
	leftlegheight := thigh_height*math.Cos(player_stance.torso+player_stance.leftthigh) + shin_height*math.Cos(player_stance.torso+player_stance.leftthigh+player_stance.leftshin)
	player_height = torso_height*math.Cos(player_stance.torso) + math.Max(rightlegheight, leftlegheight)
	if player_y-player_height > ground_y {
		player_y_velocity -= 0.03
	} else if player_y_velocity < 0 {
		// player_y_velocity = 0
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
	rightlegheight := thigh_height*math.Cos(player_stance.torso+player_stance.rightthigh) + shin_height*math.Cos(player_stance.torso+player_stance.rightthigh+player_stance.rightshin)
	leftlegheight := thigh_height*math.Cos(player_stance.torso+player_stance.leftthigh) + shin_height*math.Cos(player_stance.torso+player_stance.leftthigh+player_stance.leftshin)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("player_height: %.1f\nleg height: %.1f\nplayer_y: %.1f\nplayer_y_velocity: %.1f", player_height, math.Max(rightlegheight, leftlegheight), player_y, player_y_velocity))
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
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("trees loaded: %d\n", oak_counter))
	draw_player(screen, player_x*pixel_yard_ratio, player_y*pixel_yard_ratio, g.viewport)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(screen_width), int(screen_height)
}

func dbrun(db *sql.DB, sqlstuff string) {
	_, err := db.Exec(sqlstuff)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlstuff)
		return
	}
}

func dbget(db *sql.DB, sqlstuff string) *sql.Rows {
	rows, err := db.Query(sqlstuff)
	if err != nil {
		log.Fatal(fmt.Sprintf("%q: %s\n", err, sqlstuff))
	}
	return rows
}

func draw_img(screen *ebiten.Image, img *ebiten.Image, drawoptions ebiten.DrawImageOptions, x float64, y float64) {
	drawoptions.GeoM.Translate(x, y)
	screen.DrawImage(img, &drawoptions)
}

func draw_player(screen *ebiten.Image, x float64, y float64, vp viewport) {
	draw_player_left_arm(screen, x, y, vp)
	draw_player_torso(screen, x, y, vp)
	draw_player_head(screen, x, y, vp)
	draw_player_right_arm(screen, x, y, vp)
	draw_player_left_leg(screen, x, y, vp)
	draw_player_right_leg(screen, x, y, vp)
}

func draw_player_head(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.head+player_stance.torso, 2*math.Pi)
	torso_theta := math.Mod(player_stance.torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	if torso_theta < 0 {
		torso_theta += 2 * math.Pi
	}
	neckx := x / pixel_yard_ratio
	necky := (y / pixel_yard_ratio)
	difx, dify := torso_rotation_diff(head_width / 2)
	neckx += difx
	necky += dify
	headx := neckx - (head_width * math.Cos(theta) / 2) + (head_height * math.Sin(theta))
	heady := necky + (head_width * math.Sin(theta) / 2) + (head_height * math.Cos(theta))
	draw_player_piece(screen, 17, 12, 22, 17, playerDrawOptions, headx, heady, head_width, head_height, theta, vp)
}

func torso_rotation_diff(r float64) (float64, float64) {
	theta := math.Mod(player_stance.torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	x := 0.0
	y := 0.0
	if 0 < theta && theta < math.Pi/2 {
		x += r * math.Cos(theta)
		y -= r * math.Sin(theta)
	}
	if math.Pi/2 < theta && theta < math.Pi {
		x -= r * math.Sin(theta-math.Pi/2)
		y -= r * math.Cos(theta-math.Pi/2)
	}
	if math.Pi < theta && theta < 3*math.Pi/2 {
		x -= r * math.Cos(theta-math.Pi)
		y += r * math.Sin(theta-math.Pi)
	}
	if 3*math.Pi/2 < theta && theta < 2*math.Pi {
		x += r * math.Sin(theta-3*math.Pi/2)
		y += r * math.Cos(theta-3*math.Pi/2)
	}
	return x, y
}

func draw_player_torso(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	draw_player_piece(screen, 17, 1, 22, 10, playerDrawOptions, x/pixel_yard_ratio, y/pixel_yard_ratio, torso_width, torso_height, theta, vp)
}

func draw_player_left_arm(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.leftupperarm+player_stance.torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(torso_width)
	leftshoulderx := x/pixel_yard_ratio + difx
	leftshouldery := y/pixel_yard_ratio + dify
	leftupperarmx := leftshoulderx - (upperarm_width * math.Cos(theta) / 2)
	leftupperarmy := leftshouldery + (upperarm_width * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, playerDrawOptions, leftupperarmx, leftupperarmy, upperarm_width, upperarm_height, theta, vp)

	playerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player_stance.leftforearm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftelbowx := leftshoulderx - (upperarm_height * math.Sin(theta))
	leftelbowy := leftshouldery - (upperarm_height * math.Cos(theta))
	var leftforearmx float64
	var leftforearmy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		leftforearmx = leftelbowx - (lowerarm_width/2)*math.Cos(theta2)
		leftforearmy = leftelbowy + (lowerarm_width/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		leftforearmx = leftelbowx + (lowerarm_width/2)*math.Cos(math.Pi-theta2)
		leftforearmy = leftelbowy + (lowerarm_width/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		leftforearmx = leftelbowx + (lowerarm_width/2)*math.Sin(3*math.Pi/2-theta2)
		leftforearmy = leftelbowy - (lowerarm_width/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		leftforearmx = leftelbowx - (lowerarm_width/2)*math.Sin(theta2-3*math.Pi/2)
		leftforearmy = leftelbowy - (lowerarm_width/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 13, 1, 15, 9, playerDrawOptions, leftforearmx, leftforearmy, lowerarm_width, lowerarm_height, theta2, vp)
}

func draw_player_right_arm(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.torso+player_stance.rightupperarm, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(upperarm_width / 2)
	rightshoulderx := x/pixel_yard_ratio + difx
	rightshouldery := y/pixel_yard_ratio + dify
	rightupperarmx := rightshoulderx - (upperarm_width * math.Cos(theta) / 2)
	rightupperarmy := rightshouldery + (upperarm_width * math.Sin(theta) / 2)
	draw_player_piece(screen, 1, 1, 3, 9, playerDrawOptions, rightupperarmx, rightupperarmy, upperarm_width, upperarm_height, theta, vp)

	playerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(player_stance.torso+player_stance.rightupperarm+player_stance.rightforearm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightelbowx := rightshoulderx - (upperarm_height * math.Sin(theta))
	rightelbowy := rightshouldery - (upperarm_height * math.Cos(theta))
	var rightforearmx float64
	var rightforearmy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		rightforearmx = rightelbowx - (lowerarm_width/2)*math.Cos(theta2)
		rightforearmy = rightelbowy + (lowerarm_width/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		rightforearmx = rightelbowx + (lowerarm_width/2)*math.Cos(math.Pi-theta2)
		rightforearmy = rightelbowy + (lowerarm_width/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		rightforearmx = rightelbowx + (lowerarm_width/2)*math.Sin(3*math.Pi/2-theta2)
		rightforearmy = rightelbowy - (lowerarm_width/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		rightforearmx = rightelbowx - (lowerarm_width/2)*math.Sin(theta2-3*math.Pi/2)
		rightforearmy = rightelbowy - (lowerarm_width/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, playerDrawOptions, rightforearmx, rightforearmy, lowerarm_width, lowerarm_height, theta2, vp)
}

func draw_player_left_leg(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.torso+player_stance.leftthigh, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(torso_width - thigh_width/2)
	dify2, difx2 := torso_rotation_diff(torso_height)
	pelvis_left_x := x/pixel_yard_ratio + difx + difx2
	pelvis_left_y := y/pixel_yard_ratio + dify - dify2
	leftthighx := pelvis_left_x - (thigh_width * math.Cos(theta) / 2)
	leftthighy := pelvis_left_y + (thigh_width * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, playerDrawOptions, leftthighx, leftthighy, thigh_width, thigh_height, theta, vp)

	playerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player_stance.leftshin, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftkneex := pelvis_left_x - (thigh_height * math.Sin(theta))
	leftkneey := pelvis_left_y - (thigh_height * math.Cos(theta))
	var leftshinx float64
	var leftshiny float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		leftshinx = leftkneex - (shin_width/2)*math.Cos(theta2)
		leftshiny = leftkneey + (shin_width/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		leftshinx = leftkneex + (shin_width/2)*math.Cos(math.Pi-theta2)
		leftshiny = leftkneey + (shin_width/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		leftshinx = leftkneex + (shin_width/2)*math.Sin(3*math.Pi/2-theta2)
		leftshiny = leftkneey - (shin_width/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		leftshinx = leftkneex - (shin_width/2)*math.Sin(theta2-3*math.Pi/2)
		leftshiny = leftkneey - (shin_width/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, playerDrawOptions, leftshinx, leftshiny, shin_width, shin_height, theta2, vp)
}

func draw_player_right_leg(screen *ebiten.Image, x float64, y float64, vp viewport) {
	playerDrawOptions.GeoM.Reset()
	theta := math.Mod(player_stance.torso+player_stance.rightthigh, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(thigh_width / 2)
	dify2, difx2 := torso_rotation_diff(torso_height)
	pelvis_right_x := x/pixel_yard_ratio + difx + difx2
	pelvis_right_y := y/pixel_yard_ratio + dify - dify2
	rightthighx := pelvis_right_x - (thigh_width * math.Cos(theta) / 2)
	rightthighy := pelvis_right_y + (thigh_width * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, playerDrawOptions, rightthighx, rightthighy, thigh_width, thigh_height, theta, vp)

	playerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player_stance.rightshin, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightkneex := pelvis_right_x - (thigh_height * math.Sin(theta))
	rightkneey := pelvis_right_y - (thigh_height * math.Cos(theta))
	var rightshinx float64
	var rightshiny float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		rightshinx = rightkneex - (shin_width/2)*math.Cos(theta2)
		rightshiny = rightkneey + (shin_width/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		rightshinx = rightkneex + (shin_width/2)*math.Cos(math.Pi-theta2)
		rightshiny = rightkneey + (shin_width/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		rightshinx = rightkneex + (shin_width/2)*math.Sin(3*math.Pi/2-theta2)
		rightshiny = rightkneey - (shin_width/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		rightshinx = rightkneex - (shin_width/2)*math.Sin(theta2-3*math.Pi/2)
		rightshiny = rightkneey - (shin_width/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, playerDrawOptions, rightshinx, rightshiny, shin_width, shin_height, theta2, vp)
}

func draw_player_piece(screen *ebiten.Image, imgx1 int, imgy1 int, imgx2 int, imgy2 int, drawoptions ebiten.DrawImageOptions, igx float64, igy float64, igw float64, igh float64, theta float64, vp viewport) {
	img := ebiten.NewImageFromImage(player_img.SubImage(image.Rect(imgx1, imgy1, imgx2, imgy2)))
	wi, hi := img.Size()
	w := igw * pixel_yard_ratio
	h := igh * pixel_yard_ratio
	drawoptions.GeoM.Scale(w/float64(wi), h/float64(hi))
	drawoptions.GeoM.Rotate(theta)
	draw_img(screen, img, drawoptions, igx*pixel_yard_ratio-vp.x, getpxy(igy)+vp.y)
}

func draw_ground(screen *ebiten.Image, vp viewport) {
	groundDrawOptions.GeoM.Reset()
	wi, hi := ground_img.Size()
	groundDrawOptions.GeoM.Scale(place_width*pixel_yard_ratio/float64(wi), place_height*pixel_yard_ratio/float64(hi))
	draw_img(screen, ground_img, groundDrawOptions, -vp.x, getpxy(ground_y)+vp.y)
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
	rows := dbget(db, "select * from things WHERE placeID = \"place0\"")
	defer rows.Close()
	for rows.Next() {
		var thingID, placeID, location, offset, textureID, thingtype string
		err = rows.Scan(&thingID, &placeID, &location, &offset, &textureID, &thingtype)
		if err != nil {
			log.Fatal(err)
		}
		locsplit := strings.Split(location, " ")
		x, _ := strconv.Atoi(locsplit[0])
		y, _ := strconv.Atoi(locsplit[1])
		switch t := thingtype; t {
		case "oak":
			grid[vertex{x, y}] = append(grid[vertex{x, y}], oak)
		}
	}

	ebiten.SetWindowSize(int(screen_width), int(screen_height))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
