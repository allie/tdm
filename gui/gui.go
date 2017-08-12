package gui

import (
	"os"
	"github.com/therecipe/qt/widgets"
)

type Gui struct {
	Window *widgets.QMainWindow
}

func NewGui() *Gui {
	g := new(Gui)
	return g
}

func (g *Gui) Init() {
	widgets.NewQApplication(len(os.Args), os.Args)
	g.Window = widgets.NewQMainWindow(nil, 0)
	g.Window.SetWindowTitle("tdm")
	g.Window.SetMinimumSize2(200, 200)
	g.Window.Show()
}

func (g *Gui) Loop() {
	widgets.QApplication_Exec()
}
