package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyneMusic/myTheme"
	"os"
)

var Window fyne.Window
var App fyne.App

func RunApp() {
	App = app.NewWithID("music")				// 创建APP
	Window = App.NewWindow("网易云音乐")			// 创建窗口
	App.SetIcon(myTheme.ResourceLogoJpg)			    	// 设置logo
	App.Settings().SetTheme(&myTheme.MyTheme{})			// 设置APP主题，嵌入字体，解决乱码
	Window.Resize(fyne.NewSize(1200,800))			// 设置窗口大小
	Window.CenterOnScreen()								// 窗口居中显示
	Window.SetMaster()									// 设置为主窗口
	Window.SetCloseIntercept(func() {Window.Hide()})	// 设置窗口托盘显示
	if desk, ok := App.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				Window.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	initPreferences()
	Window.SetMainMenu(MakeMyMenu(App, Window))
	//Window.SetContent(CreateApp(App, Window))
	Window.SetContent(MakeNav(App, Window))

	Window.ShowAndRun()
}

func CreateApp(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	c1 := MakeMusicSearch()
	c2 := MakeMusicList()
	c3 := MakeMusicPlayer()
	c := container.NewBorder(c1,c3,nil,nil,c2)
	return c
}

//初始化Preferences变量
func initPreferences()  {
	downloadPath :=  App.Preferences().String("savePath")
	miguServer := App.Preferences().String("miguServer")
	neteaseServer := App.Preferences().String("neteaseServer")
	if _,err := os.Stat(downloadPath); err == nil {
		MyPlayer.DownloadPath = downloadPath
	}
	if _,err := os.Stat(miguServer); err == nil {
		MyPlayer.MiguServer = miguServer
	}
	if _,err := os.Stat(neteaseServer); err == nil {
		MyPlayer.NeteaseServer = neteaseServer
	}
}

