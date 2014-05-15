package main

import (
  "flag"
  "fmt"
  "image"
  "image/color"
  _ "image/gif"
  _ "image/jpeg"
  "image/png"
  "log"
  "os"
  "runtime"
  "runtime/pprof"

  "github.com/kurige/SLIC"
  "github.com/kurige/SLIC/lab"
)

type handlerFunc func(*os.File)

func outputPNG(img image.Image, filename string) {
  fi, _ := os.Create(filename)
  png.Encode(fi, img)
}

var (
  superpixels    = flag.Int("pixels", -1, "Number of superpixels to use")
  superpixelsize = flag.Int("size", 40, "Super pixel size")
  cpuprofile     = flag.String("cpuprofile", "", "write cpu profile to file")
  compactness    = flag.Float64("c", 20.0, "Superpixel 'compactness'")
)

func main() {
  flag.Parse()
  runtime.GOMAXPROCS(runtime.NumCPU())
  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
      log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
  }

  file_name := flag.Arg(0)
  file, err := os.Open(file_name)
  if err != nil {
    log.Println(err)
    return
  }
  defer file.Close()

  src_img, _, err := image.Decode(file)
  if err != nil {
    log.Println(err, "Could not decode image:", file_name)
    return
  }

  src_w := src_img.Bounds().Size().X
  src_h := src_img.Bounds().Size().Y

  if *superpixels != -1 {
    *superpixelsize = slic.SuperPixelSizeForCount(src_w, src_h, *superpixels)
  }
  fmt.Println("Pixel size:", *superpixelsize)

  s := slic.MakeSlic(src_img, *compactness, *superpixelsize)

  s.Run(10)
  lvec, avec, bvec := s.AverageColors()

  out := image.NewNRGBA(image.Rect(0, 0, src_w, src_h))
  for y := 0; y < src_h; y++ {
    for x := 0; x < src_w; x++ {
      i := y*src_w + x
      label := s.Labels[i]
      l := lvec[label]
      a := avec[label]
      b := bvec[label]
      R, G, B := lab.Lab2rgb(l, a, b)
      c := color.RGBA{R, G, B, 255}
      out.Set(x, y, c)
    }
  }

  outputPNG(out, "out.png")
}
