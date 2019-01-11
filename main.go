package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"os"
)

type cmdHandler func(img image.PalettedImage, pal color.Palette) error

type cmd struct {
	cmdHandler
	desc string
}

func xpalhex(img image.PalettedImage, pal color.Palette) error {
	for _, c := range pal {
		r, g, b := toRGB(c)
		fmt.Printf("0x%06x\n", uint32(r)<<16|uint32(g)<<8|uint32(b))
	}

	return nil
}

var commands = map[string]cmd{
	"xpalhex": {xpalhex, "export palatte as 32bit hex values"},
}

func exitOnError(e error) {
	if e != nil {
		abend(e.Error())
	}
}

func abend(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(-1)
}

func toRGB(c color.Color) (r uint8, g uint8, b uint8) {
	_r, _g, _b, _ := c.RGBA()
	r = uint8(_r / 0x101)
	g = uint8(_g / 0x101)
	b = uint8(_b / 0x101)
	return
}

func zeroRGB(c color.Color) bool {
	_r, _g, _b, _ := c.RGBA()
	return _r == 0 && _g == 0 && _b == 0
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
	if !ok {
		abend("Not a palletised image")
	}

	pal, ok := img.ColorModel().(color.Palette)
	if !ok {
		abend("Not a palletised image")
	}

	if cmd, ok := commands[command]; ok {
		err := cmd.cmdHandler(pimg, pal)
		exitOnError(err)
	} else {
		abend("command: " + command + " not found")
	}
}
