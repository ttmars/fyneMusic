package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyneMusic/static/icon"
	"os"
)

var Window fyne.Window
var App fyne.App

func RunApp() {
	App = app.NewWithID("fyne_music") // 创建APP
	Window = App.NewWindow("网易云音乐")   // 创建窗口
	App.SetIcon(icon.ResourceLogoJpg) // 设置logo
	//App.Settings().SetTheme(&font.MyTheme{})           // 设置APP主题，嵌入字体，解决乱码
	Window.Resize(fyne.NewSize(1200, 800))             // 设置窗口大小
	Window.CenterOnScreen()                            // 窗口居中显示
	Window.SetMaster()                                 // 设置为主窗口
	Window.SetCloseIntercept(func() { Window.Hide() }) // 设置窗口托盘显示
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
	Window.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		ListenKeyEvent(event)
	})

	go MyPlayer.PlayMusic()           // 打开播放器
	go MyPlayer.InitPlayList()        // 异步加载数据
	go MyPlayer.RandomPlay()          // 随机播放
	go MyPlayer.UpdateProgressLabel() // 动态更新进度条、歌词
	go MyPlayer.UpdateSongName()      // 动态更新播放歌名

	Window.ShowAndRun()
}

func CreateApp(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	c1 := MakeMusicSearch()
	c2 := MakeMusicList()
	c3 := MakeMusicPlayer()
	c := container.NewBorder(c1, c3, nil, nil, c2)
	return c
}

// 初始化Preferences变量
func initPreferences() {
	downloadPath := App.Preferences().String("downloadPath")
	if _, err := os.Stat(downloadPath); err == nil {
		MyPlayer.DownloadPath = downloadPath
	}

	miguServer := App.Preferences().String("miguServer")
	if miguServer != "" {
		MyPlayer.MiguServer = miguServer
	}

	neteaseServer := App.Preferences().String("neteaseServer")
	if neteaseServer != "" {
		MyPlayer.NeteaseServer = neteaseServer
	}
}
