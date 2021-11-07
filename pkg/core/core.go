package core

import (
	"math"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// this is only for drawing the player, enabling multiple players to be rendered in a multi-player environment
type Player struct {
	// the player's x,y coordinate is at their left shoulder from the viewer's perspective
	Physics PhysicsComponent

	// the player is constantly transitioning from one animation to another, for smoothness and liveliness
	Stance, WalkingStanceFrom, WalkingStanceTo Stance

	// these are for knowing how far you are from your last stance and the one youre transitioning to
	WalkingAnimationFrame, WalkingAnimationFrames int

	// these are so that continuing to hold down a key after a complex movement still moves you how you intend to move
	MovingLeft, MovingRight bool
	WalkingState            AnimationType

	DrawOptions ebiten.DrawImageOptions
}

// static objects are stored on a grid so that only nearby ones are "loaded"
type Coordinate struct {
	X, Y int
}

type Chunk struct {
	StartX, StartY, EndX, EndY int
}

// translation of the world to keep the main player centered
type Viewport Vector2

// there's too much in a stance to not organize it this way
type Stance struct {
	Head,
	Torso,
	RightUpperArm,
	LeftUpperArm,
	RightLowerArm,
	LeftLowerArm,
	RightUpperLeg,
	LeftUpperLeg,
	RightLowerLeg,
	LeftLowerLeg,
	RightFoot,
	LeftFoot,
	Weapon float64
	Direction Direction
}

// how the stance after a given stance is stored. This could probably be improved
type StanceContinuation struct {
	Start        Stance
	Continuation Stance
	Frames       int
}

type Vector2 struct {
	X, Y float64
}

func (v *Vector2) Scale(coeff float64) {
	v.X *= coeff
	v.Y *= coeff
}

func (v *Vector2) Add(other Vector2) {
	v.X += other.X
	v.Y += other.Y
}

// the block of information that the physics engine operates on and that the graphics system uses for positioning. Later an AI system will probably operate on it too
type PhysicsComponent struct {
	Position    Vector2
	Motion      Vector2
	Velocity    Vector2
	Forces      map[Force]*Vector2
	Height      float64
	Width       float64
	Obstructive bool
	Grounded    bool
}

type ThingInstance struct {
	Type    Thing
	Physics PhysicsComponent
}

// enums
type AnimationType int
type Direction int
type Thing int
type Force int
type AudioID int
type AudioType int

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
	// force type enums
	GRAVITY Force = iota + 1
	JUMP_FORCE
	KNOCKBACK
	// sound enums
	SOUND_RS AudioID   = iota + 1
	MUSIC    AudioType = iota + 1
	SFX
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
	// speeds
	PLAYER_WALK_SPEED float64 = 0.09
	// the maximum distance at which a sound source actually creates sound
	EARSHOT float64 = 10
)

//the player that the viewport centers around and the inputs control
var MainPlayer Player = Player{
	WalkingState:           STANDING,
	WalkingAnimationFrame:  0,
	WalkingAnimationFrames: VIBE_FRAMES,
	MovingLeft:             false,
	MovingRight:            false,
	Physics: PhysicsComponent{
		Position: Vector2{
			X: 70, // start in the middle-ish of the world
			Y: 0,
		},
		Forces: map[Force]*Vector2{
			GRAVITY:    {X: 0, Y: 0},
			JUMP_FORCE: {X: 0, Y: 0},
		},
		Height:   1,
		Width:    PLAYER_WIDTH,
		Grounded: false,
	},
}

// the grid of static things, so that only the nearby ones are "loaded"
var Grid map[Coordinate][]ThingInstance

//global variables for the ebiten library
var (
	StanceContinuations   []StanceContinuation
	VP                    Viewport
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

func ResizeImage(img *ebiten.Image, drawoptions *ebiten.DrawImageOptions, w float64, h float64) {
	width_int, height_int := img.Size()
	drawoptions.GeoM.Scale(w/float64(width_int), h/float64(height_int))
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

func GetContinuation(s Stance) (Stance, int) {
	// yeah this is sorta bad lol
	for i := 0; i < len(StanceContinuations); i++ {
		if StanceContinuations[i].Start == s {
			return StanceContinuations[i].Continuation, StanceContinuations[i].Frames
		}
	}
	return s, 0
}

func ChangeWalkState(player *Player, state AnimationType, new_stance Stance, frames int) {
	player.WalkingState = state
	player.WalkingStanceTo = new_stance
	player.WalkingAnimationFrames = frames
	// reset the animation clock to transition into the new stance, starting from however the player is poisitioned now
	player.WalkingStanceFrom = player.Stance
	player.WalkingAnimationFrame = 0
}

func GetChunk(p Player) (chunk Chunk) {
	chunk = Chunk{StartX: 0, StartY: 0, EndX: int(math.Floor(PLACE_WIDTH)), EndY: int(math.Floor(PLACE_HEIGHT))}
	// if the player isnt too close to the edges, shift each of the sides towards the player to construct a box around the player that's just out of view of the human player
	if math.Floor(p.Physics.Position.Y)-math.Floor(0.75*SCREEN_HEIGHT/PIXEL_YARD_RATIO) > 0 {
		chunk.StartY = int(math.Floor(p.Physics.Position.Y) - math.Floor(0.75*SCREEN_HEIGHT/PIXEL_YARD_RATIO))
	}
	if math.Floor(p.Physics.Position.Y+1)+math.Floor(0.75*SCREEN_HEIGHT/PIXEL_YARD_RATIO) < math.Floor(PLACE_HEIGHT) {
		chunk.EndY = int(math.Floor(p.Physics.Position.Y+1) + math.Floor(0.75*SCREEN_HEIGHT/PIXEL_YARD_RATIO))
	}
	if math.Floor(p.Physics.Position.X)-math.Floor(0.75*SCREEN_WIDTH/PIXEL_YARD_RATIO) > 0 {
		chunk.StartX = int(math.Floor(p.Physics.Position.X) - math.Floor(0.75*SCREEN_WIDTH/PIXEL_YARD_RATIO))
	}
	if math.Floor(p.Physics.Position.X+1)+math.Floor(0.75*SCREEN_WIDTH/PIXEL_YARD_RATIO) < math.Floor(PLACE_WIDTH) {
		chunk.EndX = int(math.Floor(p.Physics.Position.X+1) + math.Floor(0.75*SCREEN_WIDTH/PIXEL_YARD_RATIO))
	}
	return
}

func IK(first_bone_length float64, second_bone_length float64, base_x float64, base_y float64, target_x float64, target_y float64, concave_up bool) (new_base_joint_angle float64, new_connector_joint_angle float64) {
	x_dif := target_x - base_x
	y_dif := target_y - base_y
	d := math.Sqrt(math.Pow(x_dif, 2) + math.Pow(y_dif, 2))
	if d >= first_bone_length+second_bone_length {
		new_base_joint_angle = -math.Asin(x_dif / d)
		new_connector_joint_angle = 0.0
	} else {
		concavity_coefficient := 1.0
		if concave_up {
			concavity_coefficient = -1.0
		}
		angleTargetBaseConnector := concavity_coefficient * math.Acos((math.Pow(second_bone_length, 2)-math.Pow(first_bone_length, 2)-math.Pow(d, 2))/(-2*first_bone_length*d)) // law of cosines
		angleYAxisPelvisFoot := math.Acos(x_dif / d)
		angleTargetConnectorBase := math.Asin(d * math.Sin(angleTargetBaseConnector) / second_bone_length) // law of sines
		new_base_joint_angle = math.Mod(-angleTargetBaseConnector+angleYAxisPelvisFoot-math.Pi/2, 2*math.Pi)
		new_connector_joint_angle = math.Mod(angleTargetConnectorBase, 2*math.Pi)
		calculated_target_x := first_bone_length*math.Cos(new_base_joint_angle) + second_bone_length*math.Cos(new_connector_joint_angle)
		if math.Abs(calculated_target_x) > math.Abs(base_x-target_x) {
			angle_differential := 2 * (new_connector_joint_angle - math.Pi/2)
			new_connector_joint_angle = math.Mod(new_connector_joint_angle-angle_differential, 2*math.Pi)
		}
	}
	return new_base_joint_angle, new_connector_joint_angle
}

// func CurrentGround(player Player) GroundPiecewise {
// 	out := GroundPiecewise{Pieces: [5]LineEquation{}}
// 	groundchunk := Chunk{StartX: int(player.X) - 2, EndX: int(player.X) + 2, StartY: int(player.Y-player.Height) - 1, EndY: int(player.Y)}
// 	for x := groundchunk.StartX; x < groundchunk.EndX; x++ {
// 		found := false
// 		for y := groundchunk.EndY; y > groundchunk.StartY; y-- {
// 			chunklet := Grid[Coordinate{x, y}]
// 			for i := 0; i < len(chunklet); i++ {
// 				if chunklet[i] == OAK_LOG {
// 					out.SetPiece(LineEquation{Slope: 0, YIntercept: OAK_LOG_HEIGHT}, x-groundchunk.StartX)
// 					found = true
// 					break
// 				}
// 			}
// 			if found {
// 				break
// 			}
// 		}
// 		if !found {
// 			out.SetPiece(LineEquation{Slope: 0, YIntercept: 0}, x-groundchunk.StartX)
// 		}
// 	}
// 	return out
// }

// func ArrayIncludes(array []interface{}, item interface{}) bool {
// 	for i := 0; i < len(array); i++ {
// 		if array[i] == item {
// 			return true
// 		}
// 	}
// 	return false
// }
