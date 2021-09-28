package main

import (
	"database/sql"
	"fmt"
	_ "image/png"
	"log"

	// "math/rand"
	"strconv"
	"strings"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/mattn/go-sqlite3"
	"rbrewer.com/core"
	"rbrewer.com/player"
	"rbrewer.com/stances"
)

// go-sqlite3 docs https://github.com/mattn/go-sqlite3/blob/v1.14.8/_example/simple/simple.go
// ebiten docs https://ebiten.org/tour/hello_world.html

func init() {
	//this function is called automatically by ebiten

	//these stances are created outside of core, so this declaration has to happen here in main
	core.MainPlayer.WalkingStanceTo = stances.RestRight2
	core.MainPlayer.WalkingStanceFrom = stances.RestRight1
	stances.CreateStanceContinuations()

	//load the game assets
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

	core.ResizeImage(core.BackgroundImg, &core.BackgroundDrawOptions, core.SCREEN_WIDTH, core.SCREEN_HEIGHT)
}

// the struct ebiten uses as a base to operate around
type Game struct{}

func (g *Game) Update() error {
	// this function is called automatically every game tick (not every animation frame) by ebiten

	//update the main player's animation clock to move it one step closer to the stance it's approaching
	core.MainPlayer.WalkingAnimationFrame += 1

	player.SetPlayerHeight(&core.MainPlayer)

	player.ApplyGravity(&core.MainPlayer)

	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.MainPlayer.Y-core.MainPlayer.Height == core.GROUND_Y { // if theres a jump intent and the player is on the ground
		player.StartJump(&core.MainPlayer)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) { // intent to walk right has ended
		player.StopMovingRight(&core.MainPlayer)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) { // intent to walk left has ended
		player.StopMovingLeft(&core.MainPlayer)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && core.MainPlayer.X < core.PLACE_WIDTH-0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO { // intent to walk right and the right isnt obstructed
		core.ChangeWalkState(&core.MainPlayer, core.WALKING_RIGHT, stances.WalkRight1, core.WALK_TRANSITION_FRAMES)
		core.MainPlayer.MovingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) && core.MainPlayer.X > 0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO { // intent to walk left and the left isnt obstructed
		core.ChangeWalkState(&core.MainPlayer, core.WALKING_LEFT, stances.WalkLeft1, core.WALK_TRANSITION_FRAMES)
		core.MainPlayer.MovingLeft = true
	}

	if core.MainPlayer.WalkingAnimationFrame == core.MainPlayer.WalkingAnimationFrames { // reached the new stance
		player.ContinueStance(&core.MainPlayer)
		if core.MainPlayer.WalkingStanceTo == stances.JumpRight2 || core.MainPlayer.WalkingStanceTo == stances.JumpLeft2 {
			player.ActualJump(&core.MainPlayer)
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpRight3 || core.MainPlayer.WalkingStanceTo == stances.JumpLeft3 {
			player.EndJump(&core.MainPlayer)
		}
	}
	player.UpdateStance(&core.MainPlayer)

	if (core.MainPlayer.WalkingState == core.WALKING_RIGHT || core.MainPlayer.WalkingState == core.LEAPING_RIGHT) && player.CanMoveRight(&core.MainPlayer) {
		// if theres intent to move right and the right isn't obstructed
		player.MoveRight(&core.MainPlayer)
	}
	if (core.MainPlayer.WalkingState == core.WALKING_LEFT || core.MainPlayer.WalkingState == core.LEAPING_LEFT) && player.CanMoveLeft(&core.MainPlayer) {
		// if theres intent to move left and the left isnt obstructed
		player.MoveLeft(&core.MainPlayer)
	}

	player.ApplyAirResistance(&core.MainPlayer)
	player.NaturalMotion(&core.MainPlayer) // convert all the forces into actual motion

	// shift the viewport
	core.VP.X = core.MainPlayer.X*core.PIXEL_YARD_RATIO - (core.SCREEN_WIDTH / 2) + (core.PLAYER_WIDTH * core.PIXEL_YARD_RATIO / 2)
	core.VP.Y = core.MainPlayer.Y*core.PIXEL_YARD_RATIO - (core.SCREEN_HEIGHT / 2) - (core.MainPlayer.Height * core.PIXEL_YARD_RATIO / 2)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// this function is called automatically by ebiten every animation frame

	// draw background (todo: parallax)
	screen.DrawImage(core.BackgroundImg, &core.BackgroundDrawOptions)

	chunk := core.GetChunk(core.MainPlayer)
	// cycle through every 1x1 area in the big area around the player, to find nearby static objects and put them on screen
	for i := chunk.StartY; i < chunk.EndY; i++ {
		for j := chunk.StartX; j < chunk.EndX; j++ {
			chunklet := core.Grid[core.Coordinate{X: j, Y: i}]
			for k := 0; k < len(chunklet); k++ {
				if chunklet[k] == core.OAK {
					draw_oak(screen, float64(j)*core.PIXEL_YARD_RATIO, float64(i)*core.PIXEL_YARD_RATIO+core.OAK_HEIGHT*core.PIXEL_YARD_RATIO)
				} else if chunklet[k] == core.OAK_LOG {
					draw_oaklog(screen, float64(j)*core.PIXEL_YARD_RATIO, float64(i)*core.PIXEL_YARD_RATIO+core.OAK_LOG_HEIGHT*core.PIXEL_YARD_RATIO)
				}
			}
			// if there's ground here, draw some ground
			if i == 0 {
				grassblockdrawoptions := ebiten.DrawImageOptions{}
				grassblockdrawoptions.GeoM.Reset()
				core.ResizeImage(core.GrassLayerImg, &grassblockdrawoptions, core.PIXEL_YARD_RATIO, core.PIXEL_YARD_RATIO)
				grassblockdrawoptions.GeoM.Translate(float64(j)*core.PIXEL_YARD_RATIO-core.VP.X, core.GetPXY(0)+core.VP.Y)
				screen.DrawImage(core.GrassLayerImg, &grassblockdrawoptions)
				for k := 1; k < 6; k++ { // arbitrarily chosen numbers to put dirt all the way down the screen and out of sight
					grassblockdrawoptions.GeoM.Reset()
					core.ResizeImage(core.DirtLayerImg, &grassblockdrawoptions, core.PIXEL_YARD_RATIO, core.PIXEL_YARD_RATIO) // extremely inefficient
					grassblockdrawoptions.GeoM.Translate(float64(j)*core.PIXEL_YARD_RATIO-core.VP.X, core.GetPXY(0)+core.VP.Y+float64(k)*core.PIXEL_YARD_RATIO)
					screen.DrawImage(core.DirtLayerImg, &grassblockdrawoptions)
				}
			}
		}
	}

	player.DrawPlayer(screen, core.MainPlayer, core.MainPlayer.X*core.PIXEL_YARD_RATIO, core.MainPlayer.Y*core.PIXEL_YARD_RATIO)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// this function is automatically called by ebiten
	return int(core.SCREEN_WIDTH), int(core.SCREEN_HEIGHT)
}

// this function might be useful later:
// func dbrun(db *sql.DB, sqlstuff string) {
// 	_, err := db.Exec(sqlstuff)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlstuff)
// 		return
// 	}
// }

func dbget(db *sql.DB, sqlstuff string) *sql.Rows {
	// run a sql query on the database and return the results
	rows, err := db.Query(sqlstuff)
	if err != nil {
		log.Fatal(fmt.Sprintf("%q: %s\n", err, sqlstuff))
	}
	return rows
}

func draw_oak(screen *ebiten.Image, x float64, y float64) {
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	core.ResizeImage(core.OakImg, oakDrawOptions, core.OAK_WIDTH*core.PIXEL_YARD_RATIO, core.OAK_HEIGHT*core.PIXEL_YARD_RATIO) // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
}

func draw_oaklog(screen *ebiten.Image, x float64, y float64) {
	logDrawOptions := &ebiten.DrawImageOptions{}
	logDrawOptions.GeoM.Reset()
	// NOTE optimizeable by moving resizing somewhere else that's not called every tick
	core.ResizeImage(core.OakLogImg, logDrawOptions, core.OAK_LOG_WIDTH*core.PIXEL_YARD_RATIO, core.OAK_LOG_HEIGHT*core.PIXEL_YARD_RATIO)
	logDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
	screen.DrawImage(core.OakLogImg, logDrawOptions)
}

func main() {
	//open the database
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//fill out the grid
	core.Grid = make(map[core.Coordinate][]core.Thing)
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
			core.Grid[core.Coordinate{X: x, Y: y}] = append(core.Grid[core.Coordinate{X: x, Y: y}], core.OAK)
		case "oaklog":
			core.Grid[core.Coordinate{X: x, Y: y}] = append(core.Grid[core.Coordinate{X: x, Y: y}], core.OAK_LOG)
		}
	}

	//construct and run the game
	ebiten.SetWindowSize(int(core.SCREEN_WIDTH), int(core.SCREEN_HEIGHT))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
