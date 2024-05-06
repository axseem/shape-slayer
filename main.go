package main

import (
	"math/rand/v2"
	"slices"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	pos       rl.Vector2
	velocity  rl.Vector2
	size      float32
	speed     float32
	health    int
	maxHealth int
	damage    int
	direction rl.Vector2
	atackSize float32
}

type Enemy struct {
	pos      rl.Vector2
	velocity rl.Vector2
	size     float32
	speed    float32
	health   int
	damage   int
}

func main() {
	var screenWidth int32 = 640
	var screenHeight int32 = 450
	const title = "Shape Slayer"
	isGamePaused := false

	rl.InitWindow(screenWidth, screenHeight, title)
	rl.SetWindowState(rl.FlagWindowResizable)
	defer rl.CloseWindow()
	rl.SetTargetFPS(0)

	player := Player{
		pos:       rl.Vector2{},
		velocity:  rl.Vector2{},
		size:      32,
		speed:     32,
		health:    100,
		maxHealth: 100,
		damage:    10,
		atackSize: 64,
	}

	var camera = rl.Camera2D{
		Offset:   rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2},
		Target:   player.pos,
		Rotation: 0,
		Zoom:     1,
	}

	defauldEnemy := Enemy{
		size:   32,
		speed:  24,
		health: 20,
		damage: 1,
	}

	var tickRate uint = 32
	onTick := newTickLoop(tickRate)

	enemies := []Enemy{}
	enemiesPerSpawn := 1
	var enemySpawnSecondsInterval float32 = 4
	enemySpawnTicksInterval := int(enemySpawnSecondsInterval * float32(tickRate))
	enemySpawnTimer := 0

	for !rl.WindowShouldClose() {

		if rl.IsWindowResized() {
			screenWidth = int32(rl.GetScreenWidth())
			screenHeight = int32(rl.GetScreenHeight())
			camera.Offset = rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}
		}

		// update
		if !isGamePaused {

			player.direction = rl.Vector2Normalize(rl.Vector2Subtract(rl.GetMousePosition(), rl.Vector2{X: float32(screenWidth) / 2, Y: float32(screenHeight) / 2}))

			readPlayerMovment(&player)
			movePlayer(&player, rl.GetFrameTime())
			camera.Target = player.pos
			if wheel := rl.GetMouseWheelMove(); wheel != 0 {
				updateCameraZoom(&camera, wheel, 0.05)
			}

			if rl.IsMouseButtonPressed(0) {
				attack(player, &enemies)
			}

			onTick(func() {
				enemySpawnTimer += 1
				if enemySpawnTimer >= enemySpawnTicksInterval {
					for range enemiesPerSpawn {
						spawnEnemy(randPosOnEdge(player.pos, float32(screenWidth)*2, float32(screenHeight)*2, 2), defauldEnemy, &enemies)
					}
					enemySpawnTimer = 0
					enemiesPerSpawn += 1
					if enemySpawnSecondsInterval > 1 {
						enemySpawnSecondsInterval *= 0.95
						enemySpawnTicksInterval = int(enemySpawnSecondsInterval * float32(tickRate))
					}
				}

				checkPlayerEnemiesCollision(&player, &enemies)
				if player.health <= 0 {
					rl.GlClose()
					rl.CloseWindow()
				}

				enemies = slices.DeleteFunc(enemies, func(e Enemy) bool {
					if e.health <= 0 {
						return true
					}
					return false
				})
				determineEnemiesPos(&enemies, player)
			})
			moveEnemies(&enemies, rl.GetFrameTime())
		}

		if rl.IsKeyPressed('P') {
			isGamePaused = !isGamePaused
		}

		// render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(camera)

		rl.PushMatrix()
		rl.Translatef(0, 16*100, 0)
		rl.Rotatef(90, 1, 0, 0)
		rl.DrawGrid(100, 100)
		rl.PopMatrix()

		if rl.IsMouseButtonPressed(0) {
			rl.DrawCircle(int32(player.pos.X+(player.direction.X*(player.size+player.atackSize/2))), int32(player.pos.Y+(player.direction.Y*(player.size+player.atackSize/2))), player.atackSize, rl.DarkGray)
		}
		rl.DrawCircle(int32(player.pos.X), int32(player.pos.Y), player.size, rl.White)
		rl.DrawCircle(int32(player.pos.X+player.direction.X*10), int32(player.pos.Y+player.direction.Y*10), player.size/3, rl.Black)
		for _, enemy := range enemies {
			rl.DrawCircle(int32(enemy.pos.X), int32(enemy.pos.Y), enemy.size, rl.Red)
			rl.DrawCircle(int32(enemy.pos.X), int32(enemy.pos.Y), enemy.size-8, rl.Black)
		}

		rl.EndMode2D()
		rl.DrawFPS(0, 0)
		rl.DrawRectangle(32, 32, 512, 32, rl.Gray)
		rl.DrawRectangle(32, 32, int32(512*player.health/player.maxHealth), 32, rl.Green)
		rl.DrawText(strconv.Itoa(player.health)+"/"+strconv.Itoa(player.maxHealth), 36, 36, 24, rl.Black)
		rl.EndDrawing()
	}
}

func movePlayer(p *Player, delta float32) {
	p.pos = rl.Vector2Add(p.pos, rl.Vector2Scale(p.velocity, delta*20))
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

func spawnEnemy(pos rl.Vector2, enemy Enemy, enemies *[]Enemy) {
	enemy.pos = pos
	*enemies = append(*enemies, enemy)
}

func moveEnemies(enemies *[]Enemy, delta float32) {
	for i, e := range *enemies {
		(*enemies)[i].pos = rl.Vector2Add(e.pos, rl.Vector2Scale((*enemies)[i].velocity, delta*10))
	}
}

func checkPlayerEnemiesCollision(p *Player, enemies *[]Enemy) {
	for _, e := range *enemies {
		if rl.CheckCollisionCircles(p.pos, p.size, e.pos, e.size) {
			p.health -= e.damage
		}
	}
}

func randPosOnEdge(center rl.Vector2, width, height, border float32) rl.Vector2 {
	xShift := width/2 + rand.Float32()*border
	yShift := height/2 + rand.Float32()*border
	isXaxis := rand.Int32N(2) == 0
	isNegative := rand.Int32N(2) == 0

	if isXaxis {
		xShift = (rand.Float32() - 0.5) * (width + border)
		if isNegative {
			yShift *= -1
		}
	} else {
		yShift = (rand.Float32() - 0.5) * (height + border)
		if isNegative {
			xShift *= -1
		}
	}

	return rl.Vector2{
		X: center.X + xShift,
		Y: center.Y + yShift,
	}
}

func attack(p Player, enemeis *[]Enemy) {
	atackPos := rl.Vector2{
		X: p.pos.X + p.direction.X*(p.size+p.atackSize/2),
		Y: p.pos.Y + p.direction.Y*(p.size+p.atackSize/2),
	}

	for i, e := range *enemeis {
		if rl.CheckCollisionCircles(e.pos, e.size, atackPos, p.atackSize) {
			(*enemeis)[i].health = e.health - p.damage
		}
	}
}

func determineEnemiesPos(enemies *[]Enemy, p Player) {
	for e1Index, e1 := range *enemies {
		(*enemies)[e1Index].velocity = rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(p.pos, e1.pos)), e1.speed)
		for e2Index, e2 := range *enemies {
			if e1Index != e2Index && rl.CheckCollisionCircles(e1.pos, e1.size, e2.pos, e2.size) {
				(*enemies)[e1Index].pos = rl.Vector2Add(e1.pos, rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(e1.pos, e2.pos)), 1))
				(*enemies)[e2Index].pos = rl.Vector2Add(e2.pos, rl.Vector2Scale(rl.Vector2Normalize(rl.Vector2Subtract(e2.pos, e1.pos)), 1))
			}
		}
	}
}
