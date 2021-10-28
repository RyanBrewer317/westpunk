module rbrewer.com/westpunk

go 1.16

require (
	github.com/hajimehoshi/ebiten/v2 v2.1.7
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	modernc.org/sqlite v1.13.3
	rbrewer.com/core v0.0.0-00010101000000-000000000000
	rbrewer.com/physics v0.0.0-00010101000000-000000000000
	rbrewer.com/player v0.0.0-00010101000000-000000000000
	rbrewer.com/stances v0.0.0-00010101000000-000000000000
)

replace rbrewer.com/core => ../core

replace rbrewer.com/physics => ../physics

replace rbrewer.com/player => ../player

replace rbrewer.com/stances => ../stances
