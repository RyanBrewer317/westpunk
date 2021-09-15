package main

import (
	"database/sql"
	"fmt"
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
	"rbrewer.com/drawplayer"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

func init() {
	var err error
	core.PlayerImg, _, err = ebitenutil.NewImageFromFile("player.png")
	if err != nil {
		log.Fatal(err)
	}
	core.PlayerDrawOptions = ebiten.DrawImageOptions{}
	core.GroundImg, _, err = ebitenutil.NewImageFromFile("ground.png")
	if err != nil {
		log.Fatal(err)
	}
	core.GroundDrawOptions = ebiten.DrawImageOptions{}
	core.OakImg, _, err = ebitenutil.NewImageFromFile("tree.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct{}

func (g *Game) Update() error {
	core.Clock += 1
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
		core.PlayerStance = core.ShiftStance(core.WalkRight1, core.WalkRight2)
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && core.PlayerX > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.PlayerX -= 0.1
		core.PlayerStance = core.ShiftStance(core.WalkLeft1, core.WalkLeft2)
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
			chunklet := core.Grid[core.Vertex{X: j, Y: i}]
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
	drawplayer.DrawPlayer(screen, core.PlayerX*core.PixelYardRatio, core.PlayerY*core.PixelYardRatio)
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

func draw_ground(screen *ebiten.Image) {
	core.GroundDrawOptions.GeoM.Reset()
	wi, hi := core.GroundImg.Size()
	core.GroundDrawOptions.GeoM.Scale(core.PlaceWidth*core.PixelYardRatio/float64(wi), core.PlaceHeight*core.PixelYardRatio/float64(hi))
	core.DrawImage(screen, core.GroundImg, core.GroundDrawOptions, -core.VP.X, core.GetPXY(core.GroundY)+core.VP.Y)
}

func draw_oak(screen *ebiten.Image, x float64, y float64) {
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	oak_width_int, oak_height_int := core.OakImg.Size() // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Scale(core.OakWidth*core.PixelYardRatio/float64(oak_width_int), core.OakHeight*core.PixelYardRatio/float64(oak_height_int))
	oakDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PixelYardRatio)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
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
			core.Grid[core.Vertex{X: x, Y: y}] = append(core.Grid[core.Vertex{X: x, Y: y}], core.Oak)
		}
	}

	ebiten.SetWindowSize(int(core.ScreenWidth), int(core.ScreenHeight))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
