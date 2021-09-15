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
	RestPose Stance = Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
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
	WalkRight1 Stance = Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     Right,
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
	WalkRight2 Stance = Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     Right,
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
	WalkLeft1 Stance = Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     Left,
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
	WalkLeft2 Stance = Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     Left,
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
	PlayerStance Stance = Stance{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, Right}
)

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
	HeadWidth      float64 = 0.25
	HeadHeight     float64 = 0.3
	UpperArmHeight float64 = 0.35
	UpperArmWidth  float64 = 0.1
	LowerArmHeight float64 = 0.35
	LowerArmWidth  float64 = 0.1
	UpperLegWidth  float64 = 0.1
	UpperLegHeight float64 = 0.35
	LowerLegWidth  float64 = 0.1
	LowerLegHeight float64 = 0.35
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

const PixelYardRatio float64 = 100

var PlayerHeight float64 = 1

var Grid map[Vertex][]Thing
var PlayerYVelocity float64 = 0
var PlayerXVelocity float64 = 0

var PlayerX float64 = 70
var PlayerY float64 = 6

var (
	PlayerImg         *ebiten.Image
	GroundImg         *ebiten.Image
	OakImg            *ebiten.Image
	PlayerDrawOptions ebiten.DrawImageOptions
	GroundDrawOptions ebiten.DrawImageOptions
)

func GetPXY(y float64) float64 {
	return ScreenHeight - (y * PixelYardRatio)
}

func DrawImage(screen *ebiten.Image, img *ebiten.Image, drawoptions ebiten.DrawImageOptions, x float64, y float64) {
	drawoptions.GeoM.Translate(x, y)
	screen.DrawImage(img, &drawoptions)
}

var Clock int = 0

func ShiftStance(stance1 Stance, stance2 Stance) Stance {
	c := float64(Clock%(2*AnimationSpeed)) / float64(2*AnimationSpeed)
	s1 := stance1
	s2 := stance2
	if c > 0.5 {
		c -= 0.5
		tmp := s1
		s1 = s2
		s2 = tmp
	}
	return Stance{
		Head:          c * (s2.Head - s1.Head),
		Torso:         c * (s2.Torso - s1.Torso),
		RightUpperArm: c * (s2.RightUpperArm - s1.RightUpperArm),
		LeftUpperArm:  c * (s2.LeftUpperArm - s1.LeftUpperArm),
		RightLowerArm: c * (s2.RightLowerArm - s1.RightLowerArm),
		LeftLowerArm:  c * (s2.LeftLowerArm - s1.LeftLowerArm),
		RightUpperLeg: c * (s2.RightUpperLeg - s1.RightUpperLeg),
		LeftUpperLeg:  c * (s2.LeftUpperLeg - s1.LeftUpperLeg),
		RightLowerLeg: c * (s2.RightLowerLeg - s1.RightLowerLeg),
		LeftLowerLeg:  c * (s2.LeftLowerLeg - s1.LeftLowerLeg),
		RightFoot:     c * (s2.RightFoot - s1.RightFoot),
		LeftFoot:      c * (s2.LeftFoot - s1.LeftFoot),
		Weapon:        c * (s2.Weapon - s1.Weapon),
		Direction:     s1.Direction,
	}
}
