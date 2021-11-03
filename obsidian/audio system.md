DESIGN: https://www.reddit.com/r/GameAudio/comments/61b5hf/where_to_learn_game_audio_engine_architecture/dfejw2q/?context=8&depth=9
FOUNDATION: https://github.com/faiface/beep, https://pkg.go.dev/github.com/faiface/beep
PLAN:
 - ideally we can have a few songs that we run through
 - ideally we can have effects and tracks we can add to any song that make them more intense (with different levels of intensity) or fit a certain environment, etc. Vertically-mixed adaptive audio
 - ideally the achievement-based hero's journey can feed into choosing when to play what levels of intensity or whatever effects
 - we need to figure out a plan for how to make music sound more peaceful, since simply adding a track won't work well and we can't just slow down the audio lol
 - the vertical mixing needs to be immediately responsive to in-game triggers
 - we also need shorter sound effects that can be played on demand in response to an in-game trigger
 - ambiant streamer that weighted-randomly plays sfx from a given list every now and then until the streamer is stopped
 - sound virtualization based on in-game distance (integrate as much as possible instead of a new thing)
 - add/remove effects to music_mixer and ambient_mixer while it's playing
 - pitch-changing effect (this might already exist)
 - add functionality for random pitch variation for sfx variety
 - volume-based fading effect (this might already exist)
 - panning based on in-game location
 - sound importance system
 - "HDR" (this needs research lol)
 - one-shot sfx vs longer sfx distinction (important for virtualization)