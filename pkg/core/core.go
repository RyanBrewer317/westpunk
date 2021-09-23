package core

import (
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
var WalkingStanceTo Stance
var WalkingStanceFrom Stance
var MovingRight bool = false
var MovingLeft bool = false

type Direction int

const (
	Right Direction = iota + 1
	Left
)

type Thing int

const (
	Oak Thing = iota + 1
	OakLog
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
	OakLogWidth  float64 = 0.3
	OakLogHeight float64 = 0.3
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
	OakLogImg             *ebiten.Image
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
