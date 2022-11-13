package main

import (
	"fyneMusic/music"
	"log"
)

func init()  {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main()  {
	music.RunApp()                          // 运行UI
}
