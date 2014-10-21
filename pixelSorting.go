package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"sort"
	"strconv"
	"log"
	"os"
)

type RGBASlice struct{
	img *image.RGBA
	X int
}

var tol = 0.1

func (img RGBASlice) Len() int {
	pr, pg, pb, pa := img.img.At(img.X, img.img.Bounds().Min.Y).RGBA()
	for y := img.img.Bounds().Min.Y + 1; y < img.img.Bounds().Max.Y; y++ {
		r, g, b, a := img.img.At(img.X, y).RGBA()
		rdiff := math.Abs(float64(pr) - float64(r)) / (float64(pr + r)/2)
		gdiff := math.Abs(float64(pg) - float64(g)) / (float64(pg + g)/2)
		bdiff := math.Abs(float64(pb) - float64(b)) / (float64(pb + b)/2)
		adiff := math.Abs(float64(pa) - float64(a)) / (float64(pa + a)/2)
		/*fmt.Printf("%d %d\n", pr, r)
		fmt.Printf("%f %f\n", math.Abs(float64(pr) - float64(r)), float64(pr + r)/2)
		fmt.Printf("%d %d\n", pg, g)
		fmt.Printf("%d %d\n", pb, b)
		fmt.Printf("%d %d\n", pa, a)
		fmt.Printf("%f %f %f %f\n", rdiff, bdiff, gdiff, adiff)*/
		if rdiff >= tol || bdiff >= tol || gdiff >= tol || adiff >= tol {
			if y != img.img.Bounds().Max.Y {
				return y+1
			}
			return y
		}
		pr, pg, pb, pa = r, g, b, a
	}
	return img.img.Bounds().Max.Y
}

func (img RGBASlice) Less(i, j int) bool {
	ir, ig, ib, ia := img.img.At(img.X, i).RGBA()
	jr, jg, jb, ja := img.img.At(img.X, j).RGBA()
	return ir + ig + ib + ia > jr + jg + jb + ja
}

func (img RGBASlice) Swap(i, j int) {
	temp := img.img.At(img.X, i)
	img.img.Set(img.X, i, img.img.At(img.X, j))
	img.img.Set(img.X, j, temp)
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		log.Fatalln("Usage:", os.Args[0], " <src img> <dest img> [<tolerance>]")
	}
	imgSrc := os.Args[1]
	imgDest := os.Args[2]
	fmt.Println(imgSrc)
	fmt.Println(imgDest)
	var err error
	if len(os.Args) == 4 {
		tol, err = strconv.ParseFloat(os.Args[3], 64)
	}
	if err != nil {
		log.Panicln(err)
	}
	//load image from file
	//imgFile, err := os.Open(`D:\Users\Ross\Dropbox\Camera Uploads\2014-07-19 11.02.20.jpg`)
	imgFile, err := os.Open(imgSrc)
	if err != nil {
		log.Panicln(err)
	}
	img, format, err := image.Decode(imgFile)
	fmt.Println(format)
	if err != nil {
		log.Panicln(err)
	}

	//convert image to RBGA
	imgRect := img.Bounds()
	newRGBA := image.NewRGBA(imgRect)
	for x := imgRect.Min.X; x < imgRect.Max.X; x++ {
		for y := imgRect.Min.Y; y < imgRect.Max.Y; y++ {
			newRGBA.Set(x, y, img.At(x, y))
		}
	}

	for x := newRGBA.Bounds().Min.X; x < newRGBA.Bounds().Max.X; x++ {
		imgSlice := RGBASlice{newRGBA, x}
		sort.Sort(imgSlice)
	}

	//save RGBA to file
	newImgFile, err := os.Create(imgDest)
	if err != nil {
		log.Panicln(err)
	}
	err = png.Encode(newImgFile, newRGBA)
	if err != nil {
		log.Panicln(err)
	}
}
