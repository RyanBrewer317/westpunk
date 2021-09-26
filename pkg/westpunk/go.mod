module rbrewer.com/westpunk

go 1.16

require (
	github.com/hajimehoshi/ebiten/v2 v2.1.7
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	golang.org/x/sys v0.0.0-20210925032602-92d5a993a665 // indirect
	rbrewer.com/core v0.0.0-00010101000000-000000000000
	rbrewer.com/drawplayer v0.0.0-00010101000000-000000000000
	rbrewer.com/stances v0.0.0-00010101000000-000000000000
)

replace rbrewer.com/core => ../core

replace rbrewer.com/drawplayer => ../drawplayer

replace rbrewer.com/stances => ../stances
