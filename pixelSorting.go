package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

type RGBASlice struct {
	img *image.RGBA
	X   int
}

var (
	tol   = 0.1
	noise = 50
)

func (img RGBASlice) Len() int {
	pr, pg, pb, pa := img.img.At(img.X, img.img.Bounds().Min.Y).RGBA()
	for y := img.img.Bounds().Min.Y + 1; y < img.img.Bounds().Max.Y; y++ {
		r, g, b, a := img.img.At(img.X, y).RGBA()
		rdiff := math.Abs(float64(pr)-float64(r)) / (float64(pr+r) / 2)
		gdiff := math.Abs(float64(pg)-float64(g)) / (float64(pg+g) / 2)
		bdiff := math.Abs(float64(pb)-float64(b)) / (float64(pb+b) / 2)
		adiff := math.Abs(float64(pa)-float64(a)) / (float64(pa+a) / 2)
		/*fmt.Printf("%d %d\n", pr, r)
		fmt.Printf("%f %f\n", math.Abs(float64(pr) - float64(r)), float64(pr + r)/2)
		fmt.Printf("%d %d\n", pg, g)
		fmt.Printf("%d %d\n", pb, b)
		fmt.Printf("%d %d\n", pa, a)
		fmt.Printf("%f %f %f %f\n", rdiff, bdiff, gdiff, adiff)*/
		if rdiff >= tol || bdiff >= tol || gdiff >= tol || adiff >= tol {
			if y != img.img.Bounds().Max.Y {
				return y + 1
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
	return ir+ig+ib+ia > jr+jg+jb+ja
}

func (img RGBASlice) Swap(i, j int) {
	temp := img.img.At(img.X, i)
	img.img.Set(img.X, i, img.img.At(img.X, j))
	img.img.Set(img.X, j, temp)
}

func FindAlikeNeighbor(x, y, xrange, yrange int, img *image.RGBA, mutexes [][]sync.RWMutex, localRand *rand.Rand) (int, int) {
	mutexes[x][y].RLock()
	r, g, b, _ := img.At(x, y).RGBA()
	mutexes[x][y].RUnlock()
	nearX, nearY, diff := 0, 0, int(math.MaxInt32)

	iVals := localRand.Perm(xrange*2+1)
	for k := 0; k < len(iVals); k++ {
		iVals[k] += x - xrange
	}

	for iv := 0; iv < len(iVals); iv++ {
		i := iVals[iv]
		if i < 0 || i >= len(mutexes) {
			continue
		}

		jVals := localRand.Perm(yrange*2+1)
		for k := 0; k < len(jVals); k++ {
			jVals[k] += y - yrange
		}

		for jv := 0; jv < len(jVals); jv++ {
			j := jVals[jv]
			if j < 0 || j >= len(mutexes[i]) {
				continue
			}

			mutexes[i][j].RLock()
			nr, ng, nb, _ := img.At(i, j).RGBA()
			mutexes[i][j].RUnlock()
			newDiff := int(math.Abs(float64(nr-r)) + math.Abs(float64(nb-b)) +
				math.Abs(float64(ng-g)))
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
	return nearX, nearY
}

func main() {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) < 4 || len(os.Args) > 6 {
		log.Fatalln("Usage:", os.Args[0], " <src img> <dest img> <sort type> [<tolerance/range>] [<iterations>]")
	}
	imgSrc := os.Args[1]
	imgDest := os.Args[2]
	fmt.Println(imgSrc)
	fmt.Println(imgDest)
	var err error
	var sort_type int64
	xyrange := 10
	iterations := 10
	if len(os.Args) >= 4 {
		sort_type, err = strconv.ParseInt(os.Args[3], 0, 10)
	}
	if err != nil {
		log.Panicln(err)
	}
	if len(os.Args) >= 5 {
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
	if len(os.Args) == 6 {
		var iter64 int64
		iter64, err = strconv.ParseInt(os.Args[5], 0, 10)
		if err != nil {
			log.Panicln(err)
		}
		iterations = int(iter64)
	}
	fmt.Println(sort_type)
	fmt.Println(tol)
	fmt.Println(xyrange)
	fmt.Println(iterations)

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
		mutexes := make([][]sync.RWMutex, newRGBA.Bounds().Max.X)
		for i := 0; i < len(mutexes); i++ {
			mutexes[i] = make([]sync.RWMutex, newRGBA.Bounds().Max.Y)
		}
		for i := 0; i < iterations; i++ {
			wg.Add(newRGBA.Bounds().Max.X)
			xVals := rand.Perm(newRGBA.Bounds().Max.X)
			for i := 0; i < len(xVals); i++ {
				x := xVals[i]
				go func(newRGBA *image.RGBA, x int, wg *sync.WaitGroup, mutexes [][]sync.RWMutex) {
					localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
					defer wg.Done()
					yVals := localRand.Perm(newRGBA.Bounds().Max.Y)
					for j := 0; j < len(yVals); j++ {
						y := yVals[j]
						newX, newY := FindAlikeNeighbor(x, y, xyrange, xyrange, newRGBA, mutexes, localRand)
						m := math.Abs(float64(newY-y)) / math.Abs(float64(newX-x))
						var swapX, swapY int
						if newX == x && newY == y {
							continue
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
								swapX = x - 1
								swapY = y
							} else {
								swapX = x + 1
								swapY = y
							}
						}
						//fmt.Println(x, y, newRGBA.At(x, y), newX, newY, newRGBA.At(newX, newY),  swapX, swapY)
						//fmt.Println(newRGBA.At(swapX, swapY), newRGBA.At(x,y))

						//this segment doesn't really *need* to atomic...
						mutexes[x][y].RLock()
						c1 := newRGBA.At(x, y)
						mutexes[x][y].RUnlock()

						mutexes[swapX][swapY].RLock()
						c2 := newRGBA.At(swapX, swapY)
						mutexes[swapX][swapY].RUnlock()

						mutexes[x][y].Lock()
						newRGBA.Set(x, y, c2)
						mutexes[x][y].Unlock()

						mutexes[swapX][swapY].Lock()
						newRGBA.Set(swapX, swapY, c1)
						mutexes[swapX][swapY].Unlock()
						//fmt.Println(x, y, swapX, swapY)
						//fmt.Println(newRGBA.At(swapX, swapY), newRGBA.At(x,y), "\n")
					}
				}(newRGBA, x, &wg, mutexes)
			}
			wg.Wait()
			switch (i + 1) % 10 {
			case 0:
				fmt.Print("X")
				break
			case 5:
				fmt.Print("|")
				break
			default:
				fmt.Print(".")
				break
			}
		}
		fmt.Println()
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
