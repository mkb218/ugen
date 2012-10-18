package main
// plotter.go
// Joe Wass 2012
// joe@afandian.com

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"sndfile"
)

// Number of seconds worth of buffer to allocate.
const Seconds = 10

// Height per channel.
const ImageHeight = 200

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run plotter.go /path/to/wav /path/to/png")
		return
	}

	var info sndfile.Info
	soundFile, err := sndfile.Open(os.Args[1], sndfile.Read, &info)

	if err != nil {
		log.Fatal("Error", err)
	}

	defer soundFile.Close()

	imageFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer imageFile.Close()

	buffer := make([]float32, Seconds*info.Samplerate*info.Channels)

	numRead, err := soundFile.ReadItems(buffer)
	numSamples := int(numRead/int64(info.Channels))
	numChannels := int(info.Channels)

	outimage := image.NewRGBA(image.Rect(0, 0, numSamples, ImageHeight * numChannels))

	if err != nil {
		return
	}

	// Both math.Abs and math.Max operate on float64. Hm.
	max := float32(0)
	for _, v := range buffer {
		if v > max {
			max = v
		} else if v*-1 > max {
			max = v * -1
		}
	}

	// Work out scaling factor to normalise signaland get best use of space.
	mult := float32(ImageHeight/max) / 2

	// Signed float so add 1 to turn [-1, 1] into [0, 2]. 
	for i := 0; i < numSamples; i++ {
		for channel := 0; channel < numChannels; channel ++ {
			y := int(buffer[i*numChannels+channel]*mult+ImageHeight/2) + ImageHeight * channel
			outimage.Set(i, y, color.Black)
			outimage.Set(i, y+1, color.Black)

		}
		
	}

	png.Encode(imageFile, outimage)
}
