package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"

	"fyne.io/fyne/v2/app"
)

func main() {
	// 窗口大小
	var windowWidth float32 = 900
	var windowHigh float32 = 500
	// 登录表单大小
	var formWidth float32 = 300
	var formHigh float32 = 0

	myApp := app.New()		// 新建一个app
	myWindow := myApp.NewWindow("登录")		// 新建一个窗口并设置标题
	myWindow.Resize(fyne.NewSize(windowWidth,windowHigh))		// 设置窗口大小
	myWindow.CenterOnScreen()		// 窗口居中显示

	// 创建登录表单
	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "username", Widget: username},
			{Text: "password", Widget: password},
		},
		OnSubmit: func() {
			log.Println("Form submitted:", username.Text)
			log.Println("multiline:", password.Text)
		},
		SubmitText: "login",
	}

	// 通过创建无布局的容器设置表单大小
	form.Resize(fyne.NewSize(formWidth,formHigh))
	c := container.NewWithoutLayout(form)

	// 居中显示，并非真正居中...
	c = container.NewCenter(c)

	// 启动
	myWindow.SetContent(c)
	myWindow.ShowAndRun()
}