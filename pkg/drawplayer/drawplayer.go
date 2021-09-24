package drawplayer

import (
	"image"
	"math"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"rbrewer.com/core"
)

func DrawPlayer(screen *ebiten.Image, player core.Player, x float64, y float64) {
	if player.Stance.Direction == core.Right {
		draw_player_right_arm(screen, player, x, y)
		draw_player_right_leg(screen, player, x, y)
		draw_player_torso(screen, player, x, y)
		draw_player_head(screen, player, x, y)
		draw_player_left_leg(screen, player, x, y)
		draw_player_left_arm(screen, player, x, y)
	} else if player.Stance.Direction == core.Left {
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
	theta := math.Mod(player.Stance.Head+player.Stance.Torso, 2*math.Pi)
	torso_theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	if torso_theta < 0 {
		torso_theta += 2 * math.Pi
	}
	neckx := x/core.PixelYardRatio - (core.HeadWidth-core.TorsoWidth)/2
	necky := (y / core.PixelYardRatio)
	difx, dify := torso_rotation_diff(core.HeadWidth/2, player)
	neckx += difx
	necky += dify
	headx := neckx - (core.HeadWidth * math.Cos(theta) / 2) + (core.HeadHeight * math.Sin(theta))
	heady := necky + (core.HeadWidth * math.Sin(theta) / 2) + (core.HeadHeight * math.Cos(theta))
	draw_player_piece(screen, 0, 0, 131, 104, player.DrawOptions, headx, heady, core.HeadWidth, core.HeadHeight, theta, player.Stance.Direction)
}

func torso_rotation_diff(r float64, player core.Player) (float64, float64) {
	theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	x := 0.0
	y := 0.0
	if theta < math.Pi/2 {
		x += r * math.Cos(theta)
		y -= r * math.Sin(theta)
	} else if theta < math.Pi {
		x -= r * math.Sin(theta-math.Pi/2)
		y -= r * math.Cos(theta-math.Pi/2)
	} else if theta < 3*math.Pi/2 {
		x -= r * math.Cos(theta-math.Pi)
		y += r * math.Sin(theta-math.Pi)
	} else {
		x += r * math.Sin(theta-3*math.Pi/2)
		y += r * math.Cos(theta-3*math.Pi/2)
	}
	return x, y
}

func draw_player_torso(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	draw_player_piece(screen, 165, 0, 300, 240, player.DrawOptions, x/core.PixelYardRatio, y/core.PixelYardRatio, core.TorsoWidth, core.TorsoHeight, theta, player.Stance.Direction)
}

func draw_player_left_arm(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.LeftUpperArm+player.Stance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(0, player)
	leftshoulderx := x/core.PixelYardRatio + difx
	leftshouldery := y/core.PixelYardRatio + dify
	leftupperarmx := leftshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	leftupperarmy := leftshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftupperarmx, leftupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta, player.Stance.Direction)

	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.LeftLowerArm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftelbowx := leftshoulderx - (core.UpperArmHeight * math.Sin(theta))
	leftelbowy := leftshouldery - (core.UpperArmHeight * math.Cos(theta))
	var leftforearmx float64
	var leftforearmy float64
	if 0 <= theta2 && theta2 < math.Pi/2 {
		leftforearmx = leftelbowx - (core.LowerArmWidth/2)*math.Cos(theta2)
		leftforearmy = leftelbowy + (core.LowerArmWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 <= theta2 && theta2 < math.Pi {
		leftforearmx = leftelbowx + (core.LowerArmWidth/2)*math.Cos(math.Pi-theta2)
		leftforearmy = leftelbowy + (core.LowerArmWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi <= theta2 && theta2 < 3*math.Pi/2 {
		leftforearmx = leftelbowx + (core.LowerArmWidth/2)*math.Sin(3*math.Pi/2-theta2)
		leftforearmy = leftelbowy - (core.LowerArmWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 <= theta2 && theta2 < 2*math.Pi {
		leftforearmx = leftelbowx - (core.LowerArmWidth/2)*math.Sin(theta2-3*math.Pi/2)
		leftforearmy = leftelbowy - (core.LowerArmWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftforearmx, leftforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2, player.Stance.Direction)
}

func draw_player_right_arm(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.Torso+player.Stance.RightUpperArm, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth, player)
	rightshoulderx := x/core.PixelYardRatio + difx
	rightshouldery := y/core.PixelYardRatio + dify
	rightupperarmx := rightshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	rightupperarmy := rightshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightupperarmx, rightupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta, player.Stance.Direction)

	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(player.Stance.Torso+player.Stance.RightUpperArm+player.Stance.RightLowerArm, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightelbowx := rightshoulderx - (core.UpperArmHeight * math.Sin(theta))
	rightelbowy := rightshouldery - (core.UpperArmHeight * math.Cos(theta))
	var rightforearmx float64
	var rightforearmy float64
	if 0 <= theta2 && theta2 < math.Pi/2 {
		rightforearmx = rightelbowx - (core.LowerArmWidth/2)*math.Cos(theta2)
		rightforearmy = rightelbowy + (core.LowerArmWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 <= theta2 && theta2 < math.Pi {
		rightforearmx = rightelbowx + (core.LowerArmWidth/2)*math.Cos(math.Pi-theta2)
		rightforearmy = rightelbowy + (core.LowerArmWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi <= theta2 && theta2 < 3*math.Pi/2 {
		rightforearmx = rightelbowx + (core.LowerArmWidth/2)*math.Sin(3*math.Pi/2-theta2)
		rightforearmy = rightelbowy - (core.LowerArmWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 <= theta2 && theta2 < 2*math.Pi {
		rightforearmx = rightelbowx - (core.LowerArmWidth/2)*math.Sin(theta2-3*math.Pi/2)
		rightforearmy = rightelbowy - (core.LowerArmWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightforearmx, rightforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2, player.Stance.Direction)
}

func draw_player_left_leg(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.Torso+player.Stance.LeftUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.UpperLegWidth/2, player)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight, player)
	pelvis_left_x := x/core.PixelYardRatio + difx + difx2
	pelvis_left_y := y/core.PixelYardRatio + dify - dify2
	leftupperlegx := pelvis_left_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	leftupperlegy := pelvis_left_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftupperlegx, leftupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta, player.Stance.Direction)

	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.LeftLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	leftkneex := pelvis_left_x - (core.UpperLegHeight * math.Sin(theta))
	leftkneey := pelvis_left_y - (core.UpperLegHeight * math.Cos(theta))
	var leftlowerlegx float64
	var leftlowerlegy float64
	if 0 <= theta2 && theta2 < math.Pi/2 {
		leftlowerlegx = leftkneex - (core.LowerLegWidth/2)*math.Cos(theta2)
		leftlowerlegy = leftkneey + (core.LowerLegWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 <= theta2 && theta2 < math.Pi {
		leftlowerlegx = leftkneex + (core.LowerLegWidth/2)*math.Cos(math.Pi-theta2)
		leftlowerlegy = leftkneey + (core.LowerLegWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi <= theta2 && theta2 < 3*math.Pi/2 {
		leftlowerlegx = leftkneex + (core.LowerLegWidth/2)*math.Sin(3*math.Pi/2-theta2)
		leftlowerlegy = leftkneey - (core.LowerLegWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 <= theta2 && theta2 < 2*math.Pi {
		leftlowerlegx = leftkneex - (core.LowerLegWidth/2)*math.Sin(theta2-3*math.Pi/2)
		leftlowerlegy = leftkneey - (core.LowerLegWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, leftlowerlegx, leftlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2, player.Stance.Direction)
}

func draw_player_right_leg(screen *ebiten.Image, player core.Player, x float64, y float64) {
	player.DrawOptions.GeoM.Reset()
	theta := math.Mod(player.Stance.Torso+player.Stance.RightUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth-core.UpperLegWidth/2, player)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight, player)
	pelvis_right_x := x/core.PixelYardRatio + difx + difx2
	pelvis_right_y := y/core.PixelYardRatio + dify - dify2
	rightupperlegx := pelvis_right_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	rightupperlegy := pelvis_right_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightupperlegx, rightupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta, player.Stance.Direction)

	player.DrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+player.Stance.RightLowerLeg, 2*math.Pi)
	if theta2 < 0 {
		theta2 += 2 * math.Pi
	}
	rightkneex := pelvis_right_x - (core.UpperLegHeight * math.Sin(theta))
	rightkneey := pelvis_right_y - (core.UpperLegHeight * math.Cos(theta))
	var rightlowerlegx float64
	var rightlowerlegy float64
	if 0 <= theta2 && theta2 < math.Pi/2 {
		rightlowerlegx = rightkneex - (core.LowerLegWidth/2)*math.Cos(theta2)
		rightlowerlegy = rightkneey + (core.LowerLegWidth/2)*math.Sin(theta2)
	}
	if math.Pi/2 <= theta2 && theta2 < math.Pi {
		rightlowerlegx = rightkneex + (core.LowerLegWidth/2)*math.Cos(math.Pi-theta2)
		rightlowerlegy = rightkneey + (core.LowerLegWidth/2)*math.Sin(math.Pi-theta2)
	}
	if math.Pi <= theta2 && theta2 < 3*math.Pi/2 {
		rightlowerlegx = rightkneex + (core.LowerLegWidth/2)*math.Sin(3*math.Pi/2-theta2)
		rightlowerlegy = rightkneey - (core.LowerLegWidth/2)*math.Cos(3*math.Pi/2-theta2)
	}
	if 3*math.Pi/2 <= theta2 && theta2 < 2*math.Pi {
		rightlowerlegx = rightkneex - (core.LowerLegWidth/2)*math.Sin(theta2-3*math.Pi/2)
		rightlowerlegy = rightkneey - (core.LowerLegWidth/2)*math.Cos(theta2-3*math.Pi/2)
	}
	draw_player_piece(screen, 131, 0, 165, 240, player.DrawOptions, rightlowerlegx, rightlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2, player.Stance.Direction)
}

func draw_player_piece(screen *ebiten.Image, imgx1 int, imgy1 int, imgx2 int, imgy2 int, drawoptions ebiten.DrawImageOptions, igx float64, igy float64, igw float64, igh float64, theta float64, direction core.Direction) {
	img := ebiten.NewImageFromImage(core.PlayerImg.SubImage(image.Rect(imgx1, imgy1, imgx2, imgy2)))
	wi, hi := img.Size()
	w := igw * core.PixelYardRatio
	h := igh * core.PixelYardRatio
	direction_scale := 1.0
	direction_translation := 0.0
	if direction == core.Left {
		direction_scale = -1
		direction_translation = float64(wi)
	}
	drawoptions.GeoM.Scale(direction_scale, 1)
	drawoptions.GeoM.Translate(direction_translation, 0)
	drawoptions.GeoM.Scale(w/float64(wi), h/float64(hi))
	drawoptions.GeoM.Rotate(theta)
	core.DrawImage(screen, img, drawoptions, igx*core.PixelYardRatio-core.VP.X, core.GetPXY(igy)+core.VP.Y)
}
