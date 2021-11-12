package settings_ui

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"ryanbrewer.page/core"
)

type checkbox struct {
	Checked     *bool
	DrawOptions *ebiten.DrawImageOptions
	Physics     core.PhysicsComponent
}

func (c *checkbox) Draw(screen *ebiten.Image) {
	c.DrawOptions.GeoM.Reset()
	if *c.Checked {
		core.ResizeImage(checkedImage, c.DrawOptions, float64(checkboxImageWidth), float64(checkboxImageHeight))
		core.DrawImage(screen, checkedImage, *c.DrawOptions, c.Physics.Position.X, c.Physics.Position.Y)
	} else {
		core.ResizeImage(uncheckedImage, c.DrawOptions, float64(checkboxImageWidth), float64(checkboxImageHeight))
		core.DrawImage(screen, uncheckedImage, *c.DrawOptions, c.Physics.Position.X, c.Physics.Position.Y)
	}
}

var checkedImage *ebiten.Image
var uncheckedImage *ebiten.Image
var checkboxImageWidth int = 15
var checkboxImageHeight int = 15

var MuteSFX checkbox = checkbox{
	DrawOptions: &ebiten.DrawImageOptions{},
	Physics: core.PhysicsComponent{
		Width:  float64(checkboxImageWidth),
		Height: float64(checkboxImageHeight),
	},
	Checked: &core.SETTINGS_SFX_MUTED,
}

func PrepareSettingsUIImages() {
	var err error
	checkedImage, _, err = ebitenutil.NewImageFromFile("checked.png")
	uncheckedImage, _, err = ebitenutil.NewImageFromFile("unchecked.png")
	if err != nil {
		log.Fatal(err)
	}
}
