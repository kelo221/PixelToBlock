package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/cenkalti/dominantcolor"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Pixel struct {
	R int
	G int
	B int
	A int
}

var filename string

func main() {

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)

	//	Save and load the data
	handleBlockData()


	//	Handle Input image
	filename = getFilename()

	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	e := os.Remove(filename + ".mcfunction")
	if e != nil {
		println("No existing mcfunction file found")
	}else {
		println("Old mcfunction file removed")
	}


	// Convert chosen image to an array
	pixels, err := imageToArray(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}


	// Check the closest match; match the most dominant pixel (and its block data) to the one from the image
	calculateMatch(pixels)

	err2 := file.Close()
	if err2 != nil {
		return
	}

	print("Success.")

}


func handleBlockData()  {

	if _, err := os.Stat("blockdata.txt"); err == nil {
		fmt.Printf("Block data found!\n")

	} else if os.IsNotExist(err) {
		fmt.Printf("Block not data found.\n")
		fmt.Printf("Trying to generate one...\n")

		currentDirectory, err := os.Getwd()
		currentDirectory += "/block"
		if err != nil {
			log.Fatal(err)
		}
		generateBlockData()
		fmt.Printf("Success. \n")
	} else {
		print(err)
		os.Exit(99)
	}
}


var errInvalidFormat = errors.New("invalid format")

func FindDomiantColor(fileInput string) (c color.RGBA, err error) {
	f, err := os.Open(fileInput)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)
	if err != nil {
		fmt.Println("File not found:", fileInput)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return c, nil
	}

	s:=  dominantcolor.Hex(dominantcolor.Find(img))

	// Hex format, so has to be converted into RGB

	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat

	}
	return
}

func generateBlockData(){

	inputPattern := "blocks/*.png"
	files, err := filepath.Glob(inputPattern)

	if nil != err {
		log.Println(err)
		log.Println("Error: failed glob")
		return
	}

	for _, file := range files {
		cols, err := FindDomiantColor(file)
		if err != nil {
			println(file)
			log.Println(err)
			continue
		}
		println(cols.R, cols.G, cols.B, file)
		saveBlockData(uint32(cols.R), uint32(cols.G), uint32(cols.B), file)
	}

}

func saveBlockData(red uint32, blue uint32,green uint32,  filename string){


	// Strip file extension and path
	filenameNoExt := filename[:len(filename)-4]
	filenameNoExt = filenameNoExt[7:]

	colordata := strconv.Itoa(int(red)) + " " + strconv.Itoa(int(blue)) + " " + strconv.Itoa(int(green)) + " " + "minecraft:" + filenameNoExt

	f, err := os.OpenFile("blockdata.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)
	if _, err := f.WriteString(colordata+"\n"); err != nil {
		log.Println(err)
	}
}

func getFilename() string {

	fmt.Println("Enter filename for an image file: ")

	var filename string

	n, err := fmt.Scanln(&filename)
	if err != nil {
		fmt.Println(n, err)
		return "File not found"
	}
	return filename
}


func imageToArray(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if img == nil {
		return nil, err
	}

	bounds := img.Bounds()
	fmt.Printf("Current image height: %d \n", bounds.Max.Y)
	fmt.Printf("Would you like to scale it? Input new height, or 0 to skip.\n")

	var newheight uint
	_, err = fmt.Scan(&newheight)
	if err != nil {

	}

	if  newheight != 0{
		img = resize.Thumbnail(999, newheight, img, resize.Bicubic)
	}

	bounds = img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

func mcFunctiongenerator(command string)  {

	filenameNoExt := filename[:len(filename)-4]

	f, err := os.OpenFile( filenameNoExt + ".mcfunction",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
		}
	}(f)
	if _, err := f.WriteString(command+"\n"); err != nil {
		log.Println(err)
	}

}

func calculateMatch(pixels [][]Pixel)  {


	var option int

	print("Which direction do you want the blocks be facing? \n" +
		"1. West\n" +
		"2. North\n" +
		"3. Ground\n")

	_, err := fmt.Scanf("%d", &option)
	if err != nil {
		println("Input error")
		return
	}

	var temp [][]Pixel

	// Rotate matrix clockwise 90 degrees

	for y := 0; y<len(pixels[1]); y++ {
		var row []Pixel
		for x := len(pixels)-1;  x>=0; x-- {
			row = append(row,pixels[x][y])
		}
		temp = append(temp, row)
	}

	pixels = temp

	for i := range pixels {
		for j :=range pixels[1] {

			if pixels[i][j].A<70 {
				// If wanted pixels with low alpha can be set to air, instead of ignoring, but this will increase the command count
				//command := "setblock ~" + strconv.FormatInt(int64(i), 10) + " ~" + strconv.FormatInt(int64(j), 10) + " ~0 minecraft:air"
				//mcFunctiongenerator(command)
			}else {
				var currentblock string
				var chosenblock string
				difference := 0.0
				differenceCurrent := 999.0

				file, err := os.Open("blockdata.txt")
				if err != nil {
					log.Fatal(err)
				}
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {

					}
				}(file)

				scanner := bufio.NewScanner(file)

				for scanner.Scan() {

					rawdata := scanner.Text()
					split := strings.Split(rawdata, " ")
					tempR, err2 := strconv.Atoi(split[0])
					if err2 != nil {
						print("Block file corrupted")
						log.Fatal(err)
					}

					tempG, _ := strconv.Atoi(split[1])
					tempB, _ := strconv.Atoi(split[2])
					currentblock = split[3]

					mean := (pixels[i][j].R - tempR) /2
					tempR = pixels[i][j].R - tempR
					tempG = pixels[i][j].G - tempG
					tempB = pixels[i][j].B - tempB


					Rweight := 512
					Bweight := 767

					difference = math.Sqrt(float64((((Rweight + mean) * tempR * tempR) >> 8) + 4*tempG*tempG + (((Bweight - mean) * tempB * tempB) >> 8)))

					if difference<differenceCurrent {
						differenceCurrent = difference
						chosenblock = currentblock
					}
				}


				//setblock ~x ~y ~z minecraft:light_blue_glazed_terracotta
				// x=facing west z=north

				//west
				if option==1 {
					command := "setblock ~" + strconv.FormatInt(int64(i), 10) + " ~" + strconv.FormatInt(int64(j), 10) + " ~0 " + chosenblock
					mcFunctiongenerator(command)
				}
				//north
				if option==2 {
					command := "setblock ~0" + " ~" + strconv.FormatInt(int64(j), 10) + " ~" + strconv.FormatInt(int64(i), 10) +" " + chosenblock
					mcFunctiongenerator(command)
				}
				//ground
				if option==3 {
					command := "setblock ~" + strconv.FormatInt(int64(i), 10) + " ~0"  + " ~" + strconv.FormatInt(int64(j), 10) +" " + chosenblock
					mcFunctiongenerator(command)
				}

				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

