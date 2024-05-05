package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	pos      rl.Vector2
	size     rl.Vector2
	velocity rl.Vector2
	speed    float32
}

type Enemy struct {
	pos   rl.Vector2
	size  rl.Vector2
	speed float32
}

func main() {
	var screenWidth int32 = 640
	var screenHeight int32 = 450
	const title = "slayfast"

	rl.InitWindow(screenWidth, screenHeight, title)
	rl.SetWindowState(rl.FlagWindowResizable)
	defer rl.CloseWindow()
	rl.SetTargetFPS(0)

	player := Player{
		pos:      rl.Vector2{X: 0, Y: 0},
		size:     rl.Vector2{X: 32, Y: 32},
		velocity: rl.Vector2{X: 0, Y: 0},
		speed:    16,
	}
	playerInterpolatedPos := player.pos

	var camera = rl.Camera2D{
		Offset:   rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2},
		Target:   playerInterpolatedPos,
		Rotation: 0,
		Zoom:     1,
	}

	var tickRate uint = 16
	onTick := newTickLoop(tickRate)

	enemies := []Enemy{}
	enemySpawnInterval := 2 // every 4s
	enemySpawnTicksInterval := enemySpawnInterval * int(tickRate)
	enemySpawnTimer := 0

	for !rl.WindowShouldClose() {
		rl.DrawFPS(0, 0)

		if rl.IsWindowResized() {
			screenWidth = int32(rl.GetScreenWidth())
			screenHeight = int32(rl.GetScreenHeight())
			camera.Offset = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
		}

		// update
		readPlayerMovment(&player)
		playerInterpolatedPos = interpolatePos(playerInterpolatedPos, player.pos, 1/float32(tickRate), rl.GetFrameTime())
		camera.Target = playerInterpolatedPos
		if wheel := rl.GetMouseWheelMove(); wheel != 0 {
			updateCameraZoom(&camera, wheel, 0.05)
		}

		onTick(func() {
			enemySpawnTimer += 1
			if enemySpawnTimer >= enemySpawnTicksInterval {
				spawnEnemy(rl.Vector2{X: 0, Y: 0}, &enemies)
				spawnEnemy(rl.Vector2{X: 0, Y: 0}, &enemies)
				enemySpawnTimer = 0
			}
			movePlayer(&player)
			moveEnemies(&enemies, player)
		})

		// render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(camera)

		rl.PushMatrix()
		rl.Translatef(0, 16*100, 0)
		rl.Rotatef(90, 1, 0, 0)
		rl.DrawGrid(100, 100)
		rl.PopMatrix()

		rl.DrawRectangleRec(rl.Rectangle{
			X:      playerInterpolatedPos.X - player.size.X/2,
			Y:      playerInterpolatedPos.Y - player.size.Y/2,
			Width:  player.size.X,
			Height: player.size.Y,
		}, rl.White)

		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      player.pos.X - player.size.X/2,
			Y:      player.pos.Y - player.size.Y/2,
			Width:  player.size.X,
			Height: player.size.Y,
		}, 1, rl.Gray)

		for _, enemy := range enemies {
			rl.DrawRectangleLinesEx(rl.Rectangle{
				X:      enemy.pos.X - enemy.size.X/2,
				Y:      enemy.pos.Y - enemy.size.Y/2,
				Width:  enemy.size.X,
				Height: enemy.size.Y,
			}, 5, rl.Red)
		}

		rl.EndMode2D()
		rl.EndDrawing()
	}
}

func movePlayer(p *Player) {
	p.pos = rl.Vector2Add(p.pos, p.velocity)
	p.velocity = rl.Vector2{}
}

func readPlayerMovment(p *Player) {
	if rl.IsKeyDown('D') {
		p.velocity.Y = -1
	} else if rl.IsKeyDown('T') {
		p.velocity.Y = 1
	}
	if rl.IsKeyDown('R') {
		p.velocity.X = -1
	} else if rl.IsKeyDown('S') {
		p.velocity.X = 1
	}

	p.velocity = rl.Vector2Normalize(p.velocity)
	p.velocity.X *= p.speed
	p.velocity.Y *= p.speed
}

func updateCameraZoom(c *rl.Camera2D, wheel float32, step float32) {
	c.Zoom += wheel * step
	if c.Zoom < 0.5 {
		c.Zoom = 0.5
	}
	if c.Zoom > 2 {
		c.Zoom = 2
	}
}

func newTickLoop(tickRate uint) func(func()) {
	var tickTime float32 = 1 / float32(tickRate)
	var tickTimer float32 = 0

	return func(f func()) {
		tickTimer += rl.GetFrameTime()
		for tickTimer >= tickTime {
			f()
			tickTimer -= tickTime
		}
	}
}

func interpolatePos(framePos, pos rl.Vector2, tickTime, frameTime float32) rl.Vector2 {
	return rl.Vector2{
		X: framePos.X + (pos.X-framePos.X)/(tickTime/frameTime),
		Y: framePos.Y + (pos.Y-framePos.Y)/(tickTime/frameTime),
	}
}

func spawnEnemy(pos rl.Vector2, enemies *[]Enemy) {
	*enemies = append(*enemies, Enemy{
		pos:   pos,
		size:  rl.Vector2{X: 64, Y: 64},
		speed: 12,
	})
}

func moveEnemies(enemies *[]Enemy, player Player) {
	for i, e := range *enemies {
		velocity := rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(player.pos, e.pos)), e.speed)
		(*enemies)[i].pos = rl.Vector2Add(e.pos, velocity)
	}
}
