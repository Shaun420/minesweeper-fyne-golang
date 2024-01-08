package main

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
Tile extends widget.Button hence it implements all methods
of widget.Button and acts like it
*/
type Tile struct {
	id         int
	parent     *fyne.Container
	isSearched bool
	isMine     bool
	isFlagged  bool
	widget.Button
}

// Dimensions of game board
var rows, cols int = 8, 8
var total int = rows * cols

// Slices for storing mines and Tiles
var mines []*Tile
var buttons []*Tile

// Fired when user presses right click on Tile
// Flag the Tile if not already searched or flagged
func (t *Tile) TappedSecondary(*fyne.PointEvent) {
	if t.isSearched {
		return
	}
	if t.isFlagged {
		t.SetText("x")
		t.isFlagged = false
	} else {
		resourceFlagSvg, err := fyne.LoadResourceFromPath("./flag.png")
		if err != nil {
			log.Println("Error while loading flag resource.")
		}
		t.SetIcon(resourceFlagSvg)
		t.isFlagged = true
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
func newTile(id int, parent *fyne.Container, label string, icon fyne.Resource, tapped func()) *Tile {
	ret := &Tile{}
	ret.ExtendBaseWidget(ret)

	ret.id = id
	ret.parent = parent
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

	// Attach clock label for current time
	clock := widget.NewLabel("Current Time")
	div := container.NewVBox(clock)
	w.SetContent(div)
	go clock_runner(clock)

	// Create a game board and start the game
	start_game(w, div)

	// Show the window and run the app till it is closed
	w.Show()
	a.Run()
	close_app()
}

// To create a new game board
func start_game(w fyne.Window, c *fyne.Container) {
	// Get container of game board
	grid := grid_ui()
	// Attach it to main container
	c.Add(grid)
	// Resize to add padding for Tile icons
	w.Resize(fyne.NewSize(450, 450))
	generate_mines(grid)
}

// Spawn UI of game board
func grid_ui() *fyne.Container {
	t1 := time.Now()
	buttons = make([]*Tile, total)       // Store spawned Tiles
	grid := container.NewAdaptiveGrid(8) // fyne.Container holding all Tiles
	for i := 0; i < total; i++ {
		id := i
		buttons[i] = newTile(i, grid, "x", nil, func() {
			if buttons[id].isFlagged || buttons[id].isSearched {
				return
			}
			if buttons[id].isMine {
				resourceMineSvg, err := fyne.LoadResourceFromPath("./mine.svg")
				if err != nil {
					log.Println("Error while loading mine resource. Error:", err)
				}
				buttons[id].isSearched = true
				buttons[id].SetIcon(resourceMineSvg)
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
	var mines_no int = 12 // No. of mines to spawn
	var mine int = 0      // No. of mines spawned
	var id = 0            // Tile id
	// Store created mines in mines slice to avoid repeating same mine
	mines = make([]*Tile, total)
	fmt.Printf("Mine ids: ")
	for {
		if mine == mines_no {
			// Stop if enough mines are spawned
			break
		}
		id = rand.Intn(total)
		if slices.Contains(mines, buttons[id]) {
			// Do not spawn mine on already spawned mine
			continue
		}
		buttons[id].isMine = true
		mines = append(mines, buttons[id])
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

// Updates top label with current time each second
func clock_runner(c *widget.Label) {
	for range time.Tick(time.Second) {
		formatted := time.Now().Format(time.UnixDate)
		c.SetText(formatted)
	}
}

// Called when app window is closed
func close_app() {
	clear(buttons)
	clear(mines)
	fmt.Println("Exitting the program.")
}
