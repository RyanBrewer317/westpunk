package core

import (
	"math"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

type Vertex struct {
	X, Y int
}

type Viewport struct {
	X float64
	Y float64
	W float64
	H float64
}

var VP Viewport

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

const AnimationSpeed int = 10

var (
	RestRight1 Stance = Stance{
		RightUpperArm: -math.Pi / 16,
		LeftUpperArm:  math.Pi / 16,
		Direction:     Right,
		RightUpperLeg: 0,
		LeftUpperLeg:  0,
		Head:          0,
		Torso:         0,
		RightLowerArm: 0,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  0,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	RestRight2 Stance = Stance{
		RightUpperArm: -math.Pi / 20,
		LeftUpperArm:  math.Pi / 20,
		Direction:     Right,
		RightUpperLeg: 0,
		LeftUpperLeg:  0,
		Head:          0,
		Torso:         0,
		RightLowerArm: 0,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  0,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	RestLeft1 Stance = Stance{
		RightUpperArm: -math.Pi / 16,
		LeftUpperArm:  math.Pi / 16,
		Direction:     Left,
		RightUpperLeg: 0,
		LeftUpperLeg:  0,
		Head:          0,
		Torso:         0,
		RightLowerArm: 0,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  0,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	RestLeft2 Stance = Stance{
		RightUpperArm: -math.Pi / 20,
		LeftUpperArm:  math.Pi / 20,
		Direction:     Left,
		RightUpperLeg: 0,
		LeftUpperLeg:  0,
		Head:          0,
		Torso:         0,
		RightLowerArm: 0,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  0,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	WalkRight1 Stance = Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     Right,
		Head:          0,
		Torso:         0,
		RightLowerArm: -math.Pi / 20,
		RightLowerLeg: math.Pi / 20,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 20,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	WalkRight2 Stance = Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     Right,
		Head:          0,
		Torso:         0,
		RightLowerArm: -math.Pi / 20,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 20,
		LeftLowerLeg:  math.Pi / 20,
		LeftFoot:      0,
		Weapon:        0,
	}
	WalkLeft1 Stance = Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     Left,
		Head:          0,
		Torso:         0,
		RightLowerArm: math.Pi / 20,
		RightLowerLeg: -math.Pi / 20,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 20,
		LeftLowerLeg:  0,
		LeftFoot:      0,
		Weapon:        0,
	}
	WalkLeft2 Stance = Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     Left,
		Head:          0,
		Torso:         0,
		RightLowerArm: math.Pi / 20,
		RightLowerLeg: 0,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 20,
		LeftLowerLeg:  -math.Pi / 20,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpRight1 Stance = Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: -math.Pi / 4,
		LeftUpperLeg:  -math.Pi / 4,
		Direction:     Right,
		Head:          0,
		Torso:         math.Pi / 10,
		RightLowerArm: -math.Pi / 3,
		RightLowerLeg: math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 3,
		LeftLowerLeg:  math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpRight2 Stance = Stance{
		RightUpperArm: -math.Pi / 10,
		LeftUpperArm:  math.Pi / 5,
		RightUpperLeg: -math.Pi / 10,
		LeftUpperLeg:  -math.Pi / 15,
		Direction:     Right,
		Head:          0,
		Torso:         0,
		RightLowerArm: -math.Pi / 5,
		RightLowerLeg: math.Pi / 5,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 5,
		LeftLowerLeg:  math.Pi / 5,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpRight3 Stance = Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: -math.Pi / 4,
		LeftUpperLeg:  -math.Pi / 4,
		Direction:     Right,
		Head:          math.Pi / 20,
		Torso:         math.Pi / 10,
		RightLowerArm: -math.Pi / 3,
		RightLowerLeg: math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 3,
		LeftLowerLeg:  math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpLeft1 Stance = Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: math.Pi / 4,
		LeftUpperLeg:  math.Pi / 4,
		Direction:     Left,
		Head:          0,
		Torso:         -math.Pi / 10,
		RightLowerArm: math.Pi / 3,
		RightLowerLeg: -math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 3,
		LeftLowerLeg:  -math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpLeft2 Stance = Stance{
		RightUpperArm: -math.Pi / 10,
		LeftUpperArm:  math.Pi / 5,
		RightUpperLeg: math.Pi / 10,
		LeftUpperLeg:  math.Pi / 15,
		Direction:     Left,
		Head:          0,
		Torso:         0,
		RightLowerArm: math.Pi / 5,
		RightLowerLeg: -math.Pi / 5,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 5,
		LeftLowerLeg:  -math.Pi / 5,
		LeftFoot:      0,
		Weapon:        0,
	}
	JumpLeft3 Stance = Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: math.Pi / 4,
		LeftUpperLeg:  math.Pi / 4,
		Direction:     Left,
		Head:          -math.Pi / 20,
		Torso:         -math.Pi / 10,
		RightLowerArm: math.Pi / 3,
		RightLowerLeg: -math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 3,
		LeftLowerLeg:  -math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	LeapRight Stance = Stance{
		RightUpperArm: -math.Pi / 3,
		LeftUpperArm:  math.Pi / 3,
		LeftUpperLeg:  -math.Pi / 2,
		RightUpperLeg: math.Pi / 4,
		Direction:     Right,
		Head:          0,
		Torso:         0,
		RightLowerArm: -math.Pi / 3,
		LeftLowerLeg:  math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  -math.Pi / 3,
		RightLowerLeg: math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	LeapLeft Stance = Stance{
		RightUpperArm: -math.Pi / 3,
		LeftUpperArm:  math.Pi / 3,
		LeftUpperLeg:  math.Pi / 2,
		RightUpperLeg: -math.Pi / 4,
		Direction:     Left,
		Head:          0,
		Torso:         0,
		RightLowerArm: math.Pi / 3,
		LeftLowerLeg:  -math.Pi / 6,
		RightFoot:     0,
		LeftLowerArm:  math.Pi / 3,
		RightLowerLeg: -math.Pi / 6,
		LeftFoot:      0,
		Weapon:        0,
	}
	PlayerStance Stance = RestRight1
)

const (
	WalkTransitionFrames = 20
	StepFrames           = 25
	VibeFrames           = 50
	JumpTransitionFrames = 10
	JumpTimeFrames       = 10
)

type AnimationType int

const (
	Standing AnimationType = iota + 1
	WalkingRight
	WalkingLeft
	JumpingRight
	JumpingLeft
	LeapingRight
	LeapingLeft
)

var WalkingState AnimationType = Standing
var WalkingAnimationFrame int = 0
var WalkingAnimationFrames int = VibeFrames
var WalkingStanceTo Stance = RestRight1
var WalkingStanceFrom Stance = RestRight1

type Direction int

const (
	Right Direction = iota + 1
	Left
)

type Thing int

const (
	Oak Thing = iota + 1
)

const (
	TorsoWidth     float64 = 0.25
	TorsoHeight    float64 = 0.5
	HeadWidth      float64 = 0.35
	HeadHeight     float64 = 0.3
	UpperArmHeight float64 = 0.25
	UpperArmWidth  float64 = 0.05
	LowerArmHeight float64 = 0.25
	LowerArmWidth  float64 = 0.05
	UpperLegWidth  float64 = 0.05
	UpperLegHeight float64 = 0.25
	LowerLegWidth  float64 = 0.05
	LowerLegHeight float64 = 0.25
)

const (
	PlaceWidth   float64 = 256
	PlaceHeight  float64 = 128
	PlayerWidth  float64 = 0.5
	GroundHeight float64 = ScreenHeight / PixelYardRatio
	GroundY      float64 = 0
	OakHeight    float64 = 5
	OakWidth     float64 = 2
)

const ScreenHeight float64 = 540
const ScreenWidth float64 = 810

const PixelYardRatio float64 = 70

var PlayerHeight float64 = 1

var Grid map[Vertex][]Thing
var PlayerYVelocity float64 = 0
var PlayerXVelocity float64 = 0

var PlayerX float64 = 70
var PlayerY float64 = 6

var (
	PlayerImg             *ebiten.Image
	GrassLayerImg         *ebiten.Image
	DirtLayerImg          *ebiten.Image
	OakImg                *ebiten.Image
	BackgroundImg         *ebiten.Image
	PlayerDrawOptions     ebiten.DrawImageOptions
	BackgroundDrawOptions ebiten.DrawImageOptions
)

func GetPXY(y float64) float64 {
	return ScreenHeight - (y * PixelYardRatio)
}

func DrawImage(screen *ebiten.Image, img *ebiten.Image, drawoptions ebiten.DrawImageOptions, x float64, y float64) {
	drawoptions.GeoM.Translate(x, y)
	screen.DrawImage(img, &drawoptions)
}

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
