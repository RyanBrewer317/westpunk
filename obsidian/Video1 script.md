im ryan, a student and developer
for past few months ive been planning and making a game
got to a point where im like hmmmmm this could be something
so this is the start of my devlog series
The game doesn't use a game engine

...

and also is written in Go instead of, say, C++

...

But I mean it's doing alright so far lol

I'm using the Ebiten library for fast game graphics with GPU support
And the physics and rendering systems are all homemade
_Physics use a Unity-like component struct system that's vaguely object-oriented instead of the more data-oriented ECS because I think it'll work better with chunkloading and I just don't need that much speed. The rest of the game takes a more data-oriented approach however._

OH YEAH yeah the code isn't object oriented (:

The design of the game allows it to be both complex and lightweight
the render distance is small
So only about a few dozen objects would be loaded in at a time, with a conceivable max of a hundred or so
The player is a 2D rigged system, as if it were 3D
so I can have customizeably appearance and complex movement
The animation has two stances for every walking state
that it interpolates between
and also interpolates between changing states
Which has this extremely smooth effect
but in the future I want to change a lot
namely the twelve rules of animation
and inverse kinematic walking
I have the general technology figured out
but rewriting the animation system that much is hard
and this is fine for now
...
Also there's no parallax which is bad I know, I'll add that later
...
The game's design is to be a sandbox survival fighting game
like in brawlhalla, weapons will give you the available moves
the survival aspect obviously resembles terraria
but both the world's aesthetic and the gather/craft loop will be different
The game is called Westpunk because looks like rural Northern California
...
I mean obviously it doesn't... but all of this art is not final lol
...
Biomes will include redwoods, beaches, fields, marshes, and sand dunes
Which are all places I've played as a kid and I want to focus on that

The gather/craft loop will be slow and frustrating
but magically automatable
I hope starting with automation in mind will help solve power creep later

There will be bosses and cinematic elements but not a set story
Achievements will be used by RNG to create a hero's journey for the player
Which should be especially interesting in multiplayer
I also want to create a vertically-mixed adaptive soundtrack
that takes cues from the achievement heroes journey system
I think difficulty might also be customized by the pacing but ehhhhh idk

I hope to release a video every month 
because I'm a college student and free time is a myth
I have my code freely available on github
and playable executables on itch.io
that are a little scuffed but definitely playable

something something like bell subscribe BYEl