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

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/mattn/go-sqlite3"
	"rbrewer.com/core"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

func init() {
	var err error
	core.PlayerImg, _, _ = ebitenutil.NewImageFromFile("player.png")
	core.PlayerDrawOptions = ebiten.DrawImageOptions{}
	core.GroundImg, _, _ = ebitenutil.NewImageFromFile("ground.png")
	core.GroundDrawOptions = ebiten.DrawImageOptions{}
	core.OakImg, _, err = ebitenutil.NewImageFromFile("tree.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct{}

func (g *Game) Update() error {
	rightlegheight := core.UpperLegHeight*math.Cos(core.PlayerStance.Torso+core.PlayerStance.RightUpperLeg) + core.LowerLegHeight*math.Cos(core.PlayerStance.Torso+core.PlayerStance.RightUpperLeg+core.PlayerStance.RightLowerLeg)
	leftlegheight := core.UpperLegHeight*math.Cos(core.PlayerStance.Torso+core.PlayerStance.LeftUpperLeg) + core.LowerLegHeight*math.Cos(core.PlayerStance.Torso+core.PlayerStance.LeftUpperLeg+core.PlayerStance.LeftLowerLeg)
	core.PlayerHeight = core.TorsoHeight*math.Cos(core.PlayerStance.Torso) + math.Max(rightlegheight, leftlegheight)
	if core.PlayerY-core.PlayerHeight > core.GroundY {
		core.PlayerYVelocity -= 0.03
	} else if core.PlayerYVelocity < 0 {
		// core.PlayerYVelocity = 0
		core.PlayerY = core.PlayerHeight
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.PlayerY-core.PlayerHeight == core.GroundY {
		core.PlayerYVelocity += 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) && core.PlayerX < core.PlaceWidth-0.5*core.ScreenWidth/core.PixelYardRatio {
		core.PlayerX += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && core.PlayerX > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.PlayerX -= 0.1
	}
	core.PlayerXVelocity *= 0.8
	core.PlayerYVelocity *= 0.8
	core.PlayerX += core.PlayerXVelocity
	core.PlayerY += core.PlayerYVelocity
	core.VP.W = core.ScreenWidth
	core.VP.H = core.ScreenHeight
	core.VP.X = core.PlayerX*core.PixelYardRatio - (core.VP.W / 2) + (core.PlayerWidth * core.PixelYardRatio / 2)
	core.VP.Y = core.PlayerY*core.PixelYardRatio - (core.VP.H / 2) - (core.PlayerHeight * core.PixelYardRatio / 2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	oak_counter := 0
	draw_ground(screen)
	var chunk_start_y, chunk_start_x int = 0, 0
	chunk_ends_y := int(math.Floor(core.PlaceHeight))
	chunk_ends_x := int(math.Floor(core.PlaceWidth))
	if math.Floor(core.PlayerY)-math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio) > 0 {
		chunk_start_y = int(math.Floor(core.PlayerY) - math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio))
	}
	if math.Floor(core.PlayerY+1)+math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio) < math.Floor(core.PlaceHeight) {
		chunk_ends_y = int(math.Floor(core.PlayerY+1) + math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio))
	}
	if math.Floor(core.PlayerX)-math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio) > 0 {
		chunk_start_x = int(math.Floor(core.PlayerX) - math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio))
	}
	if math.Floor(core.PlayerX+1)+math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio) < math.Floor(core.PlaceWidth) {
		chunk_ends_x = int(math.Floor(core.PlayerX+1) + math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio))
	}
	for i := chunk_start_y; i < chunk_ends_y; i++ {
		for j := chunk_start_x; j < chunk_ends_x; j++ {
			chunklet := core.Grid[core.Vertex{j, i}]
			for k := 0; k < len(chunklet); k++ {
				switch t := chunklet[k]; t {
				case core.Oak:
					oak_counter++
					draw_oak(screen, float64(j)*core.PixelYardRatio, float64(i)*core.PixelYardRatio+core.OakHeight*core.PixelYardRatio)
				}
			}
		}
	}
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("trees loaded: %d\n", oak_counter))
	draw_player(screen, core.PlayerX*core.PixelYardRatio, core.PlayerY*core.PixelYardRatio)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return int(core.ScreenWidth), int(core.ScreenHeight)
}

// func dbrun(db *sql.DB, sqlstuff string) {
// 	_, err := db.Exec(sqlstuff)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlstuff)
// 		return
// 	}
// }

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

func draw_player(screen *ebiten.Image, x float64, y float64) {
	draw_player_left_arm(screen, x, y)
	draw_player_torso(screen, x, y)
	draw_player_head(screen, x, y)
	draw_player_right_arm(screen, x, y)
	draw_player_left_leg(screen, x, y)
	draw_player_right_leg(screen, x, y)
}

func draw_player_head(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Head+core.PlayerStance.Torso, 2*math.Pi)
	torso_theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	if torso_theta < 0 {
		torso_theta += 2 * math.Pi
	}
	neckx := x / core.PixelYardRatio
	necky := (y / core.PixelYardRatio)
	difx, dify := torso_rotation_diff(core.HeadWidth / 2)
	neckx += difx
	necky += dify
	headx := neckx - (core.HeadWidth * math.Cos(theta) / 2) + (core.HeadHeight * math.Sin(theta))
	heady := necky + (core.HeadWidth * math.Sin(theta) / 2) + (core.HeadHeight * math.Cos(theta))
	draw_player_piece(screen, 17, 12, 22, 17, core.PlayerDrawOptions, headx, heady, core.HeadWidth, core.HeadHeight, theta)
}

func torso_rotation_diff(r float64) (float64, float64) {
	theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
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

func draw_player_torso(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	draw_player_piece(screen, 17, 1, 22, 10, core.PlayerDrawOptions, x/core.PixelYardRatio, y/core.PixelYardRatio, core.TorsoWidth, core.TorsoHeight, theta)
}

func draw_player_left_arm(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.LeftUpperArm+core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth)
	leftshoulderx := x/core.PixelYardRatio + difx
	leftshouldery := y/core.PixelYardRatio + dify
	leftupperarmx := leftshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	leftupperarmy := leftshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, core.PlayerDrawOptions, leftupperarmx, leftupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.LeftLowerArm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftelbowx := leftshoulderx - (core.UpperArmHeight * math.Sin(theta))
	leftelbowy := leftshouldery - (core.UpperArmHeight * math.Cos(theta))
	var leftforearmx float64
	var leftforearmy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		leftforearmx = leftelbowx - (core.LowerArmWidth/2)*math.Cos(theta2)
		leftforearmy = leftelbowy + (core.LowerArmWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		leftforearmx = leftelbowx + (core.LowerArmWidth/2)*math.Cos(math.Pi-theta2)
		leftforearmy = leftelbowy + (core.LowerArmWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		leftforearmx = leftelbowx + (core.LowerArmWidth/2)*math.Sin(3*math.Pi/2-theta2)
		leftforearmy = leftelbowy - (core.LowerArmWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		leftforearmx = leftelbowx - (core.LowerArmWidth/2)*math.Sin(theta2-3*math.Pi/2)
		leftforearmy = leftelbowy - (core.LowerArmWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 13, 1, 15, 9, core.PlayerDrawOptions, leftforearmx, leftforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2)
}

func draw_player_right_arm(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperArm, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.UpperArmWidth / 2)
	rightshoulderx := x/core.PixelYardRatio + difx
	rightshouldery := y/core.PixelYardRatio + dify
	rightupperarmx := rightshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	rightupperarmy := rightshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 1, 1, 3, 9, core.PlayerDrawOptions, rightupperarmx, rightupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperArm+core.PlayerStance.RightLowerArm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightelbowx := rightshoulderx - (core.UpperArmHeight * math.Sin(theta))
	rightelbowy := rightshouldery - (core.UpperArmHeight * math.Cos(theta))
	var rightforearmx float64
	var rightforearmy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		rightforearmx = rightelbowx - (core.LowerArmWidth/2)*math.Cos(theta2)
		rightforearmy = rightelbowy + (core.LowerArmWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		rightforearmx = rightelbowx + (core.LowerArmWidth/2)*math.Cos(math.Pi-theta2)
		rightforearmy = rightelbowy + (core.LowerArmWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		rightforearmx = rightelbowx + (core.LowerArmWidth/2)*math.Sin(3*math.Pi/2-theta2)
		rightforearmy = rightelbowy - (core.LowerArmWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		rightforearmx = rightelbowx - (core.LowerArmWidth/2)*math.Sin(theta2-3*math.Pi/2)
		rightforearmy = rightelbowy - (core.LowerArmWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, core.PlayerDrawOptions, rightforearmx, rightforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2)
}

func draw_player_left_leg(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.LeftUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth - core.UpperLegWidth/2)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight)
	pelvis_left_x := x/core.PixelYardRatio + difx + difx2
	pelvis_left_y := y/core.PixelYardRatio + dify - dify2
	leftupperlegx := pelvis_left_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	leftupperlegy := pelvis_left_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, core.PlayerDrawOptions, leftupperlegx, leftupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.LeftLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftkneex := pelvis_left_x - (core.UpperLegHeight * math.Sin(theta))
	leftkneey := pelvis_left_y - (core.UpperLegHeight * math.Cos(theta))
	var leftlowerlegx float64
	var leftlowerlegy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		leftlowerlegx = leftkneex - (core.LowerLegWidth/2)*math.Cos(theta2)
		leftlowerlegy = leftkneey + (core.LowerLegWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		leftlowerlegx = leftkneex + (core.LowerLegWidth/2)*math.Cos(math.Pi-theta2)
		leftlowerlegy = leftkneey + (core.LowerLegWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		leftlowerlegx = leftkneex + (core.LowerLegWidth/2)*math.Sin(3*math.Pi/2-theta2)
		leftlowerlegy = leftkneey - (core.LowerLegWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		leftlowerlegx = leftkneex - (core.LowerLegWidth/2)*math.Sin(theta2-3*math.Pi/2)
		leftlowerlegy = leftkneey - (core.LowerLegWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, core.PlayerDrawOptions, leftlowerlegx, leftlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2)
}

func draw_player_right_leg(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.UpperLegWidth / 2)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight)
	pelvis_right_x := x/core.PixelYardRatio + difx + difx2
	pelvis_right_y := y/core.PixelYardRatio + dify - dify2
	rightupperlegx := pelvis_right_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	rightupperlegy := pelvis_right_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 5, 1, 7, 9, core.PlayerDrawOptions, rightupperlegx, rightupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.RightLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightkneex := pelvis_right_x - (core.UpperLegHeight * math.Sin(theta))
	rightkneey := pelvis_right_y - (core.UpperLegHeight * math.Cos(theta))
	var rightlowerlegx float64
	var rightlowerlegy float64
	if 0 < theta2 && theta2 < math.Pi/2 {
		rightlowerlegx = rightkneex - (core.LowerLegWidth/2)*math.Cos(theta2)
		rightlowerlegy = rightkneey + (core.LowerLegWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 < theta2 && theta2 < math.Pi {
		rightlowerlegx = rightkneex + (core.LowerLegWidth/2)*math.Cos(math.Pi-theta2)
		rightlowerlegy = rightkneey + (core.LowerLegWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi < theta2 && theta2 < 3*math.Pi/2 {
		rightlowerlegx = rightkneex + (core.LowerLegWidth/2)*math.Sin(3*math.Pi/2-theta2)
		rightlowerlegy = rightkneey - (core.LowerLegWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 < theta2 && theta2 < 2*math.Pi {
		rightlowerlegx = rightkneex - (core.LowerLegWidth/2)*math.Sin(theta2-3*math.Pi/2)
		rightlowerlegy = rightkneey - (core.LowerLegWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 9, 1, 11, 9, core.PlayerDrawOptions, rightlowerlegx, rightlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2)
}

func draw_player_piece(screen *ebiten.Image, imgx1 int, imgy1 int, imgx2 int, imgy2 int, drawoptions ebiten.DrawImageOptions, igx float64, igy float64, igw float64, igh float64, theta float64) {
	img := ebiten.NewImageFromImage(core.PlayerImg.SubImage(image.Rect(imgx1, imgy1, imgx2, imgy2)))
	wi, hi := img.Size()
	w := igw * core.PixelYardRatio
	h := igh * core.PixelYardRatio
	drawoptions.GeoM.Scale(w/float64(wi), h/float64(hi))
	drawoptions.GeoM.Rotate(theta)
	draw_img(screen, img, drawoptions, igx*core.PixelYardRatio-core.VP.X, getpxy(igy)+core.VP.Y)
}

func draw_ground(screen *ebiten.Image) {
	core.GroundDrawOptions.GeoM.Reset()
	wi, hi := core.GroundImg.Size()
	core.GroundDrawOptions.GeoM.Scale(core.PlaceWidth*core.PixelYardRatio/float64(wi), core.PlaceHeight*core.PixelYardRatio/float64(hi))
	draw_img(screen, core.GroundImg, core.GroundDrawOptions, -core.VP.X, getpxy(core.GroundY)+core.VP.Y)
}

func draw_oak(screen *ebiten.Image, x float64, y float64) {
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	oak_width_int, oak_height_int := core.OakImg.Size() // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Scale(core.OakWidth*core.PixelYardRatio/float64(oak_width_int), core.OakHeight*core.PixelYardRatio/float64(oak_height_int))
	oakDrawOptions.GeoM.Translate(x-core.VP.X, getpxy(y/core.PixelYardRatio)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
}

func getpxy(y float64) float64 {
	return core.ScreenHeight - (y * core.PixelYardRatio)
}

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	core.Grid = make(map[core.Vertex][]core.Thing)
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
			core.Grid[core.Vertex{x, y}] = append(core.Grid[core.Vertex{x, y}], core.Oak)
		}
	}

	ebiten.SetWindowSize(int(core.ScreenWidth), int(core.ScreenHeight))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
