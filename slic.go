package main

import (
  "flag"
  "image"
  "image/color"
  "image/draw"
  _ "image/gif"
  _ "image/jpeg"
  "image/png"
  "log"
  "math"
  "os"
  "runtime"
  "strconv"
  "sync"
  "time"
)

var SLIC_ITERATIONS = 10

/*
 * TODO:
 * - Switch to LAB colorspace
 * - Perturb superpixels during seeding
 * - Support hexgrid seeding(?)
 */

type SuperPixel struct {
  label   int
  R, G, B float64
  X, Y    float64
}

type SLIC struct {
  r, g, b     []int
  w, h, sz    int
  compactness float64
  step        int
  labels      []int
  distvec     []float64
  superpixels []*SuperPixel
}

func makeSlic(image image.Image, compactness float64, size int) *SLIC {
  w := image.Bounds().Size().X
  h := image.Bounds().Size().Y
  sz := w * h
  step := int(math.Sqrt(float64(size)) + 0.5)

  x_strips := int(0.5 + float64(w)/float64(step))
  y_strips := int(0.5 + float64(h)/float64(step))
  x_err := w - step*x_strips
  if x_err < 0 {
    x_strips--
    x_err = w - step*x_strips
  }
  y_err := h - step*y_strips
  if y_err < 0 {
    y_strips--
    y_err = h - step*y_strips
  }

  labels := make([]int, sz)
  for i := 0; i < sz; i++ {
    labels[i] = -1
  }
  distvec := make([]float64, sz)

  supsz := x_strips * y_strips
  superpixels := make([]*SuperPixel, supsz)

  log.Println("Image width:", w)
  log.Println("Image height:", h)
  log.Println("x-strips:", x_strips)
  log.Println("y-strips:", y_strips)

  log.Println("SLIC:")
  log.Println("\tCompactness:", compactness)
  log.Println("\tSuperpixels:", supsz)
  log.Println("\tStep:", step)

  r := make([]int, sz)
  g := make([]int, sz)
  b := make([]int, sz)
  for y := 0; y < h; y++ {
    for x := 0; x < w; x++ {
      img_r, img_g, img_b, _ := image.At(x, y).RGBA()
      i := y*w + x
      r[i] = int(img_r)
      g[i] = int(img_g)
      b[i] = int(img_b)
    }
  }

  slic := &SLIC{
    r, g, b,
    w, h, sz,
    compactness,
    step,
    labels,
    distvec,
    superpixels,
  }

  x_err_per_strip := float64(x_err) / float64(x_strips)
  y_err_per_strip := float64(y_err) / float64(y_strips)
  x_offset := step / 2
  y_offset := step / 2
  label := 0
  for y := 0; y < y_strips; y++ {
    ye := y * int(y_err_per_strip)
    for x := 0; x < x_strips; x++ {
      xe := x * int(x_err_per_strip)
      seedx := x*step + x_offset + xe
      seedy := y*step + y_offset + ye
      color := image.At(seedx, seedy)
      superpixels[label] = slic.makeSuperpixel(label, color, seedx, seedy)
      label++
    }
  }

  return slic
}

func (slic *SLIC) makeSuperpixel(label int, color color.Color, x, y int) *SuperPixel {
  r, g, b, _ := color.RGBA()
  superpixel := &SuperPixel{label, float64(r), float64(g), float64(b), float64(x), float64(y)}
  return superpixel
}

func (slic *SLIC) resetDistances() {
  for index := range slic.distvec {
    slic.distvec[index] = math.MaxFloat64
  }
}

func (slic *SLIC) labelPixelsInSuperpixel(s *SuperPixel, wg *sync.WaitGroup) {
  fstep := float64(slic.step)
  invwt := 1.0 / ((fstep / slic.compactness) * (fstep / slic.compactness))

  y1 := int(math.Max(0.0, s.Y-fstep))
  y2 := int(math.Min(float64(slic.h), s.Y+fstep))
  x1 := int(math.Max(0.0, s.X-fstep))
  x2 := int(math.Min(float64(slic.w), s.X+fstep))

  wg.Add(1)
  go func() {
    for y := y1; y < y2; y++ {
      for x := x1; x < x2; x++ {
        i := y*slic.w + x
        r1, g1, b1 := float64(slic.r[i]), float64(slic.g[i]), float64(slic.b[i])
        r2, g2, b2 := s.R, s.G, s.B
        x1, y1 := float64(x), float64(y)
        x2, y2 := s.X, s.Y
        var distc float64 = (r1-r2)*(r1-r2) + (g1-g2)*(g1-g2) + (b1-b2)*(b1-b2)
        var distxy float64 = (x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)

        dist := distc + (distxy * invwt)
        // dist := math.Sqrt(distc) + math.Sqrt(distxy*invwt) // More exact

        if dist < slic.distvec[i] {
          slic.distvec[i] = dist
          slic.labels[i] = s.label
        }
        // fmt.Print(strconv.FormatFloat(dist, 'f', 2, 64) + "\t")
      }
      // fmt.Print("\n")
    }
    wg.Done()
  }()
}

func (slic *SLIC) recalculateCentroids() {
  supsz := len(slic.superpixels)
  sigma_r := make([]float64, supsz)
  sigma_g := make([]float64, supsz)
  sigma_b := make([]float64, supsz)
  sigma_x := make([]float64, supsz)
  sigma_y := make([]float64, supsz)
  clustersize := make([]float64, supsz)

  pixel := 0
  for y := 0; y < slic.h; y++ {
    for x := 0; x < slic.w; x++ {
      label := slic.labels[pixel]
      // This needs to be handled better...
      if slic.labels[pixel] == -1 {
        pixel++
        continue
      }
      i := y*slic.w + x
      r, g, b := float64(slic.r[i]), float64(slic.g[i]), float64(slic.b[i])
      sigma_r[label] += r
      sigma_g[label] += g
      sigma_b[label] += b
      sigma_x[label] += float64(x)
      sigma_y[label] += float64(y)
      clustersize[label] += 1.0
      pixel++
    }
  }

  for n := 0; n < supsz; n++ {
    if clustersize[n] <= 0 {
      clustersize[n] = 1.0
    }

    superpixel := slic.superpixels[n]
    superpixel.R = sigma_r[n] / clustersize[n]
    superpixel.G = sigma_g[n] / clustersize[n]
    superpixel.B = sigma_b[n] / clustersize[n]
    superpixel.X = sigma_x[n] / clustersize[n]
    superpixel.Y = sigma_y[n] / clustersize[n]
  }
}

func (slic *SLIC) enforceLabelConnectivity(target_supsz int) (int, []int) {
  dx4 := [...]int{-1, 0, 1, 0}
  dy4 := [...]int{0, -1, 0, 1}

  height, width := slic.h, slic.w
  sz := slic.w * slic.h
  SUPSZ := sz / target_supsz

  log.Println("Target superpixel count:", target_supsz)
  log.Println("SUPSZ:", SUPSZ)

  xvec := make([]int, sz)
  yvec := make([]int, sz)
  oindex := 0
  adjlabel := 0

  label := 0
  nlabels := make([]int, sz)

  for i := 0; i < sz; i++ {
    nlabels[i] = -1
  }

  for j := 0; j < height; j++ {
    for k := 0; k < width; k++ {
      if 0 > nlabels[oindex] {
        nlabels[oindex] = label

        // Start a new segment
        xvec[0] = k
        yvec[0] = j

        // Quickly find an adjacent label for use later if needed
        for n := 0; n < 4; n++ {
          x := xvec[0] + dx4[n]
          y := yvec[0] + dy4[n]
          if (x >= 0 && x < width) && (y >= 0 && y < height) {
            nindex := y*width + x
            if nlabels[nindex] >= 0 {
              adjlabel = nlabels[nindex]
            }
          }
        }

        count := 1
        for c := 0; c < count; c++ {
          for n := 0; n < 4; n++ {
            x := xvec[c] + dx4[n]
            y := yvec[c] + dy4[n]

            if (x >= 0 && x < width) && (y >= 0 && y < height) {
              nindex := y*width + x

              if 0 > nlabels[nindex] && slic.labels[oindex] == slic.labels[nindex] {
                xvec[count] = x
                yvec[count] = y
                nlabels[nindex] = label
                count++
              }
            }
          }
        }

        // If segment size is less than the limit, assign an adjacent label
        // found before, and decrement label count.
        if count <= SUPSZ>>2 {
          for c := 0; c < count; c++ {
            ind := yvec[c]*width + xvec[c]
            nlabels[ind] = adjlabel
          }
          label--
        }
        label++
      }
      oindex++
    }
  }

  return label, nlabels
}

func (slic *SLIC) drawEdgesToImage(img image.Image) image.Image {
  // Create new RGBA image from source
  b := img.Bounds()
  canvas := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(canvas, canvas.Bounds(), img, b.Min, draw.Src)

  dx8 := []int{-1, -1, 0, 1, 1, 1, 0, -1}
  dy8 := []int{0, -1, -1, -1, 0, 1, 1, 1}

  width, height := slic.w, slic.h
  sz := width * height
  contourx := make([]int, sz)
  contoury := make([]int, sz)
  istaken := make([]bool, sz)
  mainindex := 0
  cind := 0

  black := color.RGBA{0, 0, 0, 254}
  red := color.RGBA{254, 0, 0, 254}

  for j := 0; j < height; j++ {
    for k := 0; k < width; k++ {
      if slic.labels[mainindex] == -1 {
        canvas.Set(k, j, red)
        mainindex++
        continue
      }

      np := 0
      for i := 0; i < 8; i++ {
        x := k + dx8[i]
        y := j + dy8[i]

        if (x >= 0 && x < width) && (y >= 0 && y < height) {
          index := y*width + x
          if !istaken[index] {
            if slic.labels[mainindex] != slic.labels[index] {
              np++
            }
          }
        }
      }
      if np > 1 {
        contourx[cind] = k
        contoury[cind] = j
        istaken[mainindex] = true
        cind++
      }
      mainindex++
    }
  }

  for j := 0; j < cind; j++ {
    x := contourx[j]
    y := contoury[j]
    canvas.Set(x, y, black)
  }

  return canvas
}

func writePNGToDisk(img image.Image, filename string) {
  out, _ := os.Create(filename)
  png.Encode(out, img)
  // png.Encode(out, img, &jpeg.Options{jpeg.DefaultQuality})
  out.Close()
}

func main() {
  // outputName := flag.String("o", "output", "\t\tName of the output filename (sans extension)")
  // outputExt := flag.Uint("e", 1, "\t\tOutput extension type:\n\t\t\t 1 \t png (default)\n\t\t\t 2 \t jpg")
  // jpgQuality := flag.Int("q", 90, "\t\tJPG output quality")
  num_cores := flag.Int("cores", 0, "Max number of cores to utilize (0 means use all available)")
  superpixels := flag.Int("superpixels", -1, "Number of superpixels to use")
  superpixelsize := flag.Int("superpixelsize", 40, "Super pixel size")
  compactness := flag.Float64("compactness", 20.0, "Superpixel 'compactness'")
  flag.Parse()

  var nc int
  if *num_cores <= 0 || *num_cores > runtime.NumCPU() {
    nc = runtime.NumCPU()
  } else {
    nc = *num_cores
  }
  runtime.GOMAXPROCS(nc)
  log.Println("Number of Cores:", nc)

  file_name := flag.Arg(0)
  file, err := os.Open(file_name)
  if err != nil {
    log.Println(err)
    // return err
    return
  }
  defer file.Close()

  src_img, _, err := image.Decode(file)
  if err != nil {
    log.Println(err, "Could not decode image:", file_name)
    // return nil
    return
  }

  width, height := src_img.Bounds().Size().X, src_img.Bounds().Size().Y
  if *superpixels != -1 {
    *superpixelsize = int(0.5 + float64(width*height)/float64(*superpixels))
  }

  slic := makeSlic(src_img, *compactness, *superpixelsize)
  start := time.Now()
  for i := 0; i < SLIC_ITERATIONS; i++ {
    log.Println("SLIC Iteration", i)

    slic.resetDistances()
    var wg sync.WaitGroup
    for n := range slic.superpixels {
      superpixel := slic.superpixels[n]
      slic.labelPixelsInSuperpixel(superpixel, &wg)
    }
    wg.Wait()

    slic.recalculateCentroids()

    writePNGToDisk(slic.drawEdgesToImage(src_img), "out_"+strconv.Itoa(i)+".png")
  }

  sz := width * height
  target_superpixels := sz / (slic.step * slic.step)
  new_labels_count, new_labels := slic.enforceLabelConnectivity(target_superpixels)

  log.Println("Final labels count:", new_labels_count)

  for i := 0; i < sz; i++ {
    // log.Println(new_labels[i])
    slic.labels[i] = new_labels[i]
  }

  elapsed := time.Since(start)
  log.Println("(", elapsed, ")")

  writePNGToDisk(slic.drawEdgesToImage(src_img), "out.png")
}
