package audio

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"ryanbrewer.page/core"
)

type music struct {
	streamer    *beep.StreamSeekCloser
	resampler   *beep.Resampler
	sample_rate beep.SampleRate
}

type sfx struct {
	buffer      *beep.Buffer
	sample_rate beep.SampleRate
}

type sfx_instance struct {
	position core.Vector2
	volume   *effects.Volume
	pan      *effects.Pan
}

type sound_file struct {
	name string
	id   core.AudioID
	kind core.AudioType
}

type music_streamer struct{}

func (s *music_streamer) Stream(samples [][2]float64) (int, bool) {
	if *paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}
	filled := 0
	for filled < len(samples) {
		if len(queue) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		n, ok := queue[0].Stream(samples[filled:])
		gain := 0.0
		if !*music_silent {
			gain = math.Pow(music_volume_base, *music_volume)
		}
		for i := range samples[:n] {
			samples[i][0] *= gain
			samples[i][1] *= gain
		}
		if !ok {
			queue = queue[1:]
			if len(queue) > 0 {
				queue[0].Seek(0)
			}
		}
		filled += n
	}
	return len(samples), true
}

func (q *music_streamer) Err() error {
	return nil
}

var (
	music_bank       map[core.AudioID]*music = make(map[core.AudioID]*music)
	sfx_bank         map[core.AudioID]*sfx   = make(map[core.AudioID]*sfx)
	fixed_samplerate beep.SampleRate         = 44100
	music_mixer      music_streamer
	sfx_mixer        beep.Mixer
	queue            []beep.StreamSeeker
	active_sfx       map[*beep.Streamer]sfx_instance = make(map[*beep.Streamer]sfx_instance)
	audio_mixer      beep.Mixer
	paused           *bool    = &core.SETTINGS_GAME_PAUSED
	music_volume     *float64 = &core.SETTINGS_MUSIC_VOLUME
	music_silent     *bool    = &core.SETTINGS_MUSIC_MUTED
	sfx_volume       *float64 = &core.SETTINGS_SFX_VOLUME
	sfx_silent       *bool    = &core.SETTINGS_SFX_MUTED
)

const music_volume_base = 2
const sfx_volume_base = 2

func Init() {
	files := [1]sound_file{{name: "0010840.mp3", id: core.SOUND_RS, kind: core.SFX}}
	for i := 0; i < len(files); i++ {
		f, err := os.Open(files[i].name)
		if err != nil {
			log.Fatal(err)
		}
		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		switch files[i].kind {
		case core.MUSIC:
			resampled := beep.Resample(4, format.SampleRate, fixed_samplerate, streamer)
			music_bank[files[i].id] = &music{
				resampler:   resampled,
				streamer:    &streamer,
				sample_rate: format.SampleRate,
			}

		case core.SFX:
			buffer := beep.NewBuffer(format)
			buffer.Append(streamer)
			streamer.Close()
			sfx_bank[files[i].id] = &sfx{
				buffer:      buffer,
				sample_rate: format.SampleRate,
			}
		}
	}

	speaker.Init(fixed_samplerate, fixed_samplerate.N(time.Second/10))
	go speaker.Play(&audio_mixer)
	audio_mixer.Add(&music_mixer)
	audio_mixer.Add(&sfx_mixer)
}

func add_to_queue(streamers ...beep.StreamSeeker) {
	queue = append(queue, streamers...)
}

func Close() {
	for _, sound := range music_bank {
		(*sound.streamer).Close()
	}
}

func PlaySFX(audio_id core.AudioID, position core.Vector2) {
	buffer := sfx_bank[audio_id].buffer
	s := buffer.Streamer(0, buffer.Len())
	resampled := beep.Resample(4, sfx_bank[audio_id].sample_rate, fixed_samplerate, s)
	volume := &effects.Volume{
		Base:     sfx_volume_base,
		Volume:   *sfx_volume,
		Streamer: resampled,
		Silent:   *sfx_silent,
	}
	pan := &effects.Pan{Streamer: volume, Pan: 0}
	var output beep.Streamer
	output = beep.Seq(pan, beep.Callback(func() {
		delete(active_sfx, &output)
	}))
	sfx_mixer.Add(output)
	active_sfx[&output] = sfx_instance{position: position, volume: volume, pan: pan}
}

func PlayMusic(audio_id core.AudioID) {
	queue = []beep.StreamSeeker{}
	add_to_queue(*music_bank[audio_id].streamer)
}

func QueueMusic(audio_id core.AudioID) {
	add_to_queue(*music_bank[audio_id].streamer)
}

func UpdateSFXBasedOnPositions(x float64) {
	speaker.Lock()
	for key := range active_sfx {
		if math.Abs(x-active_sfx[key].position.X) > core.EARSHOT {
			active_sfx[key].volume.Silent = true
			continue
		}
		active_sfx[key].volume.Silent = false
		active_sfx[key].volume.Volume = -10 * math.Abs(x-active_sfx[key].position.X) / core.EARSHOT
		active_sfx[key].pan.Pan = (active_sfx[key].position.X - x) / core.EARSHOT
	}
	speaker.Unlock()
}

func Pause() {
	speaker.Lock()
	*paused = true
	speaker.Unlock()
}

func Unpause(audio_id core.AudioID) {
	speaker.Lock()
	*paused = false
	speaker.Unlock()
}

func TogglePause(audio_id core.AudioID) {
	speaker.Lock()
	*paused = !*paused
	speaker.Unlock()
}

func SetMusicVolume(volume float64) {
	speaker.Lock()
	*music_volume = volume
	speaker.Unlock()
}

func SetSFXVolume(volume float64) {
	speaker.Lock()
	*sfx_volume = volume
	speaker.Unlock()
}

func DecibelsToVolumePercent(dB float64) float64 {
	return math.Pow(10, dB/20)
}

func VolumePercentToDecibels(vol float64) float64 {
	return 20 * math.Log10(vol)
}

// TODO: ambiant streamer that weighted-randomly plays sfx from a given list every now and then until the streamer is stopped
// TODO: sound virtualization based on in-game distance (integrate as much as possible instead of a new thing)
// TODO: add/remove effects to music_mixer while it's playing
// TODO: pitch-changing effect
// TODO: add functionality for random pitch variation for sfx
// TODO: volume-based fading effect
// TODO: panning based on in-game location
// TODO: sound importance
// TODO: "HDR"
// TODO: one-shot sfx vs longer sfx
