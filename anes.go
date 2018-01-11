package main

import "./nespkg"

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
)

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var nesPalette = []color.Color{
	color.RGBA{0x7c, 0x7c, 0x7c, 0xff},
	color.RGBA{0x00, 0x00, 0xfc, 0xff},
	color.RGBA{0x00, 0x00, 0xbc, 0xff},
	color.RGBA{0x44, 0x28, 0xbc, 0xff},
	color.RGBA{0x94, 0x00, 0x84, 0xff},
	color.RGBA{0xa8, 0x00, 0x20, 0xff},
	color.RGBA{0xa8, 0x10, 0x00, 0xff},
	color.RGBA{0x88, 0x14, 0x00, 0xff},
	color.RGBA{0x50, 0x30, 0x00, 0xff},
	color.RGBA{0x00, 0x78, 0x00, 0xff},
	color.RGBA{0x00, 0x68, 0x00, 0xff},
	color.RGBA{0x00, 0x58, 0x00, 0xff},
	color.RGBA{0x00, 0x40, 0x58, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0xbc, 0xbc, 0xbc, 0xff},
	color.RGBA{0x00, 0x78, 0xf8, 0xff},
	color.RGBA{0x00, 0x58, 0xf8, 0xff},
	color.RGBA{0x68, 0x44, 0xfc, 0xff},
	color.RGBA{0xd8, 0x00, 0xcc, 0xff},
	color.RGBA{0xe4, 0x00, 0x58, 0xff},
	color.RGBA{0xf8, 0x38, 0x00, 0xff},
	color.RGBA{0xe4, 0x5c, 0x10, 0xff},
	color.RGBA{0xac, 0x7c, 0x00, 0xff},
	color.RGBA{0x00, 0xb8, 0x00, 0xff},
	color.RGBA{0x00, 0xa8, 0x00, 0xff},
	color.RGBA{0x00, 0xa8, 0x44, 0xff},
	color.RGBA{0x00, 0x88, 0x88, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0xf8, 0xf8, 0xf8, 0xff},
	color.RGBA{0x3c, 0xbc, 0xfc, 0xff},
	color.RGBA{0x68, 0x88, 0xfc, 0xff},
	color.RGBA{0x98, 0x78, 0xf8, 0xff},
	color.RGBA{0xf8, 0x78, 0xf8, 0xff},
	color.RGBA{0xf8, 0x58, 0x98, 0xff},
	color.RGBA{0xf8, 0x78, 0x58, 0xff},
	color.RGBA{0xfc, 0xa0, 0x44, 0xff},
	color.RGBA{0xf8, 0xb8, 0x00, 0xff},
	color.RGBA{0xb8, 0xf8, 0x18, 0xff},
	color.RGBA{0x58, 0xd8, 0x54, 0xff},
	color.RGBA{0x58, 0xf8, 0x98, 0xff},
	color.RGBA{0x00, 0xe8, 0xd8, 0xff},
	color.RGBA{0x78, 0x78, 0x78, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0xfc, 0xfc, 0xfc, 0xff},
	color.RGBA{0xa4, 0xe4, 0xfc, 0xff},
	color.RGBA{0xb8, 0xb8, 0xf8, 0xff},
	color.RGBA{0xd8, 0xb8, 0xf8, 0xff},
	color.RGBA{0xf8, 0xb8, 0xf8, 0xff},
	color.RGBA{0xf8, 0xa4, 0xc0, 0xff},
	color.RGBA{0xf0, 0xd0, 0xb0, 0xff},
	color.RGBA{0xfc, 0xe0, 0xa8, 0xff},
	color.RGBA{0xf8, 0xd8, 0x78, 0xff},
	color.RGBA{0xd8, 0xf8, 0x78, 0xff},
	color.RGBA{0xb8, 0xf8, 0xb8, 0xff},
	color.RGBA{0xb8, 0xf8, 0xd8, 0xff},
	color.RGBA{0x00, 0xfc, 0xfc, 0xff},
	color.RGBA{0xf8, 0xd8, 0xf8, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0x00, 0xff},
}

var gamepadButtonMap = map[walk.Key]nespkg.GamepadButton{
	walk.KeyH: nespkg.ButtonLeft,
	walk.KeyJ: nespkg.ButtonDown,
	walk.KeyK: nespkg.ButtonUp,
	walk.KeyL: nespkg.ButtonRight,
	walk.KeyZ: nespkg.ButtonB,
	walk.KeyX: nespkg.ButtonA,
	walk.Key1: nespkg.ButtonSelect,
	walk.Key2: nespkg.ButtonStart,
}

type MyMainWindow struct {
	*walk.MainWindow
	paintWidget MyCustomWidget
}

type MyCustomWidget struct {
	*walk.CustomWidget
	display            *NesDisplay
	nes                *nespkg.Nes
	myGoRoutineCreated bool
}

func (mcw *MyCustomWidget) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_LBUTTONDOWN:
		if !mcw.myGoRoutineCreated {
			nespkg.Debug("creating timer routine\n")
			go myGoRoutine(mcw)
			mcw.myGoRoutineCreated = true
			nespkg.Debug("goroutine created\n")
		}
	case win.WM_RBUTTONDOWN:
		nespkg.Debug("Stop CPU\n")
		mcw.nes.Stop()
	}
	return mcw.CustomWidget.WndProc(hwnd, msg, wParam, lParam)
}

func makeKeyDownHandler(pad *nespkg.Gamepad) func(key walk.Key) {
	return func(key walk.Key) {
		button, ok := gamepadButtonMap[key]
		if ok {
			pad.SetButtonState(button, true)
		}
	}
}

func makeKeyUpHandler(pad *nespkg.Gamepad) func(key walk.Key) {
	return func(key walk.Key) {
		button, ok := gamepadButtonMap[key]
		if ok {
			pad.SetButtonState(button, false)
		}
	}
}

func makePaintFunc(display *NesDisplay) func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
	return func(canvas *walk.Canvas, updateBounds walk.Rectangle) error {
		bmp, err := createBitmap(display)
		if err != nil {
			return err
		}
		defer bmp.Dispose()

		if err := canvas.DrawImageStretched(bmp, updateBounds); err != nil {
			return err
		}

		return nil
	}
}

func createBitmap(display *NesDisplay) (*walk.Bitmap, error) {
	r := image.Rectangle{image.Point{0, 0}, image.Point{nespkg.ScreenSizePixX, nespkg.ScreenSizePixY}}
	im := image.NewPaletted(r, nesPalette)
	if display.screen != nil {
		for y := 0; y < nespkg.ScreenSizePixY; y++ {
			for x := 0; x < nespkg.ScreenSizePixX; x++ {
				im.SetColorIndex(x, y, uint8(display.screen[y][x]))
			}
		}
	} else {
		for y := 0; y < nespkg.ScreenSizePixY; y++ {
			for x := 0; x < nespkg.ScreenSizePixX; x++ {
				im.SetColorIndex(x, y, uint8(y%len(nesPalette)))
			}
		}
	}

	bmp, err := walk.NewBitmapFromImage(im)
	if err != nil {
		return nil, err
	}

	return bmp, nil
}

func runMyWidget(display *NesDisplay, nes *nespkg.Nes) {
	var mw *walk.MainWindow
	if err := (MainWindow{
		AssignTo:  &mw,
		Title:     "ANES",
		Size:      Size{nespkg.ScreenSizePixX * 2, nespkg.ScreenSizePixY * 2},
		OnKeyDown: makeKeyDownHandler(nes.Pad),
		OnKeyUp:   makeKeyUpHandler(nes.Pad),
		Layout:    VBox{MarginsZero: true},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	mcw, err := NewMyCustomWidget(mw, display, nes)
	if err != nil {
		log.Fatal(err)
	}

	display.mcw = mcw
	nespkg.Debug("Calling mw.Run()\n")
	mw.Run()
}

func NewConf() *nespkg.Conf {
	conf := new(nespkg.Conf)
	flag.BoolVar(&conf.DebugEnable, "d", false, "Enable debug mode")
	flag.BoolVar(&conf.TraceEnable, "t", false, "Enable instruction trace")
	flag.BoolVar(&conf.MemTraceEnable, "m", false, "Enable memory trace")
	flag.Parse()
	fmt.Println("debug: ", conf.DebugEnable)
	fmt.Println("instruction trace on: ", conf.TraceEnable)
	fmt.Println("memory trace on: ", conf.MemTraceEnable)
	return conf
}

func main() {
	conf := NewConf()
	display := NewNesDisplay()
	nes := nespkg.NewNes(conf, display)
	if len(flag.Args()) >= 1 {
		nespkg.Debug("loading: %s\n", flag.Arg(0))
		err := nes.LoadRom(flag.Arg(0))
		if err != nil {
			fmt.Println("ROM format error")
			return
		}
	} else {
		nespkg.Debug("invalid argument\n")
		return
	}

	runMyWidget(display, nes)
}

func myGoRoutine(mcw *MyCustomWidget) {
	mcw.nes.Run()
}

func NewMyCustomWidget(parent walk.Container, display *NesDisplay, nes *nespkg.Nes) (*MyCustomWidget, error) {

	paintfunc := makePaintFunc(display)
	cw, err := walk.NewCustomWidget(parent, win.WS_VISIBLE, paintfunc)
	if err != nil {
		return nil, err
	}

	mcw := &MyCustomWidget{cw, display, nes, false}
	mcw.SetClearsBackground(true)

	if err := walk.InitWrapperWindow(mcw); err != nil {
		return nil, err
	}

	return mcw, nil
}

type NesDisplay struct {
	palette []color.Color
	screen  *[nespkg.ScreenSizePixY][nespkg.ScreenSizePixX]uint8
	mcw     *MyCustomWidget
}

func (ns *NesDisplay) Render(screen *[nespkg.ScreenSizePixY][nespkg.ScreenSizePixX]uint8) {
	ns.screen = screen
	ns.mcw.Invalidate()
}

func NewNesDisplay() *NesDisplay {
	nd := new(NesDisplay)
	nd.palette = nesPalette
	return nd
}
