package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type patchData struct {
	offset int
	values [4]byte
	label  string
}

func main() {
	// Ensure that enough command line args were passed
	if len(os.Args) < 6 {
		printHelp()
		return
	}

	// Assign and verify command line args
	neoGeoA := os.Args[1]
	neoGeoB := os.Args[2]
	neoGeoC := os.Args[3]
	neoGeoD := os.Args[4]
	exePath := os.Args[5]

	fmt.Println()
	fmt.Println("Checking inputs...")
	fmt.Printf("\"%s\" is the exe to patch\n", exePath)
	fmt.Printf("Xbox button %s will be mapped to NeoGeo button A\n", neoGeoA)
	fmt.Printf("Xbox button %s will be mapped to NeoGeo button B\n", neoGeoB)
	fmt.Printf("Xbox button %s will be mapped to NeoGeo button C\n", neoGeoC)
	fmt.Printf("Xbox button %s will be mapped to NeoGeo button D\n", neoGeoD)

	buttonsValid := buttonIsValid(neoGeoA) && buttonIsValid(neoGeoB) && buttonIsValid(neoGeoC) && buttonIsValid(neoGeoD)

	if !buttonsValid {
		fmt.Println("Invalid button configuration. Exiting.")
		return
	}

	mapping := makeMappingArray(neoGeoA, neoGeoB, neoGeoC, neoGeoD)

	patches := []patchData{
		patchData{
			offset: 0x000082F4, // Humble Bundle release
			values: [4]byte{0x28, 0x2f, 0x36, 0x3d},
			label:  "Humble Bundle",
		},
		patchData{
			offset: 0x000239D0, // GOG and Amazon releases
			values: [4]byte{0x15, 0x1C, 0x23, 0x2A},
			label:  "GOG and Amazon/Twitch",
		},
		patchData{
			offset: 0x00008134, // Humble Bundle - Twinkle Star Sprites (and maybe older versions of other titles?)
			values: [4]byte{0x68, 0x6F, 0x76, 0x7D},
			label:  "Humble Bundle (older build)",
		},
	}

	for _, patch := range patches {
		fmt.Printf("\nTrying patch for %s release...\n", patch.label)
		if patchFile(exePath, mapping, patch.offset, patch.values) {
			// success - we can exit the loop
			break
		}
	}
}

func buttonIsValid(letter string) bool {
	valid := (strings.EqualFold(letter, "A") || strings.EqualFold(letter, "B") || strings.EqualFold(letter, "X") || strings.EqualFold(letter, "Y"))
	if !valid {
		fmt.Printf("Xbox button %s is invalid (use A, B, X, or Y)", letter)
	}
	return valid
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func makeMappingArray(neoGeoA string, neoGeoB string, neoGeoC string, neoGeoD string) []int {
	// Create an array where each index 0-3 represents an Xbox
	// button (index 0 is Xbox button A, etc.) and each
	// value represents a NeoGeo button (value 0 is NeoGeo A, etc.)
	mapping := make([]int, 4)
	neoGeoToXboxArray := [4]string{neoGeoA, neoGeoB, neoGeoC, neoGeoD}

	for i, xboxButton := range neoGeoToXboxArray {
		switch strings.ToUpper(xboxButton) {
		case "A":
			mapping[0] = i
		case "B":
			mapping[1] = i
		case "X":
			mapping[2] = i
		case "Y":
			mapping[3] = i
		}
	}

	return mapping
}

func generateOutputFilename(inputFilePath string) string {
	dir, file := filepath.Split(inputFilePath)
	ext := filepath.Ext(file)
	var fileNoExt string
	if ext == "" {
		fileNoExt = file
	} else {
		fileNoExt = strings.TrimSuffix(file, ext)
	}

	// The ouput file is the original filename with -remap added and
	// a number added to the end that will keep in unique
	outputFile := fmt.Sprintf("%s-remap-%d.exe", fileNoExt, time.Now().Unix())
	return filepath.Join(dir, outputFile)
}

func patchFile(exePath string, mapping []int, xboxButtonJumpTableOffset int, neoGeoHandlerBytes [4]byte) bool {
	// Does the file exist?
	if !fileExists(exePath) {
		fmt.Printf("\"%s\" does not exist.\n", exePath)
		return false
	}

	// Read bytes of the exe file
	byteData, err := ioutil.ReadFile(exePath)
	if err != nil {
		fmt.Printf("Could not read file \"%s\", error = %s\n", exePath, err)
		return false
	}

	// Check file size. The size of the exe varies, but it should be at least
	// the size of xboxButtonJumpTableOffset + 0xC + 1
	// since we read from the +0xC offset.
	if len(byteData) < (xboxButtonJumpTableOffset + 0xD) {
		fmt.Printf("File \"%s\" is too small for this patch. It is %d bytes.\n", exePath, len(byteData))
		return false
	}

	// Check byte values before we change them
	bytesOK := true
	for i := 0; i < 4; i++ {
		checkByte := byteData[xboxButtonJumpTableOffset+(i*4)]
		if checkByte != neoGeoHandlerBytes[i] {
			fmt.Printf("byte at 0x%08x is 0x%02x, expected 0x%02x\n", xboxButtonJumpTableOffset+(i*4), checkByte, neoGeoHandlerBytes[i])
			bytesOK = false
			break
		}
	}

	if !bytesOK {
		fmt.Printf("File \"%s\" isn't the right version for this patch.\n", exePath)
		return false
	}

	// Each index in the mapping array represents the Xbox button number
	// and each value represents the NeoGeo button number
	for xbox, neogeo := range mapping {
		byteData[xboxButtonJumpTableOffset+(xbox*4)] = neoGeoHandlerBytes[neogeo]
		fmt.Printf("byte at 0x%08x updated to 0x%02x\n", xboxButtonJumpTableOffset+(xbox*4), neoGeoHandlerBytes[neogeo])
	}

	// Get the output file name
	outputFileName := generateOutputFilename(exePath)
	fmt.Printf("\nWriting patched file to \"%s\"\n", outputFileName)

	file, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Could not create file \"%s\", error = %s\n", outputFileName, err)
		return false
	}

	_, err = file.Write(byteData)
	if err != nil {
		fmt.Printf("Could not write file \"%s\", error = %s\n", outputFileName, err)
		return false
	}

	fmt.Println("File written successfully.")
	file.Close()

	return true
}

func printHelp() {
	fmt.Println()
	fmt.Println("neomap remaps controller buttons in Dotemu's Neo Geo games for Windows.")
	fmt.Println()
	fmt.Println("syntax: neomap.exe [A] [B] [C] [D] [exePath]")
	fmt.Println("  where")
	fmt.Println("  - A, B, C, D are the Xbox controller button letters you want to map")
	fmt.Println("    to NeoGeo buttons A, B, C, D. Valid Xbox buttons are: A, B, X, Y")
	fmt.Println("  - exePath is the full path to the DotEmu NeoGeo executable to patch.")
	fmt.Println()
	fmt.Println("For example, let's say you want to update King of Fighter 2002 as follows:")
	fmt.Println("  - Xbox X button is mapped to NeoGeo A button")
	fmt.Println("  - Xbox A button is mapped to NeoGeo B button")
	fmt.Println("  - Xbox Y button is mapped to NeoGeo C button")
	fmt.Println("  - Xbox B button is mapped to NeoGeo D button")
	fmt.Println("Then run the tool like so:")
	fmt.Println("neomap.exe X A Y B c:\\path\\KingOfFighters2002.exe")
	fmt.Println()
	fmt.Println("This will attempt to write a new, patched exe file to the same folder as the")
	fmt.Println("original game. Your original exe file won't be modified.")
	fmt.Println()
	fmt.Println("neomap build version: 2020.06.04")
}
