package myWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const preferenceCurrentTutorial = "currentTutorial"
var content = container.NewMax()

// Screen 页面布局
type Screen struct {
	Title string
	View func(w fyne.Window) fyne.CanvasObject
}

// Nav 导航栏与CanvasObject的对应关系
var Nav = map[string]Screen{
	"music": {"网易云音乐", MusicTable},
	//"portScan": {"端口扫描", MakePortScan},
	//"nav1": {"nav1", nav1},
	//"nav2": {"nav2", nav2},
	//"nav3": {"nav3", nav3},
	//"nav4": {"nav4", nav4},
	//"nav5": {"nav5", nav5},
}

// NavIndex 导航栏子节点对应关系
//var NavIndex = map[string][]string{
//	"": {"music","portScan", "nav1", "nav2", "nav3"},		// 没有叶子节点
//	"nav3": {"nav4", "nav5"},		// 有子节点
//}
var NavIndex = map[string][]string{
	"": {"music"},
}

func nav1(win fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("hello1")
}
func nav2(win fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("hello2")
}
func nav3(win fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("hello3")
}
func nav4(win fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("hello4")
}
func nav5(win fyne.Window) fyne.CanvasObject {
	return widget.NewLabel("hello5")
}

func MakeNav(myApp fyne.App, myWindow fyne.Window) fyne.CanvasObject {
	//content := container.NewMax()

	setTutorial := func(t Screen) {
		content.Objects = []fyne.CanvasObject{t.View(myWindow)}
		//content.Refresh()
	}

	tutorial := container.NewBorder(container.NewVBox(widget.NewSeparator()), nil, nil, nil, content)
	split := container.NewHSplit(makeNav(setTutorial, true), tutorial)
	split.Offset = 0		// 调整分割线：0~1
	return  split
}

func makeNav(setTutorial func(tutorial Screen), loadPrevious bool) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return NavIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := NavIndex[uid]
			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := Nav[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
			obj.(*widget.Label).TextStyle = fyne.TextStyle{Italic: true}
			//obj.(*widget.Label).TextStyle = fyne.TextStyle{}

		},
		OnSelected: func(uid string) {
			if t, ok := Nav[uid]; ok {
				a.Preferences().SetString(preferenceCurrentTutorial, uid)
				setTutorial(t)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentTutorial, "welcome")
		tree.Select(currentPref)
	}

	// 设置自定义主题后，字体bug
	//themes := container.NewGridWithColumns(2,
	//	widget.NewButton("Dark", func() {
	//		a.Settings().SetTheme(theme.DarkTheme())
	//	}),
	//	widget.NewButton("Light", func() {
	//		a.Settings().SetTheme(theme.LightTheme())
	//	}),
	//)

	return container.NewBorder(nil, nil, nil, nil, tree)
}