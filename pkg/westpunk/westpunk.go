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
	core.WalkingStanceTo = stances.RestRight2
	core.WalkingStanceFrom = stances.RestRight1
	var err error
	core.PlayerImg, _, err = ebitenutil.NewImageFromFile("spritesheet.png")
	if err != nil {
		log.Fatal(err)
	}
	core.PlayerDrawOptions = ebiten.DrawImageOptions{}
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
	core.WalkingAnimationFrame += 1

	rightlegheight := core.UpperLegHeight*math.Cos(stances.PlayerStance.Torso+stances.PlayerStance.RightUpperLeg) + core.LowerLegHeight*math.Cos(stances.PlayerStance.Torso+stances.PlayerStance.RightUpperLeg+stances.PlayerStance.RightLowerLeg)
	leftlegheight := core.UpperLegHeight*math.Cos(stances.PlayerStance.Torso+stances.PlayerStance.LeftUpperLeg) + core.LowerLegHeight*math.Cos(stances.PlayerStance.Torso+stances.PlayerStance.LeftUpperLeg+stances.PlayerStance.LeftLowerLeg)
	core.PlayerHeight = core.TorsoHeight*math.Cos(stances.PlayerStance.Torso) + math.Max(rightlegheight, leftlegheight)

	if core.PlayerY-core.PlayerHeight > core.GroundY {
		core.PlayerYVelocity -= 0.03
	} else if core.PlayerYVelocity < 0 {
		// core.PlayerYVelocity = 0
		core.PlayerY = core.PlayerHeight
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.PlayerY-core.PlayerHeight == core.GroundY {
		if stances.PlayerStance.Direction == core.Right {
			if core.WalkingState == core.WalkingRight {
				core.WalkingState = core.LeapingRight
				core.WalkingStanceTo = stances.LeapRight
			} else {
				core.WalkingState = core.JumpingRight
				core.WalkingStanceTo = stances.JumpRight1
			}
			core.WalkingStanceFrom = stances.PlayerStance
			core.WalkingAnimationFrame = 0
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		} else {
			if core.WalkingState == core.WalkingLeft {
				core.WalkingState = core.LeapingLeft
				core.WalkingStanceTo = stances.LeapLeft
			} else {
				core.WalkingState = core.JumpingLeft
				core.WalkingStanceTo = stances.JumpLeft1
			}
			core.WalkingStanceFrom = stances.PlayerStance
			core.WalkingAnimationFrame = 0
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		}
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		if core.MovingLeft {
			core.WalkingState = core.WalkingLeft
			core.WalkingStanceTo = stances.WalkLeft1
		} else {
			core.WalkingState = core.Standing
			core.WalkingStanceTo = stances.RestRight1
		}
		core.MovingRight = false
		core.WalkingStanceFrom = stances.PlayerStance
		core.WalkingAnimationFrame = 0
		core.WalkingAnimationFrames = core.WalkTransitionFrames
		tmp := stances.WalkRight1
		stances.WalkRight1 = stances.WalkRight2
		stances.WalkRight2 = tmp
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) {
		if core.MovingRight {
			core.WalkingState = core.WalkingRight
			core.WalkingStanceTo = stances.WalkRight1
		} else {
			core.WalkingState = core.Standing
			core.WalkingStanceTo = stances.RestLeft1
		}
		core.MovingLeft = false
		core.WalkingStanceFrom = stances.PlayerStance
		core.WalkingAnimationFrame = 0
		core.WalkingAnimationFrames = core.WalkTransitionFrames
		tmp := stances.WalkLeft1
		stances.WalkLeft1 = stances.WalkLeft2
		stances.WalkLeft2 = tmp
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && core.PlayerX < core.PlaceWidth-0.5*core.ScreenWidth/core.PixelYardRatio {
		core.WalkingState = core.WalkingRight
		core.WalkingStanceFrom = stances.PlayerStance
		core.WalkingStanceTo = stances.WalkRight1
		core.WalkingAnimationFrame = 0
		core.WalkingAnimationFrames = core.WalkTransitionFrames
		core.MovingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) && core.PlayerX > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.WalkingState = core.WalkingLeft
		core.WalkingStanceFrom = stances.PlayerStance
		core.WalkingStanceTo = stances.WalkLeft1
		core.WalkingAnimationFrame = 0
		core.WalkingAnimationFrames = core.WalkTransitionFrames
		core.MovingLeft = true
	}
	if core.WalkingAnimationFrame == core.WalkingAnimationFrames {
		core.WalkingAnimationFrame = 0
		core.WalkingStanceFrom = stances.PlayerStance
		if core.WalkingStanceTo == stances.WalkLeft1 {
			core.WalkingStanceTo = stances.WalkLeft2
			core.WalkingAnimationFrames = core.StepFrames
		} else if core.WalkingStanceTo == stances.WalkLeft2 {
			core.WalkingStanceTo = stances.WalkLeft1
			core.WalkingAnimationFrames = core.StepFrames
		} else if core.WalkingStanceTo == stances.WalkRight1 {
			core.WalkingStanceTo = stances.WalkRight2
			core.WalkingAnimationFrames = core.StepFrames
		} else if core.WalkingStanceTo == stances.WalkRight2 {
			core.WalkingStanceTo = stances.WalkRight1
			core.WalkingAnimationFrames = core.StepFrames
		} else if core.WalkingStanceTo == stances.RestRight1 {
			core.WalkingStanceTo = stances.RestRight2
			core.WalkingAnimationFrames = core.VibeFrames
		} else if core.WalkingStanceTo == stances.RestRight2 {
			core.WalkingStanceTo = stances.RestRight1
			core.WalkingAnimationFrames = core.VibeFrames
		} else if core.WalkingStanceTo == stances.RestLeft1 {
			core.WalkingStanceTo = stances.RestLeft2
			core.WalkingAnimationFrames = core.VibeFrames
		} else if core.WalkingStanceTo == stances.RestLeft2 {
			core.WalkingStanceTo = stances.RestLeft1
			core.WalkingAnimationFrames = core.VibeFrames
		} else if core.WalkingStanceTo == stances.JumpRight1 || core.WalkingStanceTo == stances.LeapRight {
			core.PlayerYVelocity += 0.5
			core.WalkingStanceTo = stances.JumpRight2
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.WalkingStanceTo == stances.JumpRight2 {
			core.WalkingStanceTo = stances.JumpRight3
			core.WalkingAnimationFrames = core.JumpTimeFrames
		} else if core.WalkingStanceTo == stances.JumpRight3 {
			if core.MovingLeft {
				core.WalkingStanceTo = stances.WalkLeft1
				core.WalkingState = core.WalkingLeft
			} else if core.MovingRight {
				core.WalkingStanceTo = stances.WalkRight1
				core.WalkingState = core.WalkingRight
			} else {
				core.WalkingStanceTo = stances.RestRight1
				core.WalkingState = core.Standing
			}
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.WalkingStanceTo == stances.JumpLeft1 || core.WalkingStanceTo == stances.LeapLeft {
			core.PlayerYVelocity += 0.5
			core.WalkingStanceTo = stances.JumpLeft2
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		} else if core.WalkingStanceTo == stances.JumpLeft2 {
			core.WalkingStanceTo = stances.JumpLeft3
			core.WalkingAnimationFrames = core.JumpTimeFrames
		} else if core.WalkingStanceTo == stances.JumpLeft3 {
			if core.MovingLeft {
				core.WalkingStanceTo = stances.WalkLeft1
				core.WalkingState = core.WalkingLeft
			} else if core.MovingRight {
				core.WalkingStanceTo = stances.WalkRight1
				core.WalkingState = core.WalkingRight
			} else {
				core.WalkingStanceTo = stances.RestLeft1
				core.WalkingState = core.Standing
			}
			core.WalkingAnimationFrames = core.JumpTransitionFrames
		}
	}
	stances.PlayerStance = core.ShiftStance(core.WalkingStanceFrom, core.WalkingStanceTo, core.WalkingAnimationFrame, core.WalkingAnimationFrames)
	if (core.WalkingState == core.WalkingRight || core.WalkingState == core.LeapingRight) && core.PlayerX < core.PlaceWidth-0.5*core.ScreenWidth/core.PixelYardRatio {
		core.PlayerX += 0.09
	}
	if (core.WalkingState == core.WalkingLeft || core.WalkingState == core.LeapingLeft) && core.PlayerX > 0.5*core.ScreenWidth/core.PixelYardRatio {
		core.PlayerX -= 0.09
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
	screen.DrawImage(core.BackgroundImg, &core.BackgroundDrawOptions)
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
