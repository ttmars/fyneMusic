package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	f, err := os.Open("D:\\project\\goProject\\Demo\\小小的太陽_孙露.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	streamer1 := beep.ResampleRatio(14, 1, streamer)

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))



	done := make(chan bool)
	speaker.Play(beep.Seq(streamer1, beep.Callback(func() {
		done <- true
	})))

	<-done
	//var speed float64 = 1
	//for {
	//	fmt.Print("Press [ENTER] to pause/resume. ")
	//	fmt.Scanln()
	//
	//	speaker.Lock()
	//	streamer1.SetRatio(speed)
	//	speaker.Unlock()
	//	speed += 0.2
	//}
}