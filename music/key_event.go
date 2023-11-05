package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"time"
)

var searchInterval = 0

func ListenKeyEvent(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeySpace:
		if ctrl != nil {
			ctrl.Paused = !ctrl.Paused
			if ctrl.Paused {
				mp.PlayerPause.SetIcon(theme.MediaPlayIcon())
				mp.PlayerPause.SetText("")
			} else {
				mp.PlayerPause.SetIcon(theme.MediaPauseIcon())
				mp.PlayerPause.SetText("")
			}
		}
	case fyne.KeyReturn:
		if searchInterval != 0 {
			return
		}
		go func() {
			searchInterval = 3
			for searchInterval > 0 {
				time.Sleep(time.Second)
				searchInterval--
			}
		}()
		searchFunc(sw.SearchEngine.Text, sw.SearchEntry.Text)
	}
}
