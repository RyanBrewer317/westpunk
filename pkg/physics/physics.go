package physics

import (
	"rbrewer.com/core"
)

func CanMoveRight(p *core.PhysicsComponent) bool {
	right_border_x := core.PLACE_WIDTH - (0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO)
	return p.Position.X < right_border_x
}

func MoveRight(p *core.PhysicsComponent) {
	p.Motion.Add(core.Vector2{X: 0.09, Y: 0})
}

func CanMoveLeft(p *core.PhysicsComponent) bool {
	left_border_x := 0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO
	return p.Position.X > left_border_x
}

func MoveLeft(p *core.PhysicsComponent) {
	p.Motion.Add(core.Vector2{X: -0.09, Y: 0})
}

func ConfineToPlace(p *core.PhysicsComponent) {
	right_edge := core.PLACE_WIDTH - (0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO)
	if p.Position.X+p.Width > right_edge {
		p.Position.X = right_edge
	}
	left_edge := 0.5 * core.SCREEN_WIDTH / core.PIXEL_YARD_RATIO
	if p.Position.X < left_edge {
		p.Position.X = left_edge
	}
	if p.Position.Y-p.Height <= core.GROUND_Y+0.05 {
		p.Position.Y = p.Height
		p.Grounded = true
	} else {
		p.Grounded = false
	}
}

func Move(p *core.PhysicsComponent) {
	if force, ok := p.Forces[core.GRAVITY]; ok {
		if !p.Grounded {
			force.Add(core.Vector2{X: 0, Y: -0.03})
		} else {
			force.Scale(0)
		}
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	if force, ok := p.Forces[core.JUMP_FORCE]; ok {
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	if force, ok := p.Forces[core.KNOCKBACK]; ok {
		force.Scale(0.8)
		p.Motion.Add(*force)
	}
	p.Position.Add(p.Motion)
	p.Motion.Scale(0)
	ConfineToPlace(p)
}

func CollisionDetected(component1 core.PhysicsComponent, component2 core.PhysicsComponent) bool {
	return (component1.Position.X <= component2.Position.X+component2.Width &&
		component1.Position.X+component1.Width >= component2.Position.X &&
		component1.Position.Y >= component2.Position.Y-component2.Height &&
		component1.Position.Y-component1.Height <= component2.Position.Y)
}
