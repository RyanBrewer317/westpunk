package core

import (
	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// this is only for drawing the player, enabling multiple players to be rendered in a multi-player environment
type Player struct {
	// the player's x,y coordinate is at their left shoulder from the viewer's perspective
	X float64
	Y float64

	// forces are separated so they can be uniquely cancelled. Walking left and right is not a force
	Jump_dy    float64
	Gravity_dy float64

	// the player is constantly transitioning from one animation to another, for smoothness and liveliness
	Stance            Stance
	WalkingStanceFrom Stance
	WalkingStanceTo   Stance

	// these are for knowing how far you are from your last stance and the one youre transitioning to
	WalkingAnimationFrame  int
	WalkingAnimationFrames int

	// these are so that continuing to hold down a key after a complex movement still moves you how you intend to move
	MovingLeft   bool
	MovingRight  bool
	WalkingState AnimationType

	Height      float64
	DrawOptions ebiten.DrawImageOptions
}

// static objects are stored on a grid so that only nearby ones are "loaded"
type Coordinate struct {
	X, Y int
}

// When the player moves, these variables are updated, then everything else in the world moves to keep the player in the center
type Viewport struct {
	X float64
	Y float64
	W float64
	H float64
}

// there's too much in a stance to not organize it this way
type Stance struct {
	Head          float64
	Torso         float64
	RightUpperArm float64
	LeftUpperArm  float64
	RightLowerArm float64
	LeftLowerArm  float64
	RightUpperLeg float64
	LeftUpperLeg  float64
	RightLowerLeg float64
	LeftLowerLeg  float64
	RightFoot     float64
	LeftFoot      float64
	Weapon        float64
	Direction     Direction
}

// enums
type AnimationType int
type Direction int
type Thing int

var VP Viewport

const (
	WALK_TRANSITION_FRAMES int = 20 // how long to transition from standing to walking and back
	STEP_FRAMES            int = 25 // how long to transition from left foot forward to right foot forward and back
	VIBE_FRAMES            int = 50 // how long to transition from inhale to exhale and back
	JUMP_TRANSITION_FRAMES int = 10 // how long to transition from crouching to up in the air
	JUMP_TIME_FRAMES       int = 10 // how long to transition from in the air to landing
	// animation type enums
	STANDING AnimationType = iota + 1
	WALKING_RIGHT
	WALKING_LEFT
	JUMPING_RIGHT
	JUMPING_LEFT
	LEAPING_RIGHT
	LEAPING_LEFT
	// direction enums
	RIGHT Direction = iota + 1
	LEFT
	// thing type enums
	OAK Thing = iota + 1
	OAK_LOG
	// player body proportion constants
	TORSO_WIDTH      float64 = 0.25
	TORSO_HEIGHT     float64 = 0.5
	HEAD_WIDTH       float64 = 0.35
	HEAD_HEIGHT      float64 = 0.3
	UPPER_ARM_HEIGHT float64 = 0.25
	UPPER_ARM_WIDTH  float64 = 0.05
	LOWER_ARM_HEIGHT float64 = 0.25
	LOWER_ARM_WIDTH  float64 = 0.05
	UPPER_LEG_WIDTH  float64 = 0.05
	UPPER_LEG_HEIGHT float64 = 0.25
	LOWER_LEG_WIDTH  float64 = 0.05
	LOWER_LEG_HEIGHT float64 = 0.25
	// other proportion constants
	PLACE_WIDTH    float64 = 256
	PLACE_HEIGHT   float64 = 128
	PLAYER_WIDTH   float64 = 0.5
	GROUND_HEIGHT  float64 = SCREEN_HEIGHT / PIXEL_YARD_RATIO
	GROUND_Y       float64 = 0
	OAK_HEIGHT     float64 = 5
	OAK_WIDTH      float64 = 2
	OAK_LOG_WIDTH  float64 = 0.3
	OAK_LOG_HEIGHT float64 = 0.3
	SCREEN_HEIGHT  float64 = 540 // in pixels
	SCREEN_WIDTH   float64 = 810 // in pixels
	// the conversion factor from units to pixels
	PIXEL_YARD_RATIO float64 = 70
)

//the player that the viewport centers around and the inputs control
var MainPlayer Player = Player{
	WalkingState:           STANDING,
	WalkingAnimationFrame:  0,
	WalkingAnimationFrames: VIBE_FRAMES,
	MovingLeft:             false,
	MovingRight:            false,
	X:                      70, // start in the middle-ish of the world
	Y:                      0,
	Jump_dy:                0,
	Gravity_dy:             0,
	Height:                 1,
}

// the grid of static things, so that only the nearby ones are "loaded"
var Grid map[Coordinate][]Thing

//global variables for the ebiten library
var (
	PlayerImg             *ebiten.Image
	GrassLayerImg         *ebiten.Image
	DirtLayerImg          *ebiten.Image
	OakImg                *ebiten.Image
	OakLogImg             *ebiten.Image
	BackgroundImg         *ebiten.Image
	BackgroundDrawOptions ebiten.DrawImageOptions
)

// convert a unit y-coordinate a pixel y-coordinate
func GetPXY(y float64) float64 {
	return SCREEN_HEIGHT - (y * PIXEL_YARD_RATIO)
}

// draw an image on the screen at the given x,y coordinate and with the given options
func DrawImage(screen *ebiten.Image, img *ebiten.Image, drawoptions ebiten.DrawImageOptions, x float64, y float64) {
	drawoptions.GeoM.Translate(x, y)
	screen.DrawImage(img, &drawoptions)
}

// returns a stances that is the given distance in transition (frame/frames) from s1 to s2
func ShiftStance(s1 Stance, s2 Stance, frame int, frames int) Stance {
	c := float64(frame) / float64(frames)
	return Stance{
		Head:          c*(s2.Head-s1.Head) + s1.Head,
		Torso:         c*(s2.Torso-s1.Torso) + s1.Torso,
		RightUpperArm: c*(s2.RightUpperArm-s1.RightUpperArm) + s1.RightUpperArm,
		LeftUpperArm:  c*(s2.LeftUpperArm-s1.LeftUpperArm) + s1.LeftUpperArm,
		RightLowerArm: c*(s2.RightLowerArm-s1.RightLowerArm) + s1.RightLowerArm,
		LeftLowerArm:  c*(s2.LeftLowerArm-s1.LeftLowerArm) + s1.LeftLowerArm,
		RightUpperLeg: c*(s2.RightUpperLeg-s1.RightUpperLeg) + s1.RightUpperLeg,
		LeftUpperLeg:  c*(s2.LeftUpperLeg-s1.LeftUpperLeg) + s1.LeftUpperLeg,
		RightLowerLeg: c*(s2.RightLowerLeg-s1.RightLowerLeg) + s1.RightLowerLeg,
		LeftLowerLeg:  c*(s2.LeftLowerLeg-s1.LeftLowerLeg) + s1.LeftLowerLeg,
		RightFoot:     c*(s2.RightFoot-s1.RightFoot) + s1.RightFoot,
		LeftFoot:      c*(s2.LeftFoot-s1.LeftFoot) + s1.LeftFoot,
		Weapon:        c*(s2.Weapon-s1.Weapon) + s1.Weapon,
		Direction:     s2.Direction,
	}
}
