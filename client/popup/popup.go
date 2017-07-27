package popup

import (
	"github.com/mmogo/gxui"
	"github.com/mmogo/gxui/drivers/gl"
	"github.com/mmogo/gxui/gxfont"
	"github.com/mmogo/gxui/themes/basic"
	"github.com/mmogo/gxui/themes/dark"
)

var errorText string

func Error(err error) {
	errorText = err.Error()
	gl.StartDriver(errorPopup)
}

func errorPopup(driver gxui.Driver) {
	theme := dark.CreateTheme(driver)
	splitter := theme.CreateSplitterLayout()

	// header
	header := theme.CreateLabel()
	font, err := driver.CreateFont(gxfont.Default, 75)
	if err != nil {
		panic(err)
	}
	header.SetFont(font)
	header.SetText("MMO")
	splitter.AddChild(header)

	// error box
	errbox := basic.CreateTextBox(theme.(*basic.Theme))
	errbox.SetText(errorText)
	splitter.AddChild(errbox)

	// window
	window := theme.CreateWindow(200, 300, "MMO Error")

	// ok button
	closeBtn := theme.CreateButton()
	closeBtn.SetText("OK")
	closeBtn.OnClick(func(gxui.MouseEvent) {
		window.Close()
	})
	splitter.AddChild(closeBtn)

	window.AddChild(splitter)
	window.OnClose(driver.Terminate)

}
