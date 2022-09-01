package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// 窗口大小
	var windowWidth float32 = 900
	var windowHigh float32 = 500
	// 登录表单大小
	//var formWidth float32 = 300
	//var formHigh float32 = 0

	myApp := app.New()		// 新建一个app
	myWindow := myApp.NewWindow("登录")		// 新建一个窗口并设置标题
	myWindow.Resize(fyne.NewSize(windowWidth,windowHigh))		// 设置窗口大小
	myWindow.CenterOnScreen()		// 窗口居中显示

	// 创建登录表单
	userLable := widget.NewLabelWithStyle("username", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	username := widget.NewEntry()
	passLable := widget.NewLabelWithStyle("password", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	password := widget.NewPasswordEntry()
	loginButton := widget.NewButton("login", func() {
		fmt.Println("username:", username.Text, "password:", password.Text)
	})

	username.Resize(fyne.NewSize(150,30))
	c1 := container.NewWithoutLayout(username)
	c2 := container.New(layout.NewHBoxLayout(), userLable, c1)

	password.Resize(fyne.NewSize(150,30))
	c3 := container.NewWithoutLayout(password)
	c4 := container.New(layout.NewHBoxLayout(), passLable, c3)

	loginButton.Resize(fyne.NewSize(100, 30))
	c5 := container.NewWithoutLayout(loginButton)

	c6 := container.New(layout.NewVBoxLayout(), c2, c4, c5)

	c6 = container.NewCenter(c6)

	// 启动
	myWindow.SetContent(c6)
	myWindow.ShowAndRun()
}