package main

import (
	"database/sql"
	"fmt"
	"image/color"
	_ "image/png"
	"io/ioutil"
	"log"

	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	_ "modernc.org/sqlite"
	"ryanbrewer.page/audio"
	"ryanbrewer.page/core"
	"ryanbrewer.page/physics"
	"ryanbrewer.page/player"
	"ryanbrewer.page/settings_ui"
	"ryanbrewer.page/stances"
)

func init() {
	// this function is called automatically by ebiten

	// these stances are created outside of core, so this declaration has to happen here in package main
	// OPTIMIZABLE the current stance system is bad lol
	core.MainPlayer.WalkingStanceTo = stances.RestRight2
	core.MainPlayer.WalkingStanceFrom = stances.RestRight1
	stances.CreateStanceContinuations()

	// initialize the obstruction type table
	core.ObstructionTable = make(map[core.ThingType]core.ObstructionType)
	core.ObstructionTable[core.THING_TYPE_OAK] = core.OBSTRUCTION_TYPE_UNOBSTRUCTIVE
	core.ObstructionTable[core.THING_TYPE_OAK_LOG] = core.OBSTRUCTION_TYPE_UNOBSTRUCTIVE
	core.ObstructionTable[core.THING_TYPE_RAMP_RIGHT_45] = core.OBSTRUCTION_TYPE_RIGHT_SLANT_45

	// load the game assets
	// this feels optimizable
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
	core.RightRamp45Img, _, err = ebitenutil.NewImageFromFile("slant_right_45.png")
	if err != nil {
		log.Fatal(err)
	}
	core.BackgroundDrawOptions = ebiten.DrawImageOptions{}
	core.SettingsBackgroundImg = ebiten.NewImage(int(core.SCREEN_WIDTH), int(core.SCREEN_HEIGHT))
	core.SettingsBackgroundImg.Fill(color.White)

	// in general this loads GUI images, and may at some point do whatever other init stuff needs doing
	settings_ui.PrepareSettingsUIImages()

	// load game font
	fontBytes, err := ioutil.ReadFile("dpcomic.ttf") // TODO: change font
	if err != nil {
		log.Fatal(err)
	}
	core.FONT, err = freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	// create the background image TODO: parallax
	core.ResizeImage(core.BackgroundImg, &core.BackgroundDrawOptions, core.SCREEN_WIDTH, core.SCREEN_HEIGHT)
}

// the struct ebiten uses as a base to operate around
// I opted to use global variables instead of Game properties for a variety of reasons
// the main reason is heck OOP amirite B)
type Game struct{}

func (g *Game) Update() error {
	// this function is called automatically every game tick (not every animation frame) by ebiten

	// toggle game pause on esc keyup
	if inpututil.IsKeyJustReleased(ebiten.KeyEscape) {
		core.SETTINGS_GAME_PAUSED = !core.SETTINGS_GAME_PAUSED
	}

	if core.SETTINGS_GAME_PAUSED {
		// TODO: move this code into the settings_ui module for repurposability
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			mousex, mousey := ebiten.CursorPosition()
			mousePhysics := core.PhysicsComponent{
				Position: core.Vector2{
					X: float64(mousex),
					Y: float64(mousey),
				},
				Height: 0, // the mouse point is a 2D collider rect with 0 area
				Width:  0,
			}
			if physics.CollisionDetected(mousePhysics, settings_ui.MuteSFX.Physics) {
				core.SETTINGS_SFX_MUTED = !core.SETTINGS_SFX_MUTED
			}
		}
		return nil
	}
	// update the main player's animation clock to move it one step closer to the stance it's approaching
	// OPTIMIZABLE I don't like the stance animation system's current implementation
	core.MainPlayer.WalkingAnimationFrame += 1

	// apply physics to stuff
	physics.Move(&core.MainPlayer.Physics)

	// foot_pos := player.LeftFootPos(core.MainPlayer)
	// new_foot_y := foot_pos.Y
	// // _, new_foot_y := physics.Grounded(core.PhysicsComponent{
	// // 	Position: foot_pos,
	// // })

	// player.PositionLeftFoot(&core.MainPlayer, foot_pos.X, new_foot_y)

	// process inputs
	if (inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyW)) && core.MainPlayer.Physics.Grounded { // if theres a jump intent and the player is on the ground
		player.StartJump(&core.MainPlayer)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) { // intent to walk right has ended
		player.StopMovingRight(&core.MainPlayer)
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) { // intent to walk left has ended
		player.StopMovingLeft(&core.MainPlayer)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && physics.CanMoveRight(&core.MainPlayer.Physics) { // intent to walk right and the right isnt obstructed
		core.ChangeWalkState(&core.MainPlayer, core.ANIMATION_TYPE_WALKING_RIGHT, stances.WalkRight1, core.WALK_TRANSITION_FRAMES)
		core.MainPlayer.MovingRight = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) && physics.CanMoveLeft(&core.MainPlayer.Physics) { // intent to walk left and the left isnt obstructed
		core.ChangeWalkState(&core.MainPlayer, core.ANIMAtION_TYPE_WALKING_LEFT, stances.WalkLeft1, core.WALK_TRANSITION_FRAMES)
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

	if (core.MainPlayer.WalkingState == core.ANIMATION_TYPE_WALKING_RIGHT || core.MainPlayer.WalkingState == core.ANIMATION_TYPE_LEAPING_RIGHT) && physics.CanMoveRight(&core.MainPlayer.Physics) {
		// if theres intent to move right and the right isn't obstructed
		physics.MoveRight(&core.MainPlayer.Physics)
	}
	if (core.MainPlayer.WalkingState == core.ANIMAtION_TYPE_WALKING_LEFT || core.MainPlayer.WalkingState == core.ANIMATION_TYPE_LEAPING_LEFT) && physics.CanMoveLeft(&core.MainPlayer.Physics) {
		// if theres intent to move left and the left isnt obstructed
		physics.MoveLeft(&core.MainPlayer.Physics)
	}

	// calculate the player's height based on the angles of the body parts
	player.SetPlayerHeight(&core.MainPlayer)

	// update volume, panning, etc of sfx based on distance from player TODO: this should probably incorporate Position.Y as well
	audio.UpdateSFXBasedOnPositions(core.MainPlayer.Physics.Position.X)

	// shift the viewport
	core.VP.X = core.MainPlayer.Physics.Position.X*core.PIXEL_YARD_RATIO - (core.SCREEN_WIDTH / 2) + (core.PLAYER_WIDTH * core.PIXEL_YARD_RATIO / 2)
	core.VP.Y = (core.MainPlayer.Physics.Position.Y+core.MainPlayer.Physics.Height)*core.PIXEL_YARD_RATIO - (core.SCREEN_HEIGHT / 2) - (core.MainPlayer.Physics.Height * core.PIXEL_YARD_RATIO / 2)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// this function is called automatically by ebiten every animation frame

	if core.SETTINGS_GAME_PAUSED {
		// if the game is paused, show the settings screen and don't redraw the world
		core.DrawImage(screen, core.SettingsBackgroundImg, ebiten.DrawImageOptions{}, 0, 0)
		settings_ui.MuteSFX.Physics.Position.X = core.SCREEN_WIDTH * 0.75 // this is wip af
		settings_ui.MuteSFX.Physics.Position.Y = core.SCREEN_HEIGHT * 0.5
		text.Draw(screen, "Mute SFX", truetype.NewFace(core.FONT, &truetype.Options{}), int(core.SCREEN_WIDTH*0.5), int(core.SCREEN_HEIGHT*0.5), color.Black) // also wip af
		settings_ui.MuteSFX.Draw(screen)
		return
	}

	// draw background (TODO: parallax)
	screen.DrawImage(core.BackgroundImg, &core.BackgroundDrawOptions)

	chunk := core.GetChunk(core.MainPlayer.Physics)
	// cycle through every 1x1 area in the big area around the player, to find nearby static objects and put them on screen
	// still need to plan out what to do if an object is bigger than a chunk (ie trees). Maybe we disallow that? or maybe it is recognized as in all chunks that it's in?
	// in general that's only a problem for truly enormous things because this code goes through chunks that are slightly beyond the vision of the player
	for i := chunk.StartY; i < chunk.EndY; i++ {
		for j := chunk.StartX; j < chunk.EndX; j++ {
			chunklet := core.Grid[core.Coordinate{X: j, Y: i}]
			for k := 0; k < len(chunklet); k++ {
				x := chunklet[k].Physics.Position.X * core.PIXEL_YARD_RATIO
				y := chunklet[k].Physics.Position.Y*core.PIXEL_YARD_RATIO + chunklet[k].Physics.Height*core.PIXEL_YARD_RATIO
				switch chunklet[k].Type {
				case core.THING_TYPE_OAK:
					draw_oak(screen, x, y)
					break
				case core.THING_TYPE_OAK_LOG:
					draw_oaklog(screen, x, y)
					break
				case core.THING_TYPE_RAMP_RIGHT_45:
					draw_right_ramp_45(screen, x, y)
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

	player.DrawPlayer(screen, core.MainPlayer, core.MainPlayer.Physics.Position.X*core.PIXEL_YARD_RATIO, (core.MainPlayer.Physics.Position.Y+core.MainPlayer.Physics.Height)*core.PIXEL_YARD_RATIO) // player is drawn from the left shoulder
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
	// resize and translate the oak image, then put it on screen
	oakDrawOptions := &ebiten.DrawImageOptions{}
	oakDrawOptions.GeoM.Reset()
	core.ResizeImage(core.OakImg, oakDrawOptions, core.OAK_WIDTH*core.PIXEL_YARD_RATIO, core.OAK_HEIGHT*core.PIXEL_YARD_RATIO) // optimizable by moving scaling somewhere else that's not called every tick
	oakDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
	screen.DrawImage(core.OakImg, oakDrawOptions)
}

func draw_oaklog(screen *ebiten.Image, x float64, y float64) {
	// resize and translate the oak log image, then put it on screen
	logDrawOptions := &ebiten.DrawImageOptions{}
	logDrawOptions.GeoM.Reset()
	// NOTE optimizeable by moving resizing somewhere else that's not called every tick
	// Actually that might not be true, depending on how ebiten applies transformations. Worth testing tho
	core.ResizeImage(core.OakLogImg, logDrawOptions, core.OAK_LOG_WIDTH*core.PIXEL_YARD_RATIO, core.OAK_LOG_HEIGHT*core.PIXEL_YARD_RATIO)
	logDrawOptions.GeoM.Translate(x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
	screen.DrawImage(core.OakLogImg, logDrawOptions)
}

func draw_right_ramp_45(screen *ebiten.Image, x float64, y float64) {
	// resize and translate the right ramp 45 image, then put it on screen
	draw_options := &ebiten.DrawImageOptions{}
	draw_options.GeoM.Reset()
	core.ResizeImage(core.RightRamp45Img, draw_options, core.RAMP_RIGHT_45_WIDTH*core.PIXEL_YARD_RATIO, core.RAMP_RIGHT_45_HEIGHT*core.PIXEL_YARD_RATIO)
	core.DrawImage(screen, core.RightRamp45Img, *draw_options, x-core.VP.X, core.GetPXY(y/core.PIXEL_YARD_RATIO)+core.VP.Y)
}

func main() {
	defer audio.Close()
	audio.Init()

	// open the database
	db, err := sql.Open("sqlite", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// load the game state from sqlite
	core.Grid = make(map[core.Coordinate][]core.ThingInstance)
	rows := dbget(db, "select * from things WHERE placeID = \"place0\"") // TODO in the future we need to get the main player's place and use that instead
	defer rows.Close()
	for rows.Next() {
		var thingID, placeID, location, offset, thingtype string
		err = rows.Scan(&thingID, &placeID, &location, &offset, &thingtype)
		if err != nil {
			log.Fatal(err)
		}
		// parse database entries
		locsplit := strings.Split(location, " ")
		x, _ := strconv.Atoi(locsplit[0])
		y, _ := strconv.Atoi(locsplit[1])
		offsetsplit := strings.Split(offset, " ")
		offset_x, _ := strconv.ParseFloat(offsetsplit[0], 64)
		offset_y, _ := strconv.ParseFloat(offsetsplit[1], 64)
		// populate world
		switch t := thingtype; t {
		case "oak":
			core.Grid[core.Coordinate{X: x, Y: y}] = append(core.Grid[core.Coordinate{X: x, Y: y}], core.ThingInstance{
				Type: core.THING_TYPE_OAK,
				Physics: core.PhysicsComponent{
					Position: core.Vector2{
						X: float64(x) + offset_x,
						Y: float64(y) + offset_y,
					},
					Height: core.OAK_HEIGHT,
					Width:  core.OAK_WIDTH,
					Forces: make(map[core.ForceType]*core.Vector2),
				},
			})
		case "oaklog":
			core.Grid[core.Coordinate{X: x, Y: y}] = append(core.Grid[core.Coordinate{X: x, Y: y}], core.ThingInstance{
				Type: core.THING_TYPE_OAK_LOG,
				Physics: core.PhysicsComponent{
					Position: core.Vector2{
						X: float64(x) + offset_x,
						Y: float64(y) + offset_y,
					},
					Height: core.OAK_LOG_HEIGHT,
					Width:  core.OAK_LOG_WIDTH,
					Forces: make(map[core.ForceType]*core.Vector2),
				},
			})
		case "rightramp45":
			core.Grid[core.Coordinate{X: x, Y: y}] = append(core.Grid[core.Coordinate{X: x, Y: y}], core.ThingInstance{
				Type: core.THING_TYPE_RAMP_RIGHT_45,
				Physics: core.PhysicsComponent{
					Position: core.Vector2{
						X: float64(x) + offset_x,
						Y: float64(y) + offset_y,
					},
					Height: core.RAMP_RIGHT_45_HEIGHT,
					Width:  core.RAMP_RIGHT_45_WIDTH,
					Forces: make(map[core.ForceType]*core.Vector2),
				},
			})
		}
	}

	//construct and run the game
	ebiten.SetWindowSize(int(core.SCREEN_WIDTH), int(core.SCREEN_HEIGHT))
	ebiten.SetWindowTitle("Westpunk")
	if err = ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
