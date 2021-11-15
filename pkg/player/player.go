package player

import (
	"image"
	"math"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"ryanbrewer.page/core"
	"ryanbrewer.page/stances"
)

func DrawPlayer(screen *ebiten.Image, player core.Player, x float64, y float64) {
	// draw the limbs in a different order based on which way the player is facing
	if player.Stance.Direction == core.DIRECTION_RIGHT {
		draw_player_right_arm(screen, player, x, y)
		draw_player_right_leg(screen, player, x, y)
		draw_player_torso(screen, player, x, y)
		draw_player_head(screen, player, x, y)
		draw_player_left_leg(screen, player, x, y)
		draw_player_left_arm(screen, player, x, y)
	} else if player.Stance.Direction == core.DIRECTION_LEFT {
		draw_player_left_arm(screen, player, x, y)
		draw_player_left_leg(screen, player, x, y)
		draw_player_torso(screen, player, x, y)
		draw_player_head(screen, player, x, y)
		draw_player_right_leg(screen, player, x, y)
		draw_player_right_arm(screen, player, x, y)
	}
}

func draw_player_head(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	// rotate the drawing context by the torso and head angles
	theta := math.Mod(player.Stance.Head+player.Stance.Torso, 2*math.Pi)
	torso_theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	if torso_theta < 0 {
		torso_theta += 2 * math.Pi
	}
	// calculate the position of the neck joint
	neckx := x/core.PIXEL_YARD_RATIO - (core.HEAD_WIDTH-core.TORSO_WIDTH)/2
	necky := (y / core.PIXEL_YARD_RATIO)
	difx, dify := torso_rotation_diff(core.HEAD_WIDTH/2, player) // adjust for torso rotation
	neckx += difx
	necky += dify
	// calculate the top left corner of the head based on the location of the neck joint and the angle of the head
	headx := neckx - (core.HEAD_WIDTH * math.Cos(theta) / 2) + (core.HEAD_HEIGHT * math.Sin(theta))
	heady := necky + (core.HEAD_WIDTH * math.Sin(theta) / 2) + (core.HEAD_HEIGHT * math.Cos(theta))
	draw_player_piece(screen, 0, 0, 131, 104, player.DrawOptions, headx, heady, core.HEAD_WIDTH, core.HEAD_HEIGHT, theta, player.Stance.Direction)
}

func torso_rotation_diff(r float64, player core.Player) (float64, float64) {
	// this function takes a point along the shoulders and calculates where it would be once the torso is rotated
	// the point is given by traveling from the left shoulder to the right a distance of r
	theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	return limb_joint_to_corner(theta, 0, 0, -2*r) // the math is identical to traveling from a point on a rotated rectangle to a corner on the same side
}

func draw_player_torso(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	draw_player_piece(screen, 165, 0, 300, 240, player.DrawOptions, x/core.PIXEL_YARD_RATIO, y/core.PIXEL_YARD_RATIO, core.TORSO_WIDTH, core.TORSO_HEIGHT, theta, player.Stance.Direction)
}

func draw_player_left_arm(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	// rotate the drawing context by the torso and left upper arm angles
	theta := math.Mod(player.Stance.LeftUpperArm+player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	leftshoulderx := x / core.PIXEL_YARD_RATIO
	leftshouldery := y / core.PIXEL_YARD_RATIO
	// calculate the top left corner of the upper arm based on the location of the shoulder joint
	leftupperarmx := leftshoulderx - (core.UPPER_ARM_WIDTH * math.Cos(theta) / 2)
	leftupperarmy := leftshouldery + (core.UPPER_ARM_WIDTH * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftupperarmx, leftupperarmy, core.UPPER_ARM_WIDTH, core.UPPER_ARM_HEIGHT, theta, player.Stance.Direction)

	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.LeftLowerArm, 2*math.Pi)
	// rotate the drawing context an additional amount, the left lower arm angle
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	// calculate the position of the elbow joint
	leftelbowx := leftshoulderx - (core.UPPER_ARM_HEIGHT * math.Sin(theta))
	leftelbowy := leftshouldery - (core.UPPER_ARM_HEIGHT * math.Cos(theta))
	// calculate the top left corner of the left lower arm based on the location of the elbow joint
	leftlowerarmx, leftlowerarmy := limb_joint_to_corner(theta2, leftelbowx, leftelbowy, core.LOWER_ARM_WIDTH)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftlowerarmx, leftlowerarmy, core.LOWER_ARM_WIDTH, core.LOWER_ARM_HEIGHT, theta2, player.Stance.Direction)
}

func draw_player_right_arm(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	// rotate the drawing context by the angles of the torso and right upper arm
	theta := math.Mod(player.Stance.Torso+player.Stance.RightUpperArm, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	// calculate the right shoulder joint
	difx, dify := torso_rotation_diff(core.TORSO_WIDTH, player)
	rightshoulderx := x/core.PIXEL_YARD_RATIO + difx
	rightshouldery := y/core.PIXEL_YARD_RATIO + dify
	// calculate the top left corner of the right upper arm based on the right shoulder joint
	rightupperarmx := rightshoulderx - (core.UPPER_ARM_WIDTH * math.Cos(theta) / 2)
	rightupperarmy := rightshouldery + (core.UPPER_ARM_WIDTH * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightupperarmx, rightupperarmy, core.UPPER_ARM_WIDTH, core.UPPER_ARM_HEIGHT, theta, player.Stance.Direction)

	// rotate the drawing context an additional amount, the angle of the right lower arm
	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(player.Stance.Torso+player.Stance.RightUpperArm+player.Stance.RightLowerArm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	//calculate the right elbow joint
	rightelbowx := rightshoulderx - (core.UPPER_ARM_HEIGHT * math.Sin(theta))
	rightelbowy := rightshouldery - (core.UPPER_ARM_HEIGHT * math.Cos(theta))
	// calculate the top left corner of the right lower arm based on the right elbow joint
	rightforearmx, rightforearmy := limb_joint_to_corner(theta2, rightelbowx, rightelbowy, core.UPPER_ARM_WIDTH)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightforearmx, rightforearmy, core.LOWER_ARM_WIDTH, core.LOWER_ARM_HEIGHT, theta2, player.Stance.Direction)
}

func draw_player_left_leg(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	// rotate the drawing context by the angles of the torso and left upper leg
	theta := math.Mod(player.Stance.Torso+player.Stance.LeftUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	// calculate the location of the left pelvis joint
	difx, dify := torso_rotation_diff(core.UPPER_LEG_WIDTH/2, player)
	dify2, difx2 := torso_rotation_diff(core.TORSO_HEIGHT, player) // I reverse x and y here so that the argument r represents travelling down the body now instead of to the right
	pelvis_left_x := x/core.PIXEL_YARD_RATIO + difx + difx2
	pelvis_left_y := y/core.PIXEL_YARD_RATIO + dify - dify2
	// calculate the top left corner of the left upper leg based on the left pelvis joint
	leftupperlegx := pelvis_left_x - (core.UPPER_LEG_WIDTH * math.Cos(theta) / 2)
	leftupperlegy := pelvis_left_y + (core.UPPER_LEG_WIDTH * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftupperlegx, leftupperlegy, core.UPPER_LEG_WIDTH, core.UPPER_LEG_HEIGHT, theta, player.Stance.Direction)

	// rotate the drawing context by an additional amount, the left lower leg angle
	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.LeftLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	// calculate the left knee joint
	leftkneex := pelvis_left_x - (core.UPPER_LEG_HEIGHT * math.Sin(theta))
	leftkneey := pelvis_left_y - (core.UPPER_LEG_HEIGHT * math.Cos(theta))
	// calculate the top left corner of the left lower leg based on the left knee joint
	leftlowerlegx, leftlowerlegy := limb_joint_to_corner(theta2, leftkneex, leftkneey, core.LOWER_LEG_WIDTH)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftlowerlegx, leftlowerlegy, core.LOWER_LEG_WIDTH, core.LOWER_LEG_HEIGHT, theta2, player.Stance.Direction)
}

func draw_player_right_leg(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	// rotate the drawing context by the angles of the torso and right upper leg
	theta := math.Mod(player.Stance.Torso+player.Stance.RightUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	// calculate the right pelvis joint
	difx, dify := torso_rotation_diff(core.TORSO_WIDTH-core.UPPER_LEG_WIDTH/2, player)
	dify2, difx2 := torso_rotation_diff(core.TORSO_HEIGHT, player) // I reverse x and y here so that the argument r represents travelling down the body now instead of to the right
	pelvis_right_x := x/core.PIXEL_YARD_RATIO + difx + difx2
	pelvis_right_y := y/core.PIXEL_YARD_RATIO + dify - dify2
	// calculate the top left corner of the right upper leg based on the right pelvis joint
	rightupperlegx := pelvis_right_x - (core.UPPER_LEG_WIDTH * math.Cos(theta) / 2)
	rightupperlegy := pelvis_right_y + (core.UPPER_LEG_WIDTH * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightupperlegx, rightupperlegy, core.UPPER_LEG_WIDTH, core.UPPER_LEG_HEIGHT, theta, player.Stance.Direction)

	// rotate the drawing context by an additional amount, the right lower leg angle
	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.RightLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	// calculate the right knee joint
	rightkneex := pelvis_right_x - (core.UPPER_LEG_HEIGHT * math.Sin(theta))
	rightkneey := pelvis_right_y - (core.UPPER_LEG_HEIGHT * math.Cos(theta))
	// calculate the top left corner of the right lower leg based on the right knee joint
	rightlowerlegx, rightlowerlegy := limb_joint_to_corner(theta2, rightkneex, rightkneey, core.LOWER_LEG_WIDTH)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightlowerlegx, rightlowerlegy, core.LOWER_LEG_WIDTH, core.LOWER_LEG_HEIGHT, theta2, player.Stance.Direction)
}

func draw_player_piece(screen *ebiten.Image, imgx1 int, imgy1 int, imgx2 int, imgy2 int, drawoptions ebiten.DrawImageOptions, igx float64, igy float64, igw float64, igh float64, theta float64, direction core.Direction) {
	// slice the image out of the spritesheet
	img := ebiten.NewImageFromImage(core.PlayerImg.SubImage(image.Rect(imgx1, imgy1, imgx2, imgy2)))
	// resize the image
	wi, hi := img.Size()
	w := igw * core.PIXEL_YARD_RATIO
	h := igh * core.PIXEL_YARD_RATIO
	direction_scale := 1.0
	direction_translation := 0.0
	if direction == core.DIRECTION_LEFT {
		direction_scale = -1                // reverse the x scaling if the player is facing left
		direction_translation = float64(wi) // translate by the width of the piece if the player is facing left
	}
	drawoptions.GeoM.Scale(direction_scale, 1)
	drawoptions.GeoM.Translate(direction_translation, 0)
	drawoptions.GeoM.Scale(w/float64(wi), h/float64(hi))
	// scale the image
	drawoptions.GeoM.Rotate(theta)
	// translation happens in core.DrawImage
	core.DrawImage(screen, img, drawoptions, igx*core.PIXEL_YARD_RATIO-core.VP.X, core.GetPXY(igy)+core.VP.Y)
}

func limb_joint_to_corner(theta float64, jx float64, jy float64, w float64) (float64, float64) {
	// calculate the point that's distance w/2 away from (jx, jy) in the direction theta
	x := jx - (w/2)*math.Cos(theta)
	y := jy + (w/2)*math.Sin(theta)
	return x, y
}

func SetPlayerHeight(p *core.Player) {
	//calculate the players height (todo: add math for arms in case the player is doing a handstand or crawling or something?)
	// length of the right leg
	right_upper_leg_angle := p.Stance.Torso + p.Stance.RightUpperLeg
	right_upper_leg_height := core.UPPER_LEG_HEIGHT * math.Cos(right_upper_leg_angle)
	right_lower_leg_angle := p.Stance.Torso + p.Stance.RightUpperLeg + p.Stance.RightLowerLeg
	right_lower_leg_height := core.LOWER_LEG_HEIGHT * math.Cos(right_lower_leg_angle)
	right_leg_height := right_upper_leg_height + right_lower_leg_height
	// length of the left leg
	left_upper_leg_angle := p.Stance.Torso + p.Stance.LeftUpperLeg
	left_upper_leg_height := core.UPPER_LEG_HEIGHT * math.Cos(left_upper_leg_angle)
	left_lower_leg_angle := p.Stance.Torso + p.Stance.LeftUpperLeg + p.Stance.LeftLowerLeg
	left_lower_leg_height := core.LOWER_LEG_HEIGHT * math.Cos(left_lower_leg_angle)
	left_leg_height := left_upper_leg_height + left_lower_leg_height
	// length (height) of the whole body
	p.Physics.Height = core.TORSO_HEIGHT*math.Cos(p.Stance.Torso) + math.Max(right_leg_height, left_leg_height)
}

func StopMovingRight(p *core.Player) {
	if p.MovingLeft { // if they're still holding down the key showing intent to walk left
		core.ChangeWalkState(p, core.ANIMAtION_TYPE_WALKING_LEFT, stances.WalkLeft1, core.WALK_TRANSITION_FRAMES)
	} else { // if they have no intent of walking in either direciton
		core.ChangeWalkState(p, core.ANIMATION_TYPE_STANDING, stances.RestRight1, core.WALK_TRANSITION_FRAMES)
	}
	p.MovingRight = false
	// swap what walk1 and walk2 are referring to, so that spamming the walk key still makes the legs try to cross each time
	tmp := stances.WalkRight1
	stances.WalkRight1 = stances.WalkRight2
	stances.WalkRight2 = tmp
}

func StopMovingLeft(p *core.Player) {
	if p.MovingRight { // if they're still holding down the key showing intent to walk right
		core.ChangeWalkState(p, core.ANIMATION_TYPE_WALKING_RIGHT, stances.WalkRight1, core.WALK_TRANSITION_FRAMES)
	} else { // if they have no intent of walking in either direction
		core.ChangeWalkState(p, core.ANIMATION_TYPE_STANDING, stances.RestLeft1, core.WALK_TRANSITION_FRAMES)
	}
	p.MovingLeft = false
	// swap what walk1 and walk2 are referring to, so that spamming the walk key still makes the legs try to cross each time
	tmp := stances.WalkLeft1
	stances.WalkLeft1 = stances.WalkLeft2
	stances.WalkLeft2 = tmp
}

func ContinueStance(p *core.Player) {
	// shift from the current stance to the next
	new_stance, frames := core.GetContinuation(p.WalkingStanceTo)
	core.ChangeWalkState(p, p.WalkingState, new_stance, frames)
}

func StartJump(p *core.Player) {
	// if the player is walking to the right, they're now leaping to the right
	if p.Stance.Direction == core.DIRECTION_RIGHT {
		if p.WalkingState == core.ANIMATION_TYPE_WALKING_RIGHT {
			core.ChangeWalkState(p, core.ANIMATION_TYPE_LEAPING_RIGHT, stances.LeapRight, core.JUMP_TRANSITION_FRAMES)
		} else { // if they arent walking, the jump is straight up and down, facing right
			core.ChangeWalkState(p, core.ANIMATION_TYPE_JUMPING_RIGHT, stances.JumpRight1, core.JUMP_TRANSITION_FRAMES)
		}
	} else { // if the player is walking to the left, theyre now leaping to the left
		if p.WalkingState == core.ANIMAtION_TYPE_WALKING_LEFT {
			core.ChangeWalkState(p, core.ANIMATION_TYPE_LEAPING_LEFT, stances.LeapLeft, core.JUMP_TRANSITION_FRAMES)
		} else { // if they arent walking, the jump is straight up and down, facing left
			core.ChangeWalkState(p, core.ANIMAtiON_TYPE_JUMPING_LEFT, stances.JumpLeft1, core.JUMP_TRANSITION_FRAMES)
		}
	}
}

func ActualJump(p *core.Player) {
	// apply the jump force to the player's physics component
	p.Physics.Forces[core.FORCE_TYPE_JUMP] = &core.Vector2{X: 0, Y: 0.5}
}

func EndJump(p *core.Player) {
	// if you were transitioning to jump3, transition to either walking or standing based on if the movement keys are being held down
	if p.MovingLeft {
		core.ChangeWalkState(p, core.ANIMAtION_TYPE_WALKING_LEFT, stances.WalkLeft1, core.JUMP_TRANSITION_FRAMES)
	} else if p.MovingRight {
		core.ChangeWalkState(p, core.ANIMATION_TYPE_WALKING_RIGHT, stances.WalkRight1, core.JUMP_TRANSITION_FRAMES)
	} else if p.WalkingStanceTo.Direction == core.DIRECTION_RIGHT {
		core.ChangeWalkState(p, core.ANIMATION_TYPE_STANDING, stances.RestRight1, core.JUMP_TRANSITION_FRAMES)
	} else if p.WalkingStanceTo.Direction == core.DIRECTION_LEFT {
		core.ChangeWalkState(p, core.ANIMATION_TYPE_STANDING, stances.RestLeft1, core.JUMP_TRANSITION_FRAMES)
	}
}

func UpdateStance(p *core.Player) {
	// shift the players stance one step towards what it's transitioning to
	p.Stance = core.ShiftStance(p.WalkingStanceFrom, p.WalkingStanceTo, p.WalkingAnimationFrame, p.WalkingAnimationFrames)
}

func PositionRightFoot(player *core.Player, x float64, y float64) {
	// use inverse kinematics to position the player's right foot at (x, y)
	difx, dify := torso_rotation_diff(core.TORSO_WIDTH-core.UPPER_LEG_WIDTH/2, *player)
	dify2, difx2 := torso_rotation_diff(core.TORSO_HEIGHT, *player)
	pelvis_x := player.Physics.Position.X + difx + difx2
	pelvis_y := player.Physics.Position.Y + player.Physics.Height + dify - dify2
	player.Stance.RightUpperLeg, player.Stance.RightLowerLeg = core.IK(core.UPPER_LEG_HEIGHT, core.LOWER_LEG_HEIGHT, pelvis_x, pelvis_y, x, y, player.Stance.Direction == core.DIRECTION_LEFT)
}

func PositionLeftFoot(player *core.Player, x float64, y float64) {
	// use inverse kinematics to position the player's left foot at (x, y)
	dify, difx := torso_rotation_diff(core.TORSO_HEIGHT, *player)
	pelvis_x := player.Physics.Position.X + difx
	pelvis_y := player.Physics.Position.Y + player.Physics.Height - dify
	player.Stance.LeftUpperLeg, player.Stance.LeftLowerLeg = core.IK(core.UPPER_LEG_HEIGHT, core.LOWER_LEG_HEIGHT, pelvis_x, pelvis_y, x, y, player.Stance.Direction == core.DIRECTION_LEFT)
}

func PositionRightHand(player *core.Player, x float64, y float64) {
	// use inverse kinematics to position the player's right hand at (x, y)
	difx, dify := torso_rotation_diff(core.TORSO_WIDTH, *player)
	shoulder_x := player.Physics.Position.X + difx
	shoulder_y := player.Physics.Position.Y + player.Physics.Height + dify
	player.Stance.RightUpperArm, player.Stance.RightLowerArm = core.IK(core.UPPER_ARM_HEIGHT, core.LOWER_ARM_HEIGHT, shoulder_x, shoulder_y, x, y, player.Stance.Direction == core.DIRECTION_RIGHT)
}

func PositionLeftHand(player *core.Player, x float64, y float64) {
	// use inverse kinematics to position the player's left hand at (x, y)
	shoulder_x := player.Physics.Position.X
	shoulder_y := player.Physics.Position.Y + player.Physics.Height
	player.Stance.LeftUpperArm, player.Stance.LeftLowerArm = core.IK(core.UPPER_ARM_HEIGHT, core.LOWER_ARM_HEIGHT, shoulder_x, shoulder_y, x, y, player.Stance.Direction == core.DIRECTION_RIGHT)
}
