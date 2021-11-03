package stances

import (
	"math"

	"ryanbrewer.page/core"
)

var (
	RestRight1 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 16,
		LeftUpperArm:  math.Pi / 16,
		Direction:     core.RIGHT,
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
	RestRight2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 20,
		LeftUpperArm:  math.Pi / 20,
		Direction:     core.RIGHT,
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
	RestLeft1 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 16,
		LeftUpperArm:  math.Pi / 16,
		Direction:     core.LEFT,
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
	RestLeft2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 20,
		LeftUpperArm:  math.Pi / 20,
		Direction:     core.LEFT,
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
	WalkRight1 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     core.RIGHT,
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
	WalkRight2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     core.RIGHT,
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
	WalkLeft1 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 6,
		LeftUpperArm:  -math.Pi / 6,
		RightUpperLeg: -math.Pi / 6,
		LeftUpperLeg:  math.Pi / 6,
		Direction:     core.LEFT,
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
	WalkLeft2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 6,
		LeftUpperArm:  math.Pi / 6,
		RightUpperLeg: math.Pi / 6,
		LeftUpperLeg:  -math.Pi / 6,
		Direction:     core.LEFT,
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
	JumpRight1 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: -math.Pi / 4,
		LeftUpperLeg:  -math.Pi / 4,
		Direction:     core.RIGHT,
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
	JumpRight2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 10,
		LeftUpperArm:  math.Pi / 5,
		RightUpperLeg: -math.Pi / 10,
		LeftUpperLeg:  -math.Pi / 15,
		Direction:     core.RIGHT,
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
	JumpRight3 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: -math.Pi / 4,
		LeftUpperLeg:  -math.Pi / 4,
		Direction:     core.RIGHT,
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
	JumpLeft1 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: math.Pi / 4,
		LeftUpperLeg:  math.Pi / 4,
		Direction:     core.LEFT,
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
	JumpLeft2 core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 10,
		LeftUpperArm:  math.Pi / 5,
		RightUpperLeg: math.Pi / 10,
		LeftUpperLeg:  math.Pi / 15,
		Direction:     core.LEFT,
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
	JumpLeft3 core.Stance = core.Stance{
		RightUpperArm: math.Pi / 20,
		LeftUpperArm:  math.Pi / 8,
		RightUpperLeg: math.Pi / 4,
		LeftUpperLeg:  math.Pi / 4,
		Direction:     core.LEFT,
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
	LeapRight core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 3,
		LeftUpperArm:  math.Pi / 3,
		LeftUpperLeg:  -math.Pi / 2,
		RightUpperLeg: math.Pi / 4,
		Direction:     core.RIGHT,
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
	LeapLeft core.Stance = core.Stance{
		RightUpperArm: -math.Pi / 3,
		LeftUpperArm:  math.Pi / 3,
		LeftUpperLeg:  math.Pi / 2,
		RightUpperLeg: -math.Pi / 4,
		Direction:     core.LEFT,
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
)

func CreateStanceContinuation(s1 core.Stance, s2 core.Stance, f int) {
	var c core.StanceContinuation = core.StanceContinuation{
		Start:        s1,
		Continuation: s2,
		Frames:       f,
	}
	core.StanceContinuations = append(core.StanceContinuations, c)
}

func CreateStanceContinuations() {
	CreateStanceContinuation(RestLeft1, RestLeft2, core.VIBE_FRAMES)
	CreateStanceContinuation(RestRight1, RestRight2, core.VIBE_FRAMES)
	CreateStanceContinuation(RestLeft2, RestLeft1, core.VIBE_FRAMES)
	CreateStanceContinuation(RestRight2, RestRight1, core.VIBE_FRAMES)
	CreateStanceContinuation(JumpLeft1, JumpLeft2, core.JUMP_TRANSITION_FRAMES)
	CreateStanceContinuation(JumpRight1, JumpRight2, core.JUMP_TRANSITION_FRAMES)
	CreateStanceContinuation(JumpLeft2, JumpLeft3, core.JUMP_TIME_FRAMES)
	CreateStanceContinuation(JumpRight2, JumpRight3, core.JUMP_TIME_FRAMES)
	CreateStanceContinuation(LeapLeft, JumpLeft2, core.JUMP_TRANSITION_FRAMES)
	CreateStanceContinuation(LeapRight, JumpRight2, core.JUMP_TRANSITION_FRAMES)
	CreateStanceContinuation(WalkLeft1, WalkLeft2, core.STEP_FRAMES)
	CreateStanceContinuation(WalkRight1, WalkRight2, core.STEP_FRAMES)
	CreateStanceContinuation(WalkLeft2, WalkLeft1, core.STEP_FRAMES)
	CreateStanceContinuation(WalkRight2, WalkRight1, core.STEP_FRAMES)
}
