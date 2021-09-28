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

	//resize the background
	wi, hi := core.BackgroundImg.Size()
	core.BackgroundDrawOptions.GeoM.Scale(core.SCREEN_WIDTH/float64(wi), core.SCREEN_HEIGHT/float64(hi))
}

// the struct ebiten uses as a base to operate around
type Game struct{}

func (g *Game) Update() error {
	// this function is called automatically every game tick (not every animation frame) by ebiten

	//update the main player's animation clock to move it one step closer to the stance it's approaching
	core.MainPlayer.WalkingAnimationFrame += 1

	//calculate the players height (todo:  add math for arms in case the player is doing a handstand or crawling or something?)
	rightlegheight := core.UPPER_LEG_HEIGHT*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.RightUpperLeg) + core.LOWER_LEG_HEIGHT*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.RightUpperLeg+core.MainPlayer.Stance.RightLowerLeg)
	leftlegheight := core.UPPER_LEG_HEIGHT*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.LeftUpperLeg) + core.LOWER_LEG_HEIGHT*math.Cos(core.MainPlayer.Stance.Torso+core.MainPlayer.Stance.LeftUpperLeg+core.MainPlayer.Stance.LeftLowerLeg)
	core.MainPlayer.Height = core.TORSO_HEIGHT*math.Cos(core.MainPlayer.Stance.Torso) + math.Max(rightlegheight, leftlegheight)

	// if the player is falling, accelerate that fall. Else, use the height we just calculated to calculate the player Y (which is the left shoulder from the viewers perspective)
	if core.MainPlayer.Y-core.MainPlayer.Height > core.GROUND_Y {
		core.MainPlayer.Gravity_dy -= 0.03
	} else {
		core.MainPlayer.Y = core.MainPlayer.Height
	}

	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.MainPlayer.Y-core.MainPlayer.Height == core.GROUND_Y { // if theres a jump intent and the player is on the ground
		// if the player is walking to the right, they're now leaping to the right
		if core.MainPlayer.Stance.Direction == core.RIGHT {
			if core.MainPlayer.WalkingState == core.WALKING_RIGHT {
				core.ChangeWalkState(&core.MainPlayer, core.LEAPING_RIGHT, stances.LeapRight, core.JUMP_TRANSITION_FRAMES)
			} else { // if they arent walking, the jump is straight up and down, facing right
				core.ChangeWalkState(&core.MainPlayer, core.JUMPING_RIGHT, stances.JumpRight1, core.JUMP_TRANSITION_FRAMES)
			}
		} else { // if the player is walking to the left, theyre now leaping to the left
			if core.MainPlayer.WalkingState == core.WALKING_LEFT {
				core.ChangeWalkState(&core.MainPlayer, core.LEAPING_LEFT, stances.LeapLeft, core.JUMP_TRANSITION_FRAMES)
			} else { // if they arent walking, the jump is straight up and down, facing left
				core.ChangeWalkState(&core.MainPlayer, core.JUMPING_LEFT, stances.JumpLeft1, core.JUMP_TRANSITION_FRAMES)
			}
		}
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) { // intent to walk right has ended
		if core.MainPlayer.MovingLeft { // if they're still holding down the key showing intent to walk left
			core.ChangeWalkState(&core.MainPlayer, core.WALKING_LEFT, stances.WalkLeft1, core.WALK_TRANSITION_FRAMES)
		} else { // if they have no intent of walking in either direciton
			core.ChangeWalkState(&core.MainPlayer, core.STANDING, stances.RestRight1, core.WALK_TRANSITION_FRAMES)
		}
		core.MainPlayer.MovingRight = false
		// swap what walk1 and walk2 are referring to, so that spamming the walk key still makes the legs try to cross each time
		tmp := stances.WalkRight1
		stances.WalkRight1 = stances.WalkRight2
		stances.WalkRight2 = tmp
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) { // intent to walk left has ended
		if core.MainPlayer.MovingRight { // if they're still holding down the key showing intent to walk right
			core.ChangeWalkState(&core.MainPlayer, core.WALKING_RIGHT, stances.WalkRight1, core.WALK_TRANSITION_FRAMES)
		} else { // if they have no intent of walking in either direction
			core.ChangeWalkState(&core.MainPlayer, core.STANDING, stances.RestLeft1, core.WALK_TRANSITION_FRAMES)
		}
		core.MainPlayer.MovingLeft = false
		// swap what walk1 and walk2 are referring to, so that spamming the walk key still makes the legs try to cross each time
		tmp := stances.WalkLeft1
		stances.WalkLeft1 = stances.WalkLeft2
		stances.WalkLeft2 = tmp
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && core.MainPlayer.X < core.PLACE_WIDTH-0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO { // intent to walk right and the right isnt obstructed
		core.ChangeWalkState(&core.MainPlayer, core.WALKING_RIGHT, stances.WalkRight1, core.WALK_TRANSITION_FRAMES)
		core.MainPlayer.MovingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) && core.MainPlayer.X > 0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO { // intent to walk left and the left isnt obstructed
		// reset the animation clock to transition into the new stance, starting from however the player is positioned now
		core.ChangeWalkState(&core.MainPlayer, core.WALKING_LEFT, stances.WalkLeft1, core.WALK_TRANSITION_FRAMES)
		core.MainPlayer.MovingLeft = true
	}

	if core.MainPlayer.WalkingAnimationFrame == core.MainPlayer.WalkingAnimationFrames { // reached the new stance
		new_stance, frames := core.GetContinuation(core.MainPlayer.WalkingStanceTo)
		core.ChangeWalkState(&core.MainPlayer, core.MainPlayer.WalkingState, new_stance, frames)
		if core.MainPlayer.WalkingStanceTo == stances.JumpRight2 || core.MainPlayer.WalkingStanceTo == stances.JumpLeft2 {
			core.MainPlayer.Jump_dy += 0.5
		} else if core.MainPlayer.WalkingStanceTo == stances.JumpRight3 || core.MainPlayer.WalkingStanceTo == stances.JumpLeft3 {
			// if you were transitioning to jump3, transition to either walking or standing based on if the movement keys are being held down
			if core.MainPlayer.MovingLeft {
				core.ChangeWalkState(&core.MainPlayer, core.WALKING_LEFT, stances.WalkLeft1, core.JUMP_TRANSITION_FRAMES)
			} else if core.MainPlayer.MovingRight {
				core.ChangeWalkState(&core.MainPlayer, core.WALKING_RIGHT, stances.WalkRight1, core.JUMP_TRANSITION_FRAMES)
			} else if core.MainPlayer.WalkingStanceTo.Direction == core.RIGHT {
				core.ChangeWalkState(&core.MainPlayer, core.STANDING, stances.RestRight1, core.JUMP_TRANSITION_FRAMES)
			} else if core.MainPlayer.WalkingStanceTo.Direction == core.LEFT {
				core.ChangeWalkState(&core.MainPlayer, core.STANDING, stances.RestLeft1, core.JUMP_TRANSITION_FRAMES)
			}
		}
	}

	// shift the players stance one step towards what it's transitioning towards
	core.MainPlayer.Stance = core.ShiftStance(core.MainPlayer.WalkingStanceFrom, core.MainPlayer.WalkingStanceTo, core.MainPlayer.WalkingAnimationFrame, core.MainPlayer.WalkingAnimationFrames)
	if (core.MainPlayer.WalkingState == core.WALKING_RIGHT || core.MainPlayer.WalkingState == core.LEAPING_RIGHT) && core.MainPlayer.X < core.PLACE_WIDTH-0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO {
		// if theres intent to move right and the right isn't obstructed
		core.MainPlayer.X += 0.09
	}
	if (core.MainPlayer.WalkingState == core.WALKING_LEFT || core.MainPlayer.WalkingState == core.LEAPING_LEFT) && core.MainPlayer.X > 0.5*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO {
		// if theres intent to move left and the left isnt obstructed
		core.MainPlayer.X -= 0.09
	}

	// "air resistance" or something
	core.MainPlayer.Jump_dy *= 0.8
	core.MainPlayer.Gravity_dy *= 0.8
	// apply forces
	core.MainPlayer.Y += core.MainPlayer.Gravity_dy
	core.MainPlayer.Y += core.MainPlayer.Jump_dy
	// shift the viewport
	core.VP.W = core.SCREEN_WIDTH
	core.VP.H = core.SCREEN_HEIGHT
	core.VP.X = core.MainPlayer.X*core.PIXEL_YARD_RATIO - (core.VP.W / 2) + (core.PLAYER_WIDTH * core.PIXEL_YARD_RATIO / 2)
	core.VP.Y = core.MainPlayer.Y*core.PIXEL_YARD_RATIO - (core.VP.H / 2) - (core.MainPlayer.Height * core.PIXEL_YARD_RATIO / 2)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// this function is called automatically by ebiten every animation frame

	// draw background (todo: parallax)
	screen.DrawImage(core.BackgroundImg, &core.BackgroundDrawOptions)

	// initialize the chunk as filling the whole place
	var chunk_start_y, chunk_start_x int = 0, 0
	chunk_ends_y := int(math.Floor(core.PLACE_HEIGHT))
	chunk_ends_x := int(math.Floor(core.PLACE_WIDTH))
	// if the player isnt too close to the edges, shift each of the sides towards the player to construct a box around the player that's just out of view of the human player
	if math.Floor(core.MainPlayer.Y)-math.Floor(0.75*core.SCREEN_HEIGHT/core.PIXEL_YARD_RATIO) > 0 {
		chunk_start_y = int(math.Floor(core.MainPlayer.Y) - math.Floor(0.75*core.SCREEN_HEIGHT/core.PIXEL_YARD_RATIO))
	}
	if math.Floor(core.MainPlayer.Y+1)+math.Floor(0.75*core.SCREEN_HEIGHT/core.PIXEL_YARD_RATIO) < math.Floor(core.PLACE_HEIGHT) {
		chunk_ends_y = int(math.Floor(core.MainPlayer.Y+1) + math.Floor(0.75*core.SCREEN_HEIGHT/core.PIXEL_YARD_RATIO))
	}
	if math.Floor(core.MainPlayer.X)-math.Floor(0.75*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO) > 0 {
		chunk_start_x = int(math.Floor(core.MainPlayer.X) - math.Floor(0.75*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO))
	}
	if math.Floor(core.MainPlayer.X+1)+math.Floor(0.75*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO) < math.Floor(core.PLACE_WIDTH) {
		chunk_ends_x = int(math.Floor(core.MainPlayer.X+1) + math.Floor(0.75*core.SCREEN_WIDTH/core.PIXEL_YARD_RATIO))
	}
	// cycle through every 1x1 area in the big area around the player, to find nearby static objects and put them on screen
	for i := chunk_start_y; i < chunk_ends_y; i++ {
		for j := chunk_start_x; j < chunk_ends_x; j++ {
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
				wi, hi := core.GrassLayerImg.Size()
				grassblockdrawoptions.GeoM.Scale(core.PIXEL_YARD_RATIO/float64(wi), core.PIXEL_YARD_RATIO/float64(hi))
				grassblockdrawoptions.GeoM.Translate(float64(j)*core.PIXEL_YARD_RATIO-core.VP.X, core.GetPXY(0)+core.VP.Y)
				screen.DrawImage(core.GrassLayerImg, &grassblockdrawoptions)
				for k := 1; k < 6; k++ { // arbitrarily chosen numbers to put dirt all the way down the screen and out of sight
					grassblockdrawoptions.GeoM.Reset()
					wi, hi = core.DirtLayerImg.Size()
					grassblockdrawoptions.GeoM.Scale(core.PIXEL_YARD_RATIO/float64(wi), core.PIXEL_YARD_RATIO/float64(hi))
					grassblockdrawoptions.GeoM.Translate(float64(j)*core.PIXEL_YARD_RATIO-core.VP.X, core.GetPXY(0)+core.VP.Y+float64(k)*core.PIXEL_YARD_RATIO)
					screen.DrawImage(core.DirtLayerImg, &grassblockdrawoptions)
				}
			}
		}
	}

	drawplayer.DrawPlayer(screen, core.MainPlayer, core.MainPlayer.X*core.PIXEL_YARD_RATIO, core.MainPlayer.Y*core.PIXEL_YARD_RATIO)
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
	oak_width_int, oak_height_int := core.OakImg.Size() // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Scale(core.OAK_WIDTH*core.PIXEL_YARD_RATIO/float64(oak_width_int), core.OAK_HEIGHT*core.PIXEL_YARD_RATIO/float64(oak_height_int))
	oakDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
}

func draw_oaklog(screen *ebiten.Image, x float64, y float64) {
	logDrawOptions := &ebiten.DrawImageOptions{}
	logDrawOptions.GeoM.Reset()
	wi, hi := core.OakLogImg.Size() // optimizeable by moving scaling somewhere else that's not called every tick
	logDrawOptions.GeoM.Scale(core.OAK_LOG_WIDTH*core.PIXEL_YARD_RATIO/float64(wi), core.OAK_LOG_HEIGHT*core.PIXEL_YARD_RATIO/float64(hi))
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
