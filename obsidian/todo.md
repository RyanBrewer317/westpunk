
 - ~~rewrite sqlite integration~~
 - ~~physics forced-collider system~~
	 - ~~collision detection~~

_At this point the goal is to have presentable code for a video, that I can add features too very easily. Also it must be written so that foundational changes that are listed in the icebox won't require a huge rewrite if I wait, say, a year before implementing them_

 - stone tool attacks
 - oak wood gathering
 - thing placing
 - oak wood ramp crafting
 - wooden axe crafting (even though there's no wooden sword in the final game)
 - wooden axe attacks and using
 - mining stones and metals

**ICEBOX**
 - improve sqlite integration
 - do some hecking wild sorcery to get rid of a reliance on cgo, which ebiten uses. In an ideal world ebiten just fixes that themselves lol
 - uneven ground ik
 - dragging-foot walking algorithm, with a stablized-stance before an attack/action if unstable