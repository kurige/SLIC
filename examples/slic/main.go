package main

import (
  "bufio"
  "flag"
  "fmt"
  "image"
  _ "image/gif"
  _ "image/jpeg"
  "image/png"
  "log"
  "os"
  "runtime"
  "runtime/pprof"

  "github.com/kurige/SLIC"
)

type handlerFunc func(*os.File)

func outputPNG(img image.Image, filename string) {
  fi, _ := os.Create(filename)
  png.Encode(fi, img)
  //jpeg.Encode(fi, img, &jpeg.Options{jpeg.DefaultQuality})
}

func outputLabels(w, h int, labels []int, filename string) {
  fi, _ := os.Create(filename)
  writer := bufio.NewWriter(fi)

  for x := 0; x < w; x++ {
    for y := 0; y < h; y++ {
      i := y*w + x
      label := labels[i]
      if _, err := fmt.Fprintf(writer, "%d,", label); err != nil {
        panic(err)
      }
    }
    if _, err := fmt.Fprintf(writer, "\n"); err != nil {
      panic(err)
    }
    writer.Flush()
  }
}

var (
  // outputName = flag.String("o", "output", "\t\tName of the output filename (sans extension)")
  // outputExt = flag.Uint("e", 1, "\t\tOutput extension type:\n\t\t\t 1 \t png (default)\n\t\t\t 2 \t jpg")
  // jpgQuality = flag.Int("q", 90, "\t\tJPG output quality")
  num_cores      = flag.Int("cpu", 0, "Max number of cores to utilize (0 means use all available)")
  superpixels    = flag.Int("pixels", -1, "Number of superpixels to use")
  superpixelsize = flag.Int("size", 40, "Super pixel size")
  compactness    = flag.Float64("c", 20.0, "Superpixel 'compactness'")
  cpuprofile     = flag.String("cpuprofile", "", "write cpu profile to file")
  iterations     = flag.Int("i", 10, "Number of iterations")
)

func main() {
  flag.Parse()

  var nc int
  if *num_cores <= 0 || *num_cores > runtime.NumCPU() {
    nc = runtime.NumCPU()
  } else {
    nc = *num_cores
  }
  runtime.GOMAXPROCS(nc)

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

  w := src_img.Bounds().Size().X
  h := src_img.Bounds().Size().Y
  if *superpixels != -1 {
    *superpixelsize = slic.SuperPixelSizeForCount(w, h, *superpixels)
  }

  s := slic.MakeSlic(src_img, *compactness, *superpixelsize)
  s.Run(*iterations)

  outputPNG(s.DrawEdgesToImage(src_img), "out.png")
  // outputLabels(w, h, s.Labels, "out.labels")
}
