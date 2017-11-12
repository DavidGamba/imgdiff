// This file is part of imgdiff.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

/*
imgdiff - check differences between two images of the same size.
*/
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	// Support gif, jpeg and png
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/DavidGamba/go-getoptions"
)

var debug bool
var amplify bool
var colorHighlightCalled bool
var colorHighlight string
var diffCount int
var sizeTotal int
var baseColor uint8

func synopsis() {
	synopsis := `imgdiff <imgA> <imgB> [-o <imgC>] [--reverse] [--base <0-255>]

imgdiff <imgA> <imgB> [-o <imgC>] [--amplify] [--base <0-255>]

imgdiff <imgA> <imgB> [-o <imgC>] [--reverse] [--color [<color>]]

imgdiff [--help]

# Options:

--color [<color>]: Only "red" supported.

--base <0-255>: Base color to diff against. Defaults to 0.
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	var reverse bool
	var outputPath string
	var base int
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.BoolVar(&debug, "debug", false)
	opt.BoolVar(&reverse, "reverse", false)
	opt.BoolVar(&amplify, "amplify", false)
	opt.StringVarOptional(&colorHighlight, "color", "red")
	opt.StringVar(&outputPath, "output", "output.png")
	opt.IntVar(&base, "base", 0)
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("help") {
		synopsis()
		os.Exit(1)
	}
	if !opt.Called("debug") {
		log.SetOutput(ioutil.Discard)
	}
	if opt.Called("color") {
		colorHighlightCalled = true
	}
	log.Println(remaining)
	if len(remaining) < 2 {
		fmt.Fprintf(os.Stderr, "ERROR: Missing arguments!\n")
		synopsis()
		os.Exit(1)
	}
	baseColor = uint8(base)
	imagePath1 := remaining[0]
	imagePath2 := remaining[1]

	err = diffImages(imagePath1, imagePath2, outputPath, reverse)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Total pixels: %d, Diff  pixels: %d, Diff  percentage: %d%%\n", sizeTotal, diffCount, diffCount*100/sizeTotal)
	if diffCount > 0 {
		os.Exit(2)
	}
}

func diffImages(imagePath1, imagePath2, outputPath string, reverse bool) error {
	m1, err := decodeImage(imagePath1)
	if err != nil {
		return err
	}
	m2, err := decodeImage(imagePath2)
	if err != nil {
		return err
	}
	bounds := m1.Bounds()
	bounds2 := m2.Bounds()
	log.Printf("Bounds: %#v, %#v\n", bounds, bounds2)
	if bounds != bounds2 {
		return fmt.Errorf("ERROR: Different image sizes!\n")
	}

	img := getDrawableImage(m1)
	for xy := range pixelChannel(img) {
		sizeTotal++
		img.Set(xy.x, xy.y, diffColor(m1.At(xy.x, xy.y), m2.At(xy.x, xy.y), reverse))
	}

	return writeImageToPNGFile(img, outputPath)
}

type coordinates struct {
	x, y int
}

func pixelChannel(img image.Image) <-chan coordinates {
	c := make(chan coordinates)
	go func() {
		bounds := img.Bounds()
		// From documentation example:
		// An image's bounds do not necessarily start at (0, 0), so the two loops start
		// at bounds.Min.Y and bounds.Min.X. Looping over Y first and X second is more
		// likely to result in better memory access patterns than X first and Y second.
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c <- coordinates{x, y}
			}
		}
		close(c)
	}()
	return c
}

func decodeImage(imagePath string) (image.Image, error) {
	reader, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	img, imgType, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	log.Printf("Image: %s, type: %s\n", imagePath, imgType)
	return img, nil
}

func getDrawableImage(img image.Image) *image.NRGBA {
	newImg := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(newImg, newImg.Bounds(), img, img.Bounds().Min, draw.Src)
	return newImg
}

func writeImageToPNGFile(img *image.NRGBA, imgPath string) error {
	mapfile, err := os.Create(imgPath)
	if err != nil {
		return err
	}
	defer mapfile.Close()
	return png.Encode(mapfile, img)
}

func diffColor(col1, col2 color.Color, reverse bool) color.Color {
	r1, g1, b1, a1 := uint8ColorRGBA(col1)
	r2, g2, b2, a2 := uint8ColorRGBA(col2)

	r3 := diffUint8(r1, r2, reverse)
	g3 := diffUint8(g1, g2, reverse)
	b3 := diffUint8(b1, b2, reverse)
	if r3 != 0 || g3 != 0 || b3 != 0 {
		diffCount++
	}
	if colorHighlightCalled {
		if colorHighlight == "red" {
			if r3 != 0 || g3 != 0 || b3 != 0 {
				return color.RGBA{255, 0, 0, 255}
			}
		}
		if reverse {
			return color.RGBA{0, 0, 0, 255}
		}
		return color.RGBA{255, 255, 255, 255}
	}
	return color.RGBA{r3, g3, b3, diffAlpha(a1, a2, reverse)}
}

func uint8ColorRGBA(col color.Color) (uint8, uint8, uint8, uint8) {
	r, g, b, a := col.RGBA()
	return uint8(r), uint8(g), uint8(b), uint8(a)
}

func diffUint8(a, b uint8, reverse bool) uint8 {
	if amplify {
		if a != b {
			if baseColor != 0 {
				return 0
			}
			return 255
		}
	}
	if reverse {
		if baseColor != 0 {
			return baseColor - (b - a)
		}
		return b - a
	}
	if baseColor != 0 {
		return baseColor - (a - b)
	}
	return a - b
}

func diffAlpha(a, b uint8, reverse bool) uint8 {
	return 255
}

func saturate(a uint8, extra uint8) uint8 {
	if a+extra > 255 {
		return 255
	}
	return a + extra
}
