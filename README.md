# westpunk
https://hoping.itch.io/westpunk<br>
A two-dimensional sandbox survival game with fighting mechanics inspired by fighting games, especially Brawlhalla. 
The gather/craft loop will be limited and slow, with a focus on realism. However, there will be ways to mechanically or magically automate parts of the process. 
Characters are drawn as 2D rigged models, making it easy to change skins and add armor and clothing, as well as supporting the fighting-game aspect.
The world generates infinitely in the form of a 2D map of finite-area locations you can visit, helping the world feel more 3D.
<br><br>
The game is written in Go and uses the Ebiten graphics library (https://ebiten.org/) and no game engine.
You can run the game by running the proper executable in bin/, from the src/westpunk directory.
If this doesn't work, there are zip files in src/westpunk that, when extracted into an empty folder, should run the game as well.
If *this* doesn't work and you have Go installed, you can go into the src/westpunk directory and run `go install`, 
then go to your Go bin and run the westpunk executable created.
Note: your Go bin will likely not be in westpunk/bin!
<br><br>
Conventions:
 - Modularity is from the Go module system, instead of classes
 - The code does not abide by OOP principles, intentionally. It is organized into Go modules, which are internally data-oriented in general. There is room for improvement in this area, however.
 - Four space indentation
 - Opening curly braces don't get their own line. You know what I'm talking about
 - Comments at the end of a line of code have one space between the code and the double-slash, and one space between the double-slash and the commented text
 - File names, unexported identifiers, and modules names are all lowercase, with underscores
 - Exported identifiers have first letters of each word capitalized, no spaces
 - For the most part, instead of methods, the code uses functions that take pointers
 - IDs are enums, used as keys for hash tables, which are used a lot :)
