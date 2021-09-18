package drawplayer

import (
	"image"
	"math"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"rbrewer.com/core"
)

func DrawPlayer(screen *ebiten.Image, x float64, y float64) {
	if core.PlayerStance.Direction == core.Right {
		draw_player_right_arm(screen, x, y)
		draw_player_right_leg(screen, x, y)
		draw_player_torso(screen, x, y)
		draw_player_head(screen, x, y)
		draw_player_left_leg(screen, x, y)
		draw_player_left_arm(screen, x, y)
	} else if core.PlayerStance.Direction == core.Left {
		draw_player_left_arm(screen, x, y)
		draw_player_left_leg(screen, x, y)
		draw_player_torso(screen, x, y)
		draw_player_head(screen, x, y)
		draw_player_right_leg(screen, x, y)
		draw_player_right_arm(screen, x, y)
	}
}

func draw_player_head(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Head+core.PlayerStance.Torso, 2*math.Pi)
	torso_theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	if torso_theta < 0 {
		torso_theta += 2 * math.Pi
	}
	neckx := x/core.PixelYardRatio - (core.HeadWidth-core.TorsoWidth)/2
	necky := (y / core.PixelYardRatio)
	difx, dify := torso_rotation_diff(core.HeadWidth / 2)
	neckx += difx
	necky += dify
	headx := neckx - (core.HeadWidth * math.Cos(theta) / 2) + (core.HeadHeight * math.Sin(theta))
	heady := necky + (core.HeadWidth * math.Sin(theta) / 2) + (core.HeadHeight * math.Cos(theta))
	draw_player_piece(screen, 0, 0, 131, 104, core.PlayerDrawOptions, headx, heady, core.HeadWidth, core.HeadHeight, theta)
}

func torso_rotation_diff(r float64) (float64, float64) {
	theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
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

func draw_player_torso(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	draw_player_piece(screen, 165, 0, 300, 240, core.PlayerDrawOptions, x/core.PixelYardRatio, y/core.PixelYardRatio, core.TorsoWidth, core.TorsoHeight, theta)
}

func draw_player_left_arm(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.LeftUpperArm+core.PlayerStance.Torso, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(0)
	leftshoulderx := x/core.PixelYardRatio + difx
	leftshouldery := y/core.PixelYardRatio + dify
	leftupperarmx := leftshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	leftupperarmy := leftshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, leftupperarmx, leftupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.LeftLowerArm, 2*math.Pi)
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
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, leftforearmx, leftforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2)
}

func draw_player_right_arm(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperArm, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth)
	rightshoulderx := x/core.PixelYardRatio + difx
	rightshouldery := y/core.PixelYardRatio + dify
	rightupperarmx := rightshoulderx - (core.UpperArmWidth * math.Cos(theta) / 2)
	rightupperarmy := rightshouldery + (core.UpperArmWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, rightupperarmx, rightupperarmy, core.UpperArmWidth, core.UpperArmHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperArm+core.PlayerStance.RightLowerArm, 2*math.Pi)
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
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, rightforearmx, rightforearmy, core.LowerArmWidth, core.LowerArmHeight, theta2)
}

func draw_player_left_leg(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.LeftUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.UpperLegWidth / 2)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight)
	pelvis_left_x := x/core.PixelYardRatio + difx + difx2
	pelvis_left_y := y/core.PixelYardRatio + dify - dify2
	leftupperlegx := pelvis_left_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	leftupperlegy := pelvis_left_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, leftupperlegx, leftupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.LeftLowerLeg, 2*math.Pi)
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
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, leftlowerlegx, leftlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2)
}

func draw_player_right_leg(screen *ebiten.Image, x float64, y float64) {
	core.PlayerDrawOptions.GeoM.Reset()
	theta := math.Mod(core.PlayerStance.Torso+core.PlayerStance.RightUpperLeg, 2*math.Pi)
	if theta < 0 {
		theta += 2 * math.Pi
	}
	difx, dify := torso_rotation_diff(core.TorsoWidth - core.UpperLegWidth/2)
	dify2, difx2 := torso_rotation_diff(core.TorsoHeight)
	pelvis_right_x := x/core.PixelYardRatio + difx + difx2
	pelvis_right_y := y/core.PixelYardRatio + dify - dify2
	rightupperlegx := pelvis_right_x - (core.UpperLegWidth * math.Cos(theta) / 2)
	rightupperlegy := pelvis_right_y + (core.UpperLegWidth * math.Sin(theta) / 2)
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, rightupperlegx, rightupperlegy, core.UpperLegWidth, core.UpperLegHeight, theta)

	core.PlayerDrawOptions.GeoM.Reset()
	theta2 := math.Mod(theta+core.PlayerStance.RightLowerLeg, 2*math.Pi)
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
	draw_player_piece(screen, 131, 0, 165, 240, core.PlayerDrawOptions, rightlowerlegx, rightlowerlegy, core.LowerLegWidth, core.LowerLegHeight, theta2)
}

func draw_player_piece(screen *ebiten.Image, imgx1 int, imgy1 int, imgx2 int, imgy2 int, drawoptions ebiten.DrawImageOptions, igx float64, igy float64, igw float64, igh float64, theta float64) {
	img := ebiten.NewImageFromImage(core.PlayerImg.SubImage(image.Rect(imgx1, imgy1, imgx2, imgy2)))
	wi, hi := img.Size()
	w := igw * core.PixelYardRatio
	h := igh * core.PixelYardRatio
	direction_scale := 1.0
	direction_translation := 0.0
	if core.PlayerStance.Direction == core.Left {
		direction_scale = -1
		direction_translation = float64(wi)
	}
	drawoptions.GeoM.Scale(direction_scale, 1)
	drawoptions.GeoM.Translate(direction_translation, 0)
	drawoptions.GeoM.Scale(w/float64(wi), h/float64(hi))
	drawoptions.GeoM.Rotate(theta)
	core.DrawImage(screen, img, drawoptions, igx*core.PixelYardRatio-core.VP.X, core.GetPXY(igy)+core.VP.Y)
}
