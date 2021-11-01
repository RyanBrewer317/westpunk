package audio

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"rbrewer.com/core"
)

type sound struct {
	control  *beep.Ctrl
	streamer *beep.StreamSeekCloser
	volume   *effects.Volume
	speed    *beep.Resampler
}

type sound_file struct {
	name string
	id   core.AudioID
}

var audio_bank map[core.AudioID]*sound = make(map[core.AudioID]*sound)
var format beep.Format
var CTRL beep.Ctrl

func Init() {
	var streamer beep.StreamSeekCloser
	files := [1]sound_file{sound_file{name: "0010840.mp3", id: core.SOUND_RS}}
	for i := 0; i < len(files); i++ {
		f, err := os.Open(files[i].name)
		if err != nil {
			log.Fatal(err)
		}
		streamer, format, err = mp3.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		control := &beep.Ctrl{
			Streamer: streamer,
			Paused:   false,
		}
		volume := &effects.Volume{
			Streamer: control,
			Base:     2,
			Volume:   0,
			Silent:   false,
		}
		speed := beep.ResampleRatio(4, 1, volume.Streamer)
		audio_bank[files[i].id] = &sound{
			control:  control,
			streamer: &streamer,
			volume:   volume,
			speed:    speed,
		}
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func Play(audio_id core.AudioID) {
	speaker.Play(audio_bank[audio_id].speed)
}

func Test() {
	Play(core.SOUND_RS)
}

func Close() {
	for _, sound := range audio_bank {
		(*sound.streamer).Close()
	}
}

func Pause(audio_id core.AudioID) {
	speaker.Lock()
	audio_bank[audio_id].control.Paused = true
	speaker.Unlock()
}

func Unpause(audio_id core.AudioID) {
	speaker.Lock()
	audio_bank[audio_id].control.Paused = false
	speaker.Unlock()
}

func TogglePause(audio_id core.AudioID) {
	speaker.Lock()
	audio_bank[audio_id].control.Paused = !audio_bank[audio_id].control.Paused
	speaker.Unlock()
}

func SetVolume(audio_id core.AudioID, volume float64, differential bool) {
	speaker.Lock()
	audio_bank[audio_id].volume.Volume += volume
	if !differential {
		audio_bank[audio_id].volume.Volume = volume
	}
	speaker.Unlock()
}

func SetSpeedRatio(audio_id core.AudioID, amount float64, differential bool) {
	speaker.Lock()
	audio_bank[audio_id].speed.SetRatio(audio_bank[audio_id].speed.Ratio() + amount)
	if !differential {
		audio_bank[audio_id].speed.SetRatio(amount)
	}
	speaker.Unlock()
}
