package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"strconv"
	"strings"
)

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

func verifyImgSize(img image.PalettedImage, size image.Point) error {
	imgSz := img.Bounds().Size()
	if !size.Eq(imgSz) {
		return fmt.Errorf("Input Image must be %vx%v", size.X, size.Y)
	}
	return nil
}

func pico8(args []string) error {
	img, _, err := loadPalettedImage(args[0])
	if err != nil {
		return err
	}

	err = verifyImgSize(img, image.Pt(128, 128))
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

	err = verifyImgSize(img, image.Pt(128, 128))
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

	img := createPalettedImage(pal, image.Point{128, 128})
	err = savePNG(img, args[1])
	return err
}

func mkgif(args []string) error {
	pal, err := loadHexPalette(args[0])
	if err != nil {
		return err
	}

	img := createPalettedImage(pal, image.Point{128, 128})
	err = saveGIF(img, args[1])
	return err
}

func img2idx(args []string) error {
	src, err := loadImage(args[0])
	if err != nil {
		return err
	}

	pal, err := loadHexPalette(args[1])
	if err != nil {
		return err
	}
	size := src.Bounds().Size()
	dst := createPalettedImage(pal, size)

	draw.Draw(dst.(*image.Paletted), src.Bounds(), src, image.Point{}, draw.Src)

	return savePNG(dst, args[2])
}

type cmdHandler func(args []string) error

type cmd struct {
	name string
	cmdHandler
	desc     string
	nargs    int
	argsdesc string
}

var commands = []cmd{
	{"png", mkpng, "create png file from hex palette", 2,
		"<palette.hex> <output.png>"},
	{"gif", mkgif, "create gif file from hex palette", 2,
		"<palette.hex> <output.gif>"},
	{"img2idx", img2idx, "convert input image to indexed colour using supplied hex palette", 3,
		"<input.[png|gif]> <palette.hex> <output.[png.gif]>"},
	{"xpalhex", xpalhex, "export palette as 32bit hex values", 1,
		"<input.[png|gif]>"},
	{"xpalhexc", xpalhexc, "export palette as 32bit hex values as C code", 1,
		"<input.[png|gif]>"},
	{"xpallua", xpallua, "export palette as {r,g,b} values as lua code", 1,
		"<input.[png|gif]>"},
	{"pico8", pico8, "export pixel data as pico8 sprite data", 1,
		"<input.[png|gif]>"},
	{"tac08", tac08, "export pixel data as tac08 extended sprite data", 1,
		"<input.[png|gif]>"},
}

func getCommand(name string) *cmd {
	for i := 0; i < len(commands); i++ {
		if name == commands[i].name {
			return &commands[i]
		}
	}
	return nil
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

	for i := 0; i < len(commands); i++ {
		fmt.Printf("%12s : %s\n", commands[i].name, commands[i].desc)
		fmt.Printf("%12s   %s %s %s\n", "", "imgtool", commands[i].name, commands[i].argsdesc)
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

	pal = trimPalette(pal)

	return pimg, pal, nil
}

func createPalettedImage(pal color.Palette, size image.Point) image.PalettedImage {

	img := image.NewPaletted(image.Rectangle{Max: size}, pal)

	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
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

func saveGIF(img image.PalettedImage, name string) error {
	w, err := os.Create(name)
	if err != nil {
		return err
	}

	err = gif.Encode(w, img, nil)
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
		p[i] = color.RGBA{0, 0, 0, 0xff}
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

	cmd := getCommand(command)
	if cmd != nil {
		args := os.Args[2:]
		if len(args) != cmd.nargs {
			abend(fmt.Sprintf("invalid args:  %s %s %s\n", "imgtool", command, cmd.argsdesc))
		}
		err := cmd.cmdHandler(args)
		exitOnError(err)
	} else {
		abend("command: " + command + " not found")
	}
}
