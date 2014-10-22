package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"runtime"
	"sort"
	"sync"
	"strconv"
	"log"
	"os"
)

type RGBASlice struct{
	img *image.RGBA
	X int
}

var (
	tol = 0.1
	noise = 50
)

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

func FindAlikeNeighbor(x, y, xrange, yrange int, img *image.RGBA, mutex *sync.RWMutex) (int, int) {
	mutex.RLock()
	r, g, b, a := img.At(x, y).RGBA()
	mutex.RUnlock()
	nearX, nearY, diff := 0, 0, int(math.MaxInt32)
	for i := x; i < x + xrange; i++ {
		for j := y; j < y + yrange; j++ {
			if (image.Point{x, y}.In(img.Rect)) {
				mutex.RLock()
				nr, ng, nb, na := img.At(i, j).RGBA()
				mutex.RUnlock()
				newDiff := int(math.Abs(float64(nr - r)) + math.Abs(float64(nb - b)) +
					math.Abs(float64(ng - g)) + math.Abs(float64(na - a)))
				if newDiff < diff {
					nearX = i
					nearY = j
					diff = newDiff
					if i == x && j == y {
						diff += noise
					}
				}
			}
		}
	}
	for i := x-1; i > x - xrange; i-- {
		for j := y-1; j > y - yrange; j-- {
			if (image.Point{i, j}.In(img.Rect)) {
				mutex.RLock()
				nr, ng, nb, na := img.At(i, j).RGBA()
				mutex.RUnlock()
				newDiff := int(math.Abs(float64(nr - r)) + math.Abs(float64(nb - b)) +
					math.Abs(float64(ng - g)) + math.Abs(float64(na - a)))
				if newDiff < diff {
					nearX = i
					nearY = j
					diff = newDiff
					if i == x && j == y {
						diff += noise
					}
				}
			}
		}
	}
	return nearX, nearY
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) < 4 || len(os.Args) > 5 {
		log.Fatalln("Usage:", os.Args[0], " <src img> <dest img> <sort type> [<tolerance/range>]")
	}
	imgSrc := os.Args[1]
	imgDest := os.Args[2]
	fmt.Println(imgSrc)
	fmt.Println(imgDest)
	var err error
	var sort_type int64
	xyrange := 10
	if len(os.Args) >= 4 {
		sort_type, err = strconv.ParseInt(os.Args[3], 0, 10)
	}
	if err != nil {
		log.Panicln(err)
	}
	if len(os.Args) == 5 {
		if sort_type == 0 {
			tol, err = strconv.ParseFloat(os.Args[4], 64)
		} else if sort_type == 1 {
			var xy64 int64
			xy64, err = strconv.ParseInt(os.Args[4], 0, 10)
			if err != nil {
				log.Panicln(err)
			}
			xyrange = int(xy64)
		}
	}
	fmt.Println(sort_type)
	fmt.Println(tol)
	fmt.Println(xyrange)

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

	switch sort_type {
	case 0:
		for x := newRGBA.Bounds().Min.X; x < newRGBA.Bounds().Max.X; x++ {
			imgSlice := RGBASlice{newRGBA, x}
			go sort.Sort(imgSlice)
		}
		break
	case 1:
		var wg sync.WaitGroup
		var mutex sync.RWMutex
		for i := 0; i < 10; i++{
			wg.Add(newRGBA.Bounds().Max.X - newRGBA.Bounds().Min.X)
			for x := newRGBA.Bounds().Min.X; x < newRGBA.Bounds().Max.X; x++ {
				go func(newRGBA *image.RGBA, x int, wg *sync.WaitGroup, mutex *sync.RWMutex) {
					defer wg.Done()
					for y := newRGBA.Bounds().Min.Y; y < newRGBA.Bounds().Max.Y; y++ {
						newX, newY := FindAlikeNeighbor(x, y, xyrange, xyrange, newRGBA, mutex)
						m := math.Abs(float64(newY - y))/math.Abs(float64(newX - x))
						var swapX, swapY int
						if newX == x && newY == y {
							swapX = x
							swapY = y
						} else if m > 1 {
							if newY < y {
								swapX = x
								swapY = y - 1
							} else {
								swapX = x
								swapY = y + 1
							}
						} else {
							if newX < x {
								swapX = x -1
								swapY = y
							} else {
								swapX = x + 1
								swapY = y
							}
						}
						//fmt.Println(x, y, newRGBA.At(x, y), newX, newY, newRGBA.At(newX, newY),  swapX, swapY)
						//fmt.Println(newRGBA.At(swapX, swapY), newRGBA.At(x,y))
						mutex.Lock()
						ctemp := newRGBA.At(x, y)
						newRGBA.Set(x, y, newRGBA.At(swapX, swapY))
						newRGBA.Set(swapX, swapY, ctemp)
						mutex.Unlock()
						//fmt.Println(x, y, swapX, swapY)
						//fmt.Println(newRGBA.At(swapX, swapY), newRGBA.At(x,y), "\n")
					}
				} (newRGBA, x, &wg, &mutex)
			}
			wg.Wait()
		}
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
