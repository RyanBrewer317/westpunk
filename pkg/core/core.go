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
}

var (
	RestPose     Stance = Stance{0.01, 0.01, 5 * math.Pi / 6, -1, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01}
	PlayerStance Stance = RestPose
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
