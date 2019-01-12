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

func xpalhexc(img image.PalettedImage, pal color.Palette) error {
	fmt.Println("uint32_t palette[] = {")

	for i, c := range pal {
		r, g, b := toRGB(c)
		fmt.Printf("\t0x%06x", uint32(r)<<16|uint32(g)<<8|uint32(b))
		if i < len(pal)-1 {
			fmt.Println(",")
		}
	}
	fmt.Println("\n};")

	return nil
}

func pico8(img image.PalettedImage, pal color.Palette) error {
	return nil
}

var commands = map[string]cmd{
	"xpalhex":  {xpalhex, "export palatte as 32bit hex values"},
	"xpalhexc": {xpalhexc, "export palatte as 32bit hex values as C code"},
	"pico8":    {pico8, "export pixel data as pico8 sprite data"},
	"tac08":    {pico8, "export pixel data as tac08 extended sprite data"},
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

func isZeroRGB(c color.Color) bool {
	_r, _g, _b, _ := c.RGBA()
	return _r == 0 && _g == 0 && _b == 0
}

func trimPalette(pal color.Palette) color.Palette {
	for len(pal) > 0 && isZeroRGB(pal[len(pal)-1]) {
		pal = pal[:len(pal)-1]
	}
	return pal
}

func listCommands() {
	fmt.Printf("Commands:\n")

	for k, v := range commands {
		fmt.Printf("%12s : %s\n", k, v.desc)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: imagetool command imagefile")
		listCommands()
		os.Exit(-1)
	}

	command := os.Args[1]
	f := os.Args[2]

	file, err := os.Open(f)
	exitOnError(err)
	defer file.Close()

	img, _, err := image.Decode(file)
	exitOnError(err)

	pimg, ok := img.(image.PalettedImage)
	if !ok {
		abend("Not a palletised image")
	}

	pal, ok := img.ColorModel().(color.Palette)
	if !ok {
		abend("Not a palletised image")
	}

	if cmd, ok := commands[command]; ok {
		err := cmd.cmdHandler(pimg, trimPalette(pal))
		exitOnError(err)
	} else {
		abend("command: " + command + " not found")
	}
}
