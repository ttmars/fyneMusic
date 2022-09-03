package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"net/http"
)

func main() {
	myApp := app.New()
	w := myApp.NewWindow("Image")

	u := "https://mcontent.migu.cn/newlv2/new/album/20210410/1000002142/s_OiN9DXXpHPOLD2Td.jpg"

	image := createImage(u)

	w.SetContent(image)

	w.ShowAndRun()
}

func createImage(pic string) fyne.CanvasObject {
	r,_ := http.Get(pic)
	defer r.Body.Close()
	image := canvas.NewImageFromReader(r.Body, "jpg")
	//image.FillMode = canvas.ImageFillOriginal
	return image
}