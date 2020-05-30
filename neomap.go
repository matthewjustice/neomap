package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Ensure that enough command line args were passed
	if len(os.Args) < 6 {
		printHelp()
		return
	}

	// Assign and verify command line args
	exePath := os.Args[1]
	neoGeoA := os.Args[2]
	neoGeoB := os.Args[3]
	neoGeoC := os.Args[4]
	neoGeoD := os.Args[5]

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

	// Patch the file
	// First try the byte offset and values for the GOG and Amazon releases
	fmt.Println("\nTrying patch for GOG and Amazon releases...")
	if !patchFile(exePath, mapping, [4]int{0x239D0, 0x239D4, 0x239D8, 0x239DC}, [4]byte{0x15, 0x1c, 0x23, 0x2a}) {
		// if that didn't work, try again with the offsets for the Humble Bundle release
		fmt.Println("\nTrying patch for Humble releases...")
		patchFile(exePath, mapping, [4]int{0x82F4, 0x82F8, 0x82FC, 0x8300}, [4]byte{0x28, 0x2f, 0x36, 0x3d})
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

func patchFile(exePath string, mapping []int, xboxButtonJumpTableOffsets [4]int, neoGeoHandlerBytes [4]byte) bool {
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
	// the size of the last offset +1 since we read from that offset.
	if len(byteData) < (xboxButtonJumpTableOffsets[3] + 1) {
		fmt.Printf("File \"%s\" is too small for this patch. It is %d bytes.\n", exePath, len(byteData))
		return false
	}

	// Check byte values before we change them
	bytesOK := true
	for i := 0; i < 4; i++ {
		checkByte := byteData[xboxButtonJumpTableOffsets[i]]
		if checkByte != neoGeoHandlerBytes[i] {
			fmt.Printf("byte at 0x%08x is 0x%02x, expected 0x%02x\n", xboxButtonJumpTableOffsets[i], checkByte, neoGeoHandlerBytes[i])
			bytesOK = false
			break
		}
	}

	if !bytesOK {
		fmt.Printf("File \"%s\" had an unexpected byte value.\n", exePath)
		return false
	}

	// Each index in the mapping array represents the Xbox button number
	// and each value represents the NeoGeo button number
	for xbox, neogeo := range mapping {
		byteData[xboxButtonJumpTableOffsets[xbox]] = neoGeoHandlerBytes[neogeo]
		fmt.Printf("byte at 0x%08x updated to 0x%02x\n", xboxButtonJumpTableOffsets[xbox], neoGeoHandlerBytes[neogeo])
	}

	// Get the output file name
	outputFileName := generateOutputFilename(exePath)
	fmt.Printf("\nWriting patched file to \"%s\"...\n", outputFileName)

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
	fmt.Println("neomap remaps controller buttons in DotEmu's Neo Geo games for Windows.")
	fmt.Println()
	fmt.Println("syntax: neomap.exe [exePath] [A] [B] [C] [D]")
	fmt.Println("  where")
	fmt.Println("  - exePath is the full path to the DotEmu NeoGeo executable to patch.")
	fmt.Println("  - A, B, C, D are the Xbox controller button letters you want to map")
	fmt.Println("    to NeoGeo buttons A, B, C, D. Valid Xbox buttons are: A, B, X, Y")
	fmt.Println()
	fmt.Println("For example, let's say you want to update King of Fighter 2002 as follows:")
	fmt.Println("  - Xbox X button is mapped to NeoGeo A button")
	fmt.Println("  - Xbox A button is mapped to NeoGeo B button")
	fmt.Println("  - Xbox Y button is mapped to NeoGeo C button")
	fmt.Println("  - Xbox B button is mapped to NeoGeo D button")
	fmt.Println("Then run the tool like so:")
	fmt.Println("neomap.exe c:\\path\\KingOfFighters2002.exe X A Y B")
	fmt.Println()
	fmt.Println("This will attempt to write a new, patched exe file to the same folder as the")
	fmt.Println("original game. Your original exe file won't be modified.")
}
