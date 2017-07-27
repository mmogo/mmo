package main

import (
	"log"

	"github.com/mmogo/gxui"
	"github.com/mmogo/gxui/drivers/gl"
	"github.com/mmogo/gxui/gxfont"
	"github.com/mmogo/gxui/themes/basic"
	"github.com/mmogo/gxui/themes/dark"
)

var errtext string
var idtext string
var srvtext string

func gui() {
	idtext = *playerID
	srvtext = *addr
	gl.StartDriver(patcherApp)
}

func patcherApp(driver gxui.Driver) {
	theme := dark.CreateTheme(driver)
	splitter := theme.CreateSplitterLayout()
	// input id
	header := theme.CreateLabel()
	font, err := driver.CreateFont(gxfont.Default, 75)
	if err != nil {
		panic(err)
	}
	header.SetFont(font)
	header.SetText("MMO")
	splitter.AddChild(header)
	idlabel := theme.CreateLabel()
	idlabel.SetText("ID")
	idbox := basic.CreateTextBox(theme.(*basic.Theme))
	idbox.SetText(idtext)
	log.Println(idbox.Text())
	srvlabel := theme.CreateLabel()
	srvlabel.SetText("Server")
	addrbox := basic.CreateTextBox(theme.(*basic.Theme))
	addrbox.SetText(srvtext)
	log.Println(addrbox.Text())
	splitter.AddChild(idlabel)
	splitter.AddChild(idbox)
	splitter.AddChild(srvlabel)
	splitter.AddChild(addrbox)
	// errors
	errbox := basic.CreateTextBox(theme.(*basic.Theme))
	errbox.SetTextColor(gxui.Red)
	window := theme.CreateWindow(800, 600, "MMO")
	clientName := getClientName()
	downloadBtn := theme.CreateButton()
	downloadBtn.SetText("Check for update")
	downloadBtn.OnClick(func(gxui.MouseEvent) {
		*addr = addrbox.Text()
		if err := downloadClient(clientName); err != nil {
			errtext += err.Error() + "\n"
			errbox.SetText(errtext)
			log.Println(err)
		} else {
			errbox.SetTextColor(gxui.Green)
			errbox.SetText("client up to date. click play!")
		}

		// ungrey playBtn

	})

	playBtn := theme.CreateButton()
	playBtn.SetText("Play")

	playBtn.OnClick(func(gxui.MouseEvent) {
		*addr = addrbox.Text()
		*playerID = idbox.Text()
		window.Close()
		runClient(clientName)
	})

	splitter.AddChild(downloadBtn)
	splitter.AddChild(playBtn)
	splitter.AddChild(errbox)

	window.AddChild(splitter)
	window.OnClose(driver.Terminate)

}
