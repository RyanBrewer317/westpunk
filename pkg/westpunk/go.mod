module rbrewer.com/westpunk

go 1.16

require (
	github.com/hajimehoshi/ebiten/v2 v2.1.6
	github.com/mattn/go-sqlite3 v1.14.8
	rbrewer.com/core v0.0.0-00010101000000-000000000000
	rbrewer.com/drawplayer v0.0.0-00010101000000-000000000000
	rbrewer.com/stances v0.0.0-00010101000000-000000000000 // indirect
)

replace rbrewer.com/core => ../core

replace rbrewer.com/drawplayer => ../drawplayer

replace rbrewer.com/stances => ../stances
