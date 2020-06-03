package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type jumpTable struct {
	offset int
	values [4]uint32
}

func main() {
	// Ensure that enough command line args were passed
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	inputFilePath := os.Args[1]
	ext := filepath.Ext(inputFilePath)

	csvMode := false
	if (len(os.Args) >= 3) && (os.Args[2] == "csv") {
		csvMode = true
		fmt.Println("Filename,File size,File modified,Table,Jump A,Jump B,Jump C,Jump D,Install path")
	}

	if ext == ".txt" {
		// Assume the input is a text file of exe paths to process
		file, err := os.Open(inputFilePath)
		if err != nil {
			fmt.Printf("Unable to open text file %s\n", inputFilePath)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		// Process each line as an input exe file
		for scanner.Scan() {
			examineExecutableFile(scanner.Text(), csvMode)
		}
	} else {
		// Assume we've been given a single binary exe to examine
		examineExecutableFile(inputFilePath, csvMode)
	}
}

func examineExecutableFile(exePath string, csvMode bool) {
	dir, file := filepath.Split(exePath)

	if csvMode {
		fmt.Printf("%s,", file)
	} else {
		fmt.Printf("\nProcessing %s\n", exePath)
	}

	// Does the file exist?
	if !fileExists(exePath) {
		fmt.Printf("\"%s\" does not exist.\n", exePath)
		return
	}

	//Read bytes of the exe file
	byteData, err := ioutil.ReadFile(exePath)
	if err != nil {
		fmt.Printf("Could not read file \"%s\", error = %s\n", exePath, err)
		return
	}

	if csvMode {
		fmt.Printf("%d,", len(byteData))
		fmt.Printf("%s,", fileModified(exePath))
	} else {
		fmt.Printf("File size = %d\n", len(byteData))
		fmt.Printf("File modified = %s\n", fileModified(exePath))
	}

	knownJumpTables := []jumpTable{
		jumpTable{
			offset: 0x000082F4, // Humble Bundle release
			values: [4]uint32{0x00408e28, 0x00408e2f, 0x00408e36, 0x00408e3d},
		},
		jumpTable{
			offset: 0x000239D0, // GOG and Amazon releases
			values: [4]uint32{0x00424515, 0x0042451C, 0x00424523, 0x0042452A},
		},
	}

	// Check for the known jump tables in this file.
	for _, table := range knownJumpTables {
		readValues := readUInt32ArrayLittleEndian(byteData, table.offset, 4)

		if readValues[0] == table.values[0] &&
			readValues[1] == table.values[1] &&
			readValues[2] == table.values[2] &&
			readValues[3] == table.values[3] {

			// If the values match a known jump table, print it
			if csvMode {
				fmt.Printf(
					"0x%08x,0x%08x,0x%08x,0x%08x,0x%08x,",
					table.offset,
					readValues[0],
					readValues[1],
					readValues[2],
					readValues[3])
			} else {
				fmt.Printf(
					"Data at %08x is %08x, %08x, %08x, %08x\n",
					table.offset,
					readValues[0],
					readValues[1],
					readValues[2],
					readValues[3])
			}
		}
	}

	if csvMode {
		fmt.Printf("%s\n", dir)
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func fileModified(filePath string) string {
	const dateLayout = "01/02/2006"
	fileInfo, err := os.Stat(filePath)
	var timeString = ""
	if err == nil {
		timeString = fileInfo.ModTime().Format(dateLayout)
	}
	return timeString
}

func printHelp() {
	fmt.Println("Usage: fileinfo.exe [file_to_process] [csv]")
	fmt.Println("file_to_process is either an exe file or a text file with one exe path per line.")
	fmt.Println("The csv argument is optional.")
}

func readUInt32LittleEndian(data []byte, start int) uint32 {
	var result uint32 = 0

	// We'll read data[start] through data[start+3]
	// make sure data is large enough
	lastByte := start + 3
	if len(data) < (lastByte + 1) {
		result = 0
	} else {
		result = uint32(data[start]) | uint32(data[start+1])<<8 | uint32(data[start+2])<<16 | uint32(data[start+3])<<24
	}

	return result
}

func readUInt32ArrayLittleEndian(data []byte, start int, count int) []uint32 {
	result := make([]uint32, count)
	for i := 0; i < count; i++ {
		result[i] = readUInt32LittleEndian(data, start+(i*4))
	}

	return result
}
