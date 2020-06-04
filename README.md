# neomap

Neomap allows you to remap your controller buttons on Dotemu's ports of Neo Geo games for Windows.

## Getting Started

Here we'll cover what you need to use neomap. 

### Introduction

Neomap applies to various Dotemu releases of Neo Geo game for Windows. **The problem** it addresses: these games don't include a way to remap your controller buttons; the mapping are are hard-coded. **The fix**: neomap works by modifying the button mapping within the game's executable code. Neomap doesn't change the original exe file on disk; it creates a new exe with updated code.

### Identifying a Game to Update

First, you need one of the games in question; one of the Dotemu-released PC ports of classic Neo Geo titles. These have been included with the 
[Humble NEOGEO 25th Anniversary Bundle](https://www.pcmag.com/news/dont-miss-the-neogeo-25th-anniversary-humble-bundle) (in 2015), as separate 
[titles on GOG](https://www.gog.com/news/release_15_snk_neogeo_classics) (since 2017), and most recently as a 
[free giveaway](https://www.gamespot.com/articles/amazon-prime-subscribers-can-get-16-games-for-free/1100-6476703/) to Amazon Twitch Prime members (in 2020). 

These patches do **not** apply (nor are they needed) to the Steam releases of Neo Games, such as the Code Mystics port of *Garou: Mark of the Wolves*.

To update your game, you need to find the executable file (*.exe file). By default, on 64-bit Windows, these games install to one of the following locations:
- `C:\Program Files (x86)\NeoGeo 25th Anniversery` *(Humble Bundle releases)*
- `C:\GOG Games` *(GOG releases)*
- `C:\Amazon Games\Library` *(Amazon Twitch releases)*

For example, if you have *King of Fighters 2002*, the exe file you need to update will probably be in one of these locations, depending on the version of the game (Humble, GOG, or Amazon):
- `C:\Program Files (x86)\NeoGeo 25th Anniversery\KingOfFighters2002\KingOfFighters2002.exe`
- `C:\GOG Games\The King of Fighters 2002\KingOfFighters2002.exe`
- `C:\Amazon Games\Library\The King of Fighters 2002\KingOfFighters2002.exe`

### Determine Your Desired Button Mapping
Assuming you are using an Xbox controller, let's say you want to map your Xbox buttons to the virtual Neo Geo buttons like so:

- Xbox X button is mapped to Neo Geo A button
- Xbox A button is mapped to Neo Geo B button
- Xbox Y button is mapped to Neo Geo C button
- Xbox B button is mapped to Neo Geo D button

Your button mapping configuration, for the above example, will be `X A Y B`. This reflects the Xbox buttons you want to use for Neo Geo buttons A, B, C, and D, in that order.

### How to Update a Game

Once you know the location of the executable file for your game and your desired button mapping, you can update the game.

1. Download the latest version of neomap.
2. Extract the zip file contents, for example to `c:\temp\neomap.exe`.
3. Open a command prompt window. 
NOTE: You many to run as administrator to update games under `Program Files`.
4. Change directories to the folder from step 2. For example:
    ```
    C:\>cd \temp
    ```
5. Run `neomap.exe`, specifying your preferred button layout and the path to the game's exe file. The path to the exe file should be in quotes if there are spaces in the path. For example:
    ```
    C:\temp\>neomap.exe X A Y B "C:\GOG Games\The King of Fighters 2002\KingOfFighters2002.exe"
    ```
6. Expect to see output like the following:
    ```
    Writing patched file to "C:\GOG Games\The King of Fighters 2002\KingOfFighters2002-remap-1591274072.exe"
    File written successfully.
    ```
7. The original game exe file will be unmodified. To play the patched version with remapped buttons, run the exe file specified in the previous steps (`KingOfFighters2002-remap-1591274072.exe` in this case).
8. Once you are satisfied that the updated version works, you can back up your original, unmodified exe file, and replace it with the modified version. This is optional.
9. If running `neomap.exe` gave you access denied errors, try again from an elevated command prompt (run as administrator).

### Disclaimer

Neomap hasn't been extensively tested. It was written as a fun side-project and made available in hopes of helping others who want to remap their buttons. By running it on your system, you assume any associated risk!

## Developer Information

neomap is currently intended for Windows only.

To build neomap:
1. [Download Go](https://golang.org/dl/) and install.
2. Run with: `go run neomap.go`
3. Build 32-bit: 
    ```
    set GOARCH=386
    go build neomap.go
    ```

## Built With

- [Go](https://golang.org/)


## Authors

- **Matthew Justice** [matthewjustice](https://github.com/matthewjustice)

See also the list of [contributors](https://github.com/matthewjustice/pumpkinpi/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
