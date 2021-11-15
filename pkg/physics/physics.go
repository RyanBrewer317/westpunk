package physics

import (
	"fmt"

	"ryanbrewer.page/core"
)

func CanMoveRight(p *core.PhysicsComponent) bool {
	// detect obstructions on the right side of the player
	right_border_x := core.PLACE_WIDTH - (0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO)
	potentially_relevant_things := append(core.Grid[core.Coordinate{X: int(p.Position.X), Y: int(p.Position.Y)}], core.Grid[core.Coordinate{X: int(p.Position.X) + 1, Y: int(p.Position.Y)}]...)
	for i := 0; i < len(potentially_relevant_things); i++ {
		thing := potentially_relevant_things[i]
		ot := core.ObstructionTable[thing.Type]
		// if the thing is obstructive and in a position to obstruct
		if (ot == core.OBSTRUCTION_TYPE_RIGHT_SLANT_45 || ot == core.OBSTRUCTION_TYPE_OBSTRUCTIVE) &&
			thing.Physics.Position.X > (p.Position.X+p.Width) &&
			thing.Physics.Position.X-(p.Position.X+p.Width) < 0.1 &&
			(thing.Physics.Position.Y >= p.Position.Y-p.Height && p.Position.Y >= thing.Physics.Position.Y-thing.Physics.Height) {
			return false
		}
	}
	// if no thing obstructed the right, return whether the component is left of the right place border
	return p.Position.X+p.Width < right_border_x
}

func MoveRight(p *core.PhysicsComponent) {
	// move the player right
	p.Motion.Add(core.Vector2{X: core.PLAYER_WALK_SPEED, Y: 0})
}

func CanMoveLeft(p *core.PhysicsComponent) bool {
	// detect obstructions on the left side of the player
	potentially_relevant_things := append(core.Grid[core.Coordinate{X: int(p.Position.X), Y: int(p.Position.Y)}], core.Grid[core.Coordinate{X: int(p.Position.X) - 1, Y: int(p.Position.Y)}]...)
	left_border_x := 0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO
	for i := 0; i < len(potentially_relevant_things); i++ {
		thing := potentially_relevant_things[i]
		ot := core.ObstructionTable[thing.Type]
		// if the thing is obstructive and in a position to obstruct
		if (ot == core.OBSTRUCTION_TYPE_LEFT_SLANT_45 || ot == core.OBSTRUCTION_TYPE_OBSTRUCTIVE) &&
			thing.Physics.Position.X+thing.Physics.Width > p.Position.X &&
			p.Position.X-(thing.Physics.Position.X+thing.Physics.Width) < 0.1 &&
			(thing.Physics.Position.Y >= p.Position.Y-p.Height && p.Position.Y >= thing.Physics.Position.Y-thing.Physics.Height) {
			return false
		}
	}
	// if no thing obstructed the left, return whether the component is right of the left world border
	return p.Position.X > left_border_x
}

func MoveLeft(p *core.PhysicsComponent) {
	// apply a leftward force at the player walk velocity
	p.Motion.Add(core.Vector2{X: -core.PLAYER_WALK_SPEED, Y: 0})
}

func ConfineToPlace(p *core.PhysicsComponent) {
	// stop the player from leaving the place. In the future, leaving the place will bring you to the place map
	right_edge := core.PLACE_WIDTH - (0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO)
	if p.Position.X+p.Width > right_edge {
		p.Position.X = right_edge - p.Width
	}
	left_edge := 0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO
	if p.Position.X < left_edge {
		p.Position.X = left_edge
	}
	// TODO: this grounded logic should probably done elsewhere or else this function should be renamed
	var ground_y float64
	p.Grounded, ground_y = Grounded(*p)
	if p.Grounded {
		p.Position.Y = ground_y
	} else {
		fmt.Print(0)
	}
}

func Move(p *core.PhysicsComponent) {
	// apply all the existing forces onto the component's position, then do whatever cleanup is necessary
	if force, ok := p.Forces[core.FORCE_TYPE_GRAVITY]; ok {
		if !p.Grounded {
			force.Add(core.Vector2{X: 0, Y: -0.03})
		} else {
			force.Scale(0)
		}
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	if force, ok := p.Forces[core.FORCE_TYPE_JUMP]; ok {
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	if force, ok := p.Forces[core.FORCE_TYPE_KNOCKBACK]; ok {
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	p.Position.Add(p.Motion)
	p.Motion.Scale(0)
	ConfineToPlace(p)
}

func CollisionDetected(component1 core.PhysicsComponent, component2 core.PhysicsComponent) bool {
	// currently we just use a basic box detection
	return (component1.Position.X <= component2.Position.X+component2.Width &&
		component1.Position.X+component1.Width >= component2.Position.X &&
		component1.Position.Y >= component2.Position.Y-component2.Height &&
		component1.Position.Y-component1.Height <= component2.Position.Y)
}

func Grounded(p core.PhysicsComponent) (bool, float64) {
	// calculate whether or not the component is grounded. Also return the y value of the ground, to correct the Y value should it be slightly too low
	// this feels optimizable

	// the area around the feet
	chunklets := append(core.Grid[core.Coordinate{X: int(p.Position.X), Y: int(p.Position.Y)}], core.Grid[core.Coordinate{X: int(p.Position.X), Y: int(p.Position.Y) - 1}]...)
	chunklets = append(chunklets, core.Grid[core.Coordinate{X: int(p.Position.X) + 1, Y: int(p.Position.Y)}]...)
	chunklets = append(chunklets, core.Grid[core.Coordinate{X: int(p.Position.Y) + 1, Y: int(p.Position.Y) - 1}]...)
	if len(chunklets) == 0 { // there's nothing around that area so just return whether or not the player is standing on the ground
		return p.Position.Y <= core.GROUND_Y+0.01, core.GROUND_Y
	}
	// to find the y value of the ground at x, get the largest y value of everything below x
	// the y value of each thing is calculated as thing_top_at_x
	top_height_under_x := -1.0
	for i := 0; i < len(chunklets); i++ {
		thing := chunklets[i]
		// if the thing is unobstructive or isn't below the physics component, ignore it
		if core.ObstructionTable[thing.Type] == core.OBSTRUCTION_TYPE_UNOBSTRUCTIVE || (p.Position.X+p.Width < thing.Physics.Position.X || thing.Physics.Position.X+thing.Physics.Width < p.Position.X) {
			continue
		}
		// start with the assumption that it's just the basic type of obstructive
		thing_top_at_x := thing.Physics.Position.Y + thing.Physics.Height
		// if it's a slant, recalculate the top at x to reflect that
		if core.ObstructionTable[thing.Type] == core.OBSTRUCTION_TYPE_LEFT_SLANT_45 {
			thing_top_at_x = thing.Physics.Position.Y + (p.Position.X - thing.Physics.Position.X)
		}
		if core.ObstructionTable[thing.Type] == core.OBSTRUCTION_TYPE_RIGHT_SLANT_45 {
			thing_top_at_x = thing.Physics.Position.Y + (thing.Physics.Position.X - (p.Position.X + p.Width))
		}
		// if the top at x is higher than any found so far, set it to top_height_under_x
		if thing_top_at_x > float64(top_height_under_x) {
			top_height_under_x = thing_top_at_x
		}
	}
	// if top_height_under_x is still negative, it means everything in the chunk was unobstructive or not under the component
	// in which case we return whether or not the component is on the ground
	if top_height_under_x < 0 {
		return p.Position.Y <= core.GROUND_Y+0.01, core.GROUND_Y
	}
	// finally, return whether the component is at least at the top height
	return top_height_under_x >= p.Position.Y, top_height_under_x
}
