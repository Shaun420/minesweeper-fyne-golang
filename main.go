//go:generate fyne bundle -o bundled.go image1.png
//go:generate fyne bundle -o bundled.go -append image2.png

package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

/*
Tile extends widget.Button hence it implements all methods
of widget.Button and acts like it
*/
type Tile struct {
	id         int
	isSearched bool
	isMine     bool
	isFlagged  bool
	widget.Button
}

// Custom fyne theme
type apptheme struct{}

var _ fyne.Theme = (*apptheme)(nil)

// Dimensions of game board
var rows, cols int = 8, 8
var total int = rows * cols // No. of Tiles to spawn
var totalmines int = 12     // No. of mines to spawn

/*
	To check the state of the game

0 - Not Started
1 - Currently playing
2 - Ended
*/
var gameState int = 0

// Time when game has started
var gameStartTime time.Time

// Slices for storing mines and Tiles
var buttons []*Tile = make([]*Tile, total)
var mines []*Tile = make([]*Tile, totalmines)

// Fired when user presses right click on Tile
// Flag the Tile if not already searched or flagged
func (t *Tile) TappedSecondary(*fyne.PointEvent) {
	if t.isSearched {
		return
	}
	if t.isFlagged {
		t.SetText("x")
		t.SetIcon(nil)
		t.isFlagged = false
	} else {
		t.SetText("")
		t.SetIcon(resourceFlagPng)
		t.isFlagged = true
		if get_unflagged_mines() == 0 {
			// No mines left to flag
			end_game()
		}
	}
	if gameState == 0 {
		gameState = 1
		gameStartTime = time.Now()
	}

	log.Printf("Flagged %d tile.\n", t.id)
}

/* Placeholder handlers, not required

func (t *Tile) DoubleTapped(*fyne.PointEvent) {
	resourceMineSvg, err := fyne.LoadResourceFromPath("./mine.svg")
	if err != nil {
		log.Println("Error while loading mine resource.")
	}
	// fmt.Printf("Mine: %v\n", resourceMineSvg)

	t.SetIcon(resourceMineSvg)

	fmt.Printf("Tile %d pos: %+v\n", t.id, t.Position())
	log.Printf("Double Tapped %d tile.\n", t.id)
}

func (t *Tile) Dragged(*fyne.DragEvent) {
	log.Printf("Dragged %d tile.\n", t.id)
}

func (t *Tile) DragEnd() {
	log.Printf("Dragging Ended %d tile.\n", t.id)
}*/

// Create a new Tile extending widget.Button and initialize it
func newTile(id int, label string, icon fyne.Resource, tapped func()) *Tile {
	ret := &Tile{}
	ret.ExtendBaseWidget(ret)

	ret.id = id
	ret.isSearched = false
	ret.isMine = false
	ret.isFlagged = false
	ret.Text = label
	ret.Icon = icon
	ret.OnTapped = tapped

	return ret
}

func main() {
	// Make fyne.App instance
	a := app.New()
	// Open a window in the app
	w := a.NewWindow("Minesweeper")
	// Set apptheme as theme
	a.Settings().SetTheme(&apptheme{})

	// Attach clock label for current time
	clock := widget.NewLabel("Current Time")
	maindiv := container.NewVBox(clock)
	w.SetContent(maindiv)
	go clock_runner(clock)

	// Create a restart button
	restartbtn := widget.NewButtonWithIcon("Restart", theme.ViewRefreshIcon(), func() {
		gameState = 0
		start_game(w, maindiv)
	})
	maindiv.Add(restartbtn)

	// Create a game board and start the game
	start_game(w, maindiv)

	// Show the window and run the app till it is closed
	w.Show()
	a.Run()
	close_app()
}

// To create a new game board
func start_game(w fyne.Window, c *fyne.Container) {
	// Delete previous game board
	if len(c.Objects) > 2 {
		c.Objects[2].Hide()
		c.Remove(c.Objects[2])
	}
	// Create container of game board
	board := grid_ui()
	// Attach it to main container
	c.Add(board)
	// Resize to add padding for Tile icons
	w.Resize(fyne.NewSize(450, 450))
	generate_mines(board)
}

// To show all mines and end the game
func end_game() {
	// Disable all buttons on Tiles and show all mines
	for _, t := range buttons {
		if t.isMine {
			if t.isFlagged {
				t.SetIcon(resourceFlaggedMinePng)
			} else {
				t.SetIcon(resourceMineSvg)
			}
			t.SetText("")
		}
		t.isSearched = true
	}
	gameState = 2
	fmt.Println("Game ended")
}

// Spawn UI of game board
func grid_ui() *fyne.Container {
	t1 := time.Now()
	buttons = make([]*Tile, total)       // Store spawned Tiles
	grid := container.NewAdaptiveGrid(8) // fyne.Container holding all Tiles
	for i := 0; i < total; i++ {
		id := i
		buttons[i] = newTile(i, "x", nil, func() {
			if buttons[id].isFlagged || buttons[id].isSearched {
				return
			}
			if gameState == 0 {
				gameState = 1
				gameStartTime = time.Now()
			}
			if buttons[id].isMine {
				// Mine exploded
				buttons[id].SetIcon(resourceMineSvg)
				end_game()
			} else {
				buttons[id].isSearched = true
				buttons[id].Disable()
				buttons[id].SetText(strconv.Itoa(count_nearby_mines(buttons[id])))
				buttons[id].SetIcon(nil)
			}
			log.Printf("Button %d clicked\n", id)
		})
		grid.Add(buttons[i])
	}
	fmt.Println("gridui duration:", time.Since(t1))
	return grid
}

// Choose which Tiles to make mines
func generate_mines(grid *fyne.Container) {
	var mine int = 0 // No. of mines spawned
	var id = 0       // Tile id
	// Store created mines in mines slice to avoid repeating same mine
	clear(mines)
	fmt.Printf("Mine ids: ")
	for {
		if mine == totalmines {
			// Stop if enough mines are spawned
			break
		}
		id = rand.Intn(total)
		if slices.Contains(mines, buttons[id]) {
			// Do not spawn mine on already spawned mine
			continue
		}
		buttons[id].isMine = true
		mines[mine] = buttons[id]
		mine++
		fmt.Printf("%d ", id)
		/* Debugging code
		resourceMineSvg, err := fyne.LoadResourceFromPath("./mine.svg")
		if err != nil {
			log.Println("Error while loading mine resource. Error:", err)
		}
		buttons[id].SetIcon(resourceMineSvg)
		*/
	}
	fmt.Printf("\n")
}

// Finds the no. of mines near a given Tile
func count_nearby_mines(t *Tile) int {
	// Min 3 Max 8
	var mines int = 0
	var col int = t.id % cols

	// Top left
	if is_valid_id(t.id-cols-1) && (col != 0) {
		if buttons[t.id-cols-1].isMine {
			mines++
		}
	}
	// Top middle
	if is_valid_id(t.id - cols) {
		if buttons[t.id-cols].isMine {
			mines++
		}
	}
	// Top right
	if is_valid_id(t.id-cols+1) && (col != cols-1) {
		if buttons[t.id-cols+1].isMine {
			mines++
		}
	}
	// Middle left
	if is_valid_id(t.id-1) && (col != 0) {
		if buttons[t.id-1].isMine {
			mines++
		}
	}
	// Middle right
	if is_valid_id(t.id+1) && (col != cols-1) {
		if buttons[t.id+1].isMine {
			mines++
		}
	}
	// Bottom left
	if is_valid_id(t.id+cols-1) && (col != 0) {
		if buttons[t.id+cols-1].isMine {
			mines++
		}
	}
	// Bottom middle
	if is_valid_id(t.id + cols) {
		if buttons[t.id+cols].isMine {
			mines++
		}
	}
	// Bottom right
	if is_valid_id(t.id+cols+1) && (col != cols-1) {
		if buttons[t.id+cols+1].isMine {
			mines++
		}
	}

	// fmt.Println("mines:", mines)
	return mines
}

// Checks if id is within bounds
func is_valid_id(id int) bool {
	return (id >= 0) && (id < total)
}

// Get no. of mines that are not flagged
func get_unflagged_mines() int {
	// fmt.Println("Mines:", mines)
	count := 0
	for _, t := range mines {
		if !t.isFlagged {
			count++
		}
	}
	return count
}

// Updates top label with current time each second
func clock_runner(c *widget.Label) {
	for range time.Tick(time.Second) {
		if gameState == 0 {
			formatted := time.Now().Format(time.UnixDate)
			c.SetText(formatted)
		} else if gameState == 1 {
			c.SetText(time.Since(gameStartTime).Truncate(time.Second).String())
		}
	}
}

// Theme methods
func (m apptheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (m apptheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m apptheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m apptheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Called when app window is closed
func close_app() {
	clear(buttons)
	clear(mines)
	fmt.Println("Exitting the program.")
}
