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
	"rbrewer.com/stances"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

func init() {
	core.MainPlayer.WalkingStanceTo = stances.RestRight2
	core.MainPlayer.WalkingStanceFrom = stances.RestRight1
	var err error
	core.PlayerImg, _, err = ebitenutil.NewImageFromFile("spritesheet.png")
	if err != nil {
		log.Fatal(err)
	}
	core.MainPlayer.DrawOptions = ebiten.DrawImageOptions{}
	core.GrassLayerImg, _, err = ebitenutil.NewImageFromFile("grasslayerground.png")
	if err != nil {
		log.Fatal(err)
	}
	core.DirtLayerImg, _, err = ebitenutil.NewImageFromFile("dirtlayerground.png")
	if err != nil {
		log.Fatal(err)
	}
	core.OakImg, _, err = ebitenutil.NewImageFromFile("tree.png")
	if err != nil {
		log.Fatal(err)
	}
	core.OakLogImg, _, err = ebitenutil.NewImageFromFile("oaklog.png")
	if err != nil {
		log.Fatal(err)
	}
	core.BackgroundImg, _, err = ebitenutil.NewImageFromFile("background.png")
	if err != nil {
		log.Fatal(err)
	}
	core.BackgroundDrawOptions = ebiten.DrawImageOptions{}
	wi, hi := core.BackgroundImg.Size()
	core.BackgroundDrawOptions.GeoM.Scale(core.ScreenWidth/float64(wi), core.ScreenHeight/float64(hi))
}

type Game struct{}

func (g *Game) Update() error {
	core.MainPlayer.WalkingAnimationFrame += 1

	rightlegheight := core.UpperLegHeight*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.RightUpperLeg) + core.LowerLegHeight*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.RightUpperLeg+core.MainPlayer.Stance.RightLowerLeg)
	leftlegheight := core.UpperLegHeight*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.LeftUpperLeg) + core.LowerLegHeight*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.LeftUpperLeg+core.MainPlayer.Stance.LeftLowerLeg)
	core.MainPlayer.Height = core.TorsoHeight*math.Cos(core.MainPlayer.Stance.Torso) + math.Max(rightlegheight, leftlegheight)

	if core.MainPlayer.Y-core.MainPlayer.Height > core.GroundY {
		core.MainPlayer.Gravity_dy -= 0.03
	} else {
		core.MainPlayer.Y = core.MainPlayer.Height
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.MainPlayer.Y-core.MainPlayer.Height == core.GroundY {
		if core.MainPlayer.Stance.Direction == core.Right {
			if core.MainPlayer.WalkingState == core.WalkingRight {
				core.MainPlayer.WalkingState = core.LeapingRight
				core.MainPlayer.WalkingStanceTo = stances.LeapRight
			} else {
				core.MainPlayer.WalkingState = core.JumpingRight
				core.MainPlayer.WalkingStanceTo = stances.JumpRight1
			}
			core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
			core.MainPlayer.WalkingAnimationFrame = 0
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		} else {
			if core.MainPlayer.WalkingState == core.WalkingLeft {
				core.MainPlayer.WalkingState = core.LeapingLeft
				core.MainPlayer.WalkingStanceTo = stances.LeapLeft
			} else {
				core.MainPlayer.WalkingState = core.JumpingLeft
				core.MainPlayer.WalkingStanceTo = stances.JumpLeft1
			}
			core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
			core.MainPlayer.WalkingAnimationFrame = 0
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		}
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		if core.MainPlayer.MovingLeft {
			core.MainPlayer.WalkingState = core.WalkingLeft
			core.MainPlayer.WalkingStanceTo = stances.WalkLeft1
		} else {
			core.MainPlayer.WalkingState = core.Standing
			core.MainPlayer.WalkingStanceTo = stances.RestRight1
		}
		core.MainPlayer.MovingRight = false
		core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
		core.MainPlayer.WalkingAnimationFrame = 0
		core.MainPlayer.WalkingAnimationFrames = core.WalkTransitionFrames
		tmp := stances.WalkRight1
		stances.WalkRight1 = stances.WalkRight2
		stances.WalkRight2 = tmp
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) {
		if core.MainPlayer.MovingRight {
			core.MainPlayer.WalkingState = core.WalkingRight
			core.MainPlayer.WalkingStanceTo = stances.WalkRight1
		} else {
			core.MainPlayer.WalkingState = core.Standing
			core.MainPlayer.WalkingStanceTo = stances.RestLeft1
		}
		core.MainPlayer.MovingLeft = false
		core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
		core.MainPlayer.WalkingAnimationFrame = 0
		core.MainPlayer.WalkingAnimationFrames = core.WalkTransitionFrames
		tmp := stances.WalkLeft1
		stances.WalkLeft1 = stances.WalkLeft2
		stances.WalkLeft2 = tmp
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && core.MainPlayer.X < core.PlaceWidth-0.5*core.ScreenWidth/core.PixelYardRatio {
		core.MainPlayer.WalkingState = core.WalkingRight
		core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
		core.MainPlayer.WalkingStanceTo = stances.WalkRight1
		core.MainPlayer.WalkingAnimationFrame = 0
		core.MainPlayer.WalkingAnimationFrames = core.WalkTransitionFrames
		core.MainPlayer.MovingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) && core.MainPlayer.X > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.MainPlayer.WalkingState = core.WalkingLeft
		core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
		core.MainPlayer.WalkingStanceTo = stances.WalkLeft1
		core.MainPlayer.WalkingAnimationFrame = 0
		core.MainPlayer.WalkingAnimationFrames = core.WalkTransitionFrames
		core.MainPlayer.MovingLeft = true
	}
	if core.MainPlayer.WalkingAnimationFrame == core.MainPlayer.WalkingAnimationFrames {
		core.MainPlayer.WalkingAnimationFrame = 0
		core.MainPlayer.WalkingStanceFrom = core.MainPlayer.Stance
		if core.MainPlayer.WalkingStanceTo == stances.WalkLeft1 {
			core.MainPlayer.WalkingStanceTo = stances.WalkLeft2
			core.MainPlayer.WalkingAnimationFrames = core.StepFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.WalkLeft2 {
			core.MainPlayer.WalkingStanceTo = stances.WalkLeft1
			core.MainPlayer.WalkingAnimationFrames = core.StepFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.WalkRight1 {
			core.MainPlayer.WalkingStanceTo = stances.WalkRight2
			core.MainPlayer.WalkingAnimationFrames = core.StepFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.WalkRight2 {
			core.MainPlayer.WalkingStanceTo = stances.WalkRight1
			core.MainPlayer.WalkingAnimationFrames = core.StepFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.RestRight1 {
			core.MainPlayer.WalkingStanceTo = stances.RestRight2
			core.MainPlayer.WalkingAnimationFrames = core.VibeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.RestRight2 {
			core.MainPlayer.WalkingStanceTo = stances.RestRight1
			core.MainPlayer.WalkingAnimationFrames = core.VibeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.RestLeft1 {
			core.MainPlayer.WalkingStanceTo = stances.RestLeft2
			core.MainPlayer.WalkingAnimationFrames = core.VibeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.RestLeft2 {
			core.MainPlayer.WalkingStanceTo = stances.RestLeft1
			core.MainPlayer.WalkingAnimationFrames = core.VibeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpRight1 || core.MainPlayer.WalkingStanceTo == stances.LeapRight {
			core.MainPlayer.Jump_dy += 0.5
			core.MainPlayer.WalkingStanceTo = stances.JumpRight2
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpRight2 {
			core.MainPlayer.WalkingStanceTo = stances.JumpRight3
			core.MainPlayer.WalkingAnimationFrames = core.JumpTimeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpRight3 {
			if core.MainPlayer.MovingLeft {
				core.MainPlayer.WalkingStanceTo = stances.WalkLeft1
				core.MainPlayer.WalkingState = core.WalkingLeft
			} else if core.MainPlayer.MovingRight {
				core.MainPlayer.WalkingStanceTo = stances.WalkRight1
				core.MainPlayer.WalkingState = core.WalkingRight
			} else {
				core.MainPlayer.WalkingStanceTo = stances.RestRight1
				core.MainPlayer.WalkingState = core.Standing
			}
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpLeft1 || core.MainPlayer.WalkingStanceTo == stances.LeapLeft {
			core.MainPlayer.Jump_dy += 0.5
			core.MainPlayer.WalkingStanceTo = stances.JumpLeft2
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpLeft2 {
			core.MainPlayer.WalkingStanceTo = stances.JumpLeft3
			core.MainPlayer.WalkingAnimationFrames = core.JumpTimeFrames
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpLeft3 {
			if core.MainPlayer.MovingLeft {
				core.MainPlayer.WalkingStanceTo = stances.WalkLeft1
				core.MainPlayer.WalkingState = core.WalkingLeft
			} else if core.MainPlayer.MovingRight {
				core.MainPlayer.WalkingStanceTo = stances.WalkRight1
				core.MainPlayer.WalkingState = core.WalkingRight
			} else {
				core.MainPlayer.WalkingStanceTo = stances.RestLeft1
				core.MainPlayer.WalkingState = core.Standing
			}
			core.MainPlayer.WalkingAnimationFrames = core.JumpTransitionFrames
		}
	}
	core.MainPlayer.Stance = core.ShiftStance(core.MainPlayer.WalkingStanceFrom, core.MainPlayer.WalkingStanceTo, core.MainPlayer.WalkingAnimationFrame, core.MainPlayer.WalkingAnimationFrames)
	if (core.MainPlayer.WalkingState == core.WalkingRight || core.MainPlayer.WalkingState == core.LeapingRight) && core.MainPlayer.X < core.PlaceWidth-0.5*core.ScreenWidth/core.PixelYardRatio {
		core.MainPlayer.X += 0.09
	}
	if (core.MainPlayer.WalkingState == core.WalkingLeft || core.MainPlayer.WalkingState == core.LeapingLeft) && core.MainPlayer.X > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.MainPlayer.X -= 0.09
	}
	core.MainPlayer.Jump_dy *= 0.8
	core.MainPlayer.Gravity_dy *= 0.8
	core.MainPlayer.Y += core.MainPlayer.Gravity_dy
	core.MainPlayer.Y += core.MainPlayer.Jump_dy
	core.VP.W = core.ScreenWidth
	core.VP.H = core.ScreenHeight
	core.VP.X = core.MainPlayer.X*core.PixelYardRatio - (core.VP.W / 2) + (core.PlayerWidth * core.PixelYardRatio / 2)
	core.VP.Y = core.MainPlayer.Y*core.PixelYardRatio - (core.VP.H / 2) - (core.MainPlayer.Height * core.PixelYardRatio / 2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(core.BackgroundImg, &core.BackgroundDrawOptions)
	var chunk_start_y, chunk_start_x int = 0, 0
	chunk_ends_y := int(math.Floor(core.PlaceHeight))
	chunk_ends_x := int(math.Floor(core.PlaceWidth))
	if math.Floor(core.MainPlayer.Y)-math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio) > 0 {
		chunk_start_y = int(math.Floor(core.MainPlayer.Y) - math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio))
	}
	if math.Floor(core.MainPlayer.Y+1)+math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio) < math.Floor(core.PlaceHeight) {
		chunk_ends_y = int(math.Floor(core.MainPlayer.Y+1) + math.Floor(0.75*core.ScreenHeight/core.PixelYardRatio))
	}
	if math.Floor(core.MainPlayer.X)-math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio) > 0 {
		chunk_start_x = int(math.Floor(core.MainPlayer.X) - math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio))
	}
	if math.Floor(core.MainPlayer.X+1)+math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio) < math.Floor(core.PlaceWidth) {
		chunk_ends_x = int(math.Floor(core.MainPlayer.X+1) + math.Floor(0.75*core.ScreenWidth/core.PixelYardRatio))
	}
	for i := chunk_start_y; i < chunk_ends_y; i++ {
		for j := chunk_start_x; j < chunk_ends_x; j++ {
			chunklet := core.Grid[core.Vertex{X: j, Y: i}]
			for k := 0; k < len(chunklet); k++ {
				if chunklet[k] == core.Oak {
					draw_oak(screen, float64(j)*core.PixelYardRatio, float64(i)*core.PixelYardRatio+core.OakHeight*core.PixelYardRatio)
				} else if chunklet[k] == core.OakLog {
					draw_oaklog(screen, float64(j)*core.PixelYardRatio, float64(i)*core.PixelYardRatio+core.OakLogHeight*core.PixelYardRatio)
				}
			}
			if i == 0 {
				grassblockdrawoptions := ebiten.DrawImageOptions{}
				grassblockdrawoptions.GeoM.Reset()
				wi, hi := core.GrassLayerImg.Size()
				grassblockdrawoptions.GeoM.Scale(core.PixelYardRatio/float64(wi), core.PixelYardRatio/float64(hi))
				grassblockdrawoptions.GeoM.Translate(float64(j)*core.PixelYardRatio-core.VP.X, core.GetPXY(0)+core.VP.Y)
				screen.DrawImage(core.GrassLayerImg, &grassblockdrawoptions)
				for k := 1; k < 6; k++ {
					grassblockdrawoptions.GeoM.Reset()
					wi, hi = core.DirtLayerImg.Size()
					grassblockdrawoptions.GeoM.Scale(core.PixelYardRatio/float64(wi), core.PixelYardRatio/float64(hi))
					grassblockdrawoptions.GeoM.Translate(float64(j)*core.PixelYardRatio-core.VP.X, core.GetPXY(0)+core.VP.Y+float64(k)*core.PixelYardRatio)
					screen.DrawImage(core.DirtLayerImg, &grassblockdrawoptions)
				}
			}
		}
	}
	// ebitenutil.DebugPrint(screen, fmt.Sprintf("trees loaded: %d\n", oak_counter))
	drawplayer.DrawPlayer(screen, core.MainPlayer, core.MainPlayer.X*core.PixelYardRatio, core.MainPlayer.Y*core.PixelYardRatio)
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

func draw_oak(screen *ebiten.Image, x float64, y float64) {
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	oak_width_int, oak_height_int := core.OakImg.Size() // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Scale(core.OakWidth*core.PixelYardRatio/float64(oak_width_int), core.OakHeight*core.PixelYardRatio/float64(oak_height_int))
	oakDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PixelYardRatio)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
}

func draw_oaklog(screen *ebiten.Image, x float64, y float64) {
	logDrawOptions := &ebiten.DrawImageOptions{}
	logDrawOptions.GeoM.Reset()
	wi, hi := core.OakLogImg.Size()
	logDrawOptions.GeoM.Scale(core.OakLogWidth*core.PixelYardRatio/float64(wi), core.OakLogHeight*core.PixelYardRatio/float64(hi))
	logDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PixelYardRatio)+core.VP.Y)
	screen.DrawImage(core.OakLogImg, logDrawOptions)
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
		case "oaklog":
			core.Grid[core.Vertex{X: x, Y: y}] = append(core.Grid[core.Vertex{X: x, Y: y}], core.OakLog)
		}
	}

	ebiten.SetWindowSize(int(core.ScreenWidth), int(core.ScreenHeight))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
