package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"os"
)

func exitOnError(e error) {
	if e != nil {
		abend(e.Error())
	}
}

func abend(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(-1)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: imagetool command imagefile")
		os.Exit(-1)
	}

	command := os.Args[1]
	f := os.Args[2]

	_ = command

	file, err := os.Open(f)
	exitOnError(err)
	defer file.Close()

	img, format, err := image.Decode(file)
	exitOnError(err)

	fmt.Println(format, " ", img.Bounds())
	pimg, ok := img.(image.PalettedImage)
	if ok {
		_ = pimg
	} else {
		abend("Not a palletised image")
	}

	pal, ok := img.ColorModel().(color.Palette)
	if ok {
		for n := 0; n < len(pal); n++ {
			r, g, b, _ := pal[n].RGBA()
			println("pal ", n, " = ", r/0x101, g/0x101, b/0x101)

		}
	} else {
		abend("Not a palletised image")
	}

}
