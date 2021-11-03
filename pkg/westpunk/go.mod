module ryanbrewer.page/westpunk

go 1.16

require (
	github.com/hajimehoshi/ebiten/v2 v2.1.7
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	modernc.org/sqlite v1.13.3
	ryanbrewer.page/audio v0.0.0-00010101000000-000000000000
	ryanbrewer.page/core v0.0.0-00010101000000-000000000000
	ryanbrewer.page/physics v0.0.0-00010101000000-000000000000
	ryanbrewer.page/player v0.0.0-00010101000000-000000000000
	ryanbrewer.page/stances v0.0.0-00010101000000-000000000000
)

replace ryanbrewer.page/core => ../core

replace ryanbrewer.page/physics => ../physics

replace ryanbrewer.page/player => ../player

replace ryanbrewer.page/stances => ../stances

replace ryanbrewer.page/audio => ../audio
