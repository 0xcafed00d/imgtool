package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/png"
	"os"
	"strconv"
	"strings"
)

type cmdHandler func(args []string) error

type cmd struct {
	cmdHandler
	desc string
}

func xpalhex(args []string) error {
	_, pal, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}
	for _, c := range pal {
		r, g, b := toRGB(c)
		fmt.Printf("0x%06x\n", uint32(r)<<16|uint32(g)<<8|uint32(b))
	}
	return nil
}

func xpalhexc(args []string) error {
	_, pal, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("const size_t palette_sz = %d;\n", len(pal))
	fmt.Println("uint32_t palette[palette_sz] = {")

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

func xpallua(args []string) error {
	_, pal, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}

	fmt.Println("palette = {")

	for i, c := range pal {
		r, g, b := toRGB(c)
		fmt.Printf("\t{%d, %d, %d}", r, g, b)
		if i < len(pal)-1 {
			fmt.Println(",")
		}
	}
	fmt.Println("\n}")
	return nil
}

func pico8(args []string) error {
	img, _, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}

	fmt.Println("__gfx__")

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x += 2 {
			pixel1 := img.ColorIndexAt(x, y)
			pixel2 := img.ColorIndexAt(x+1, y)
			fmt.Printf("%02x", (pixel1<<4)|(pixel2&0xf))
		}
		fmt.Println("")
	}
	return nil
}

func tac08(args []string) error {
	img, _, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}

	fmt.Println("__gfx8__")

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			pixel := img.ColorIndexAt(x, y)
			fmt.Printf("%02x", pixel)
		}
		fmt.Println("")
	}
	return nil
}

func mkpng(args []string) error {
	pal, err := loadHexPalette(args[0])
	if err != nil {
		return err
	}

	img := createPalettedImage(pal)
	err = savePNG(img, args[1])
	return err
}

func mkgif(args []string) error {
	return nil
}

func rgb2idx(args []string) error {
	return nil
}

var commands = map[string]cmd{
	"png":      {mkpng, "create png file from hex palette"},
	"gif":      {mkgif, "create gif file from hex palette"},
	"rgb2idx":  {rgb2idx, "convert input rgb image to indexed colour using supplied hex palette"},
	"xpalhex":  {xpalhex, "export palatte as 32bit hex values"},
	"xpalhexc": {xpalhexc, "export palatte as 32bit hex values as C code"},
	"xpallua":  {xpallua, "export palatte as {r,g,b} values as lua code"},
	"pico8":    {pico8, "export pixel data as pico8 sprite data"},
	"tac08":    {tac08, "export pixel data as tac08 extended sprite data"},
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
	_r, _g, _b := toRGB(c)
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

func loadImage(name string) (image.Image, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func loadPalettedImage(name string) (image.PalettedImage, color.Palette, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, nil, err
	}

	pimg, ok := img.(image.PalettedImage)
	if !ok {
		return nil, nil, fmt.Errorf("Not a palettised image: %s", name)
	}

	pal, ok := img.ColorModel().(color.Palette)
	if !ok {
		return nil, nil, fmt.Errorf("Not a palettised image: %s", name)
	}

	return pimg, pal, nil
}

func createPalettedImage(pal color.Palette) image.PalettedImage {
	w, h := 128, 128

	img := image.NewPaletted(image.Rect(0, 0, w, h), pal)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetColorIndex(x, y, uint8(((x*2)>>4)+((y*2)&0xf0)))
		}
	}
	return img
}

func savePNG(img image.PalettedImage, name string) error {

	w, err := os.Create(name)
	if err != nil {
		return err
	}

	err = png.Encode(w, img)
	return err
}

func loadHexPalette(name string) (color.Palette, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	p := make(color.Palette, 256)
	for i := 0; i < len(p); i++ {
		p[i] = color.RGBA{0, 0, 0, 0}
	}

	line := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := scanner.Text()
		l = strings.Replace(l, "0x", "", -1)
		n, err := strconv.ParseUint(l, 16, 32)
		if err != nil {
			return nil, err
		}
		p[line] = color.RGBA{uint8((n >> 16) & 0xff), uint8((n >> 8) & 0xff), uint8(n & 0xff), 0xff}
		line++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: imagetool <command> <command args>")
		listCommands()
		os.Exit(-1)
	}

	command := os.Args[1]

	if cmd, ok := commands[command]; ok {
		err := cmd.cmdHandler(os.Args[2:])
		exitOnError(err)
	} else {
		abend("command: " + command + " not found")
	}
}
