package main

import (
	musicplayer "demo/example/pkg"
)

func main() {
	app := &musicplayer.AppGUI{}
	app.Run()
}