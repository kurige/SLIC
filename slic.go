package slic

import (
  "image"
  "image/color"
  "image/draw"
  "math"

  "github.com/kurige/SLIC/lab"
)

/*
 * TODO:
 * - More accurate LAB color diffing
 * - "Upgrade" to SLICO (or make it an option)
 * - Use all avaialble cores
 * - Perturb superpixels during seeding
 * - Support hexgrid seeding(?)
 */

type SuperPixel struct {
  label   int
  L, A, B float64
  X, Y    float64
}

type SLIC struct {
  image       lab.Image
  compactness float64
  step        int
  distvec     []float64
  Superpixels []*SuperPixel
  XStrips     int
  YStrips     int

  Labels     []int
  labelCount int
}

func SuperPixelSizeForCount(width, height, count int) int {
  return int(0.5 + float64(width*height)/float64(count))
}

func MakeSlic(image image.Image, compactness float64, supsz int) *SLIC {
  var (
    w    = image.Bounds().Size().X
    h    = image.Bounds().Size().Y
    sz   = w * h
    step = int(math.Sqrt(float64(supsz)) + 0.5)
  )
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

  // Overwrite user selected superpixel count if necessary.
  supsz = x_strips * y_strips
  superpixels := make([]*SuperPixel, supsz)

  img := lab.ImageToLab(image)

  slic := &SLIC{
    img,
    compactness,
    step,
    distvec,
    superpixels,

    x_strips,
    y_strips,

    labels,
    0,
  }

  x_err_per_strip := float64(x_err) / float64(x_strips)
  y_err_per_strip := float64(y_err) / float64(y_strips)
  x_offset := step / 2
  y_offset := step / 2
  label := 0
  for y := 0; y < y_strips; y++ {
    ye := y * int(y_err_per_strip)
    for x := 0; x < x_strips; x++ {
      var (
        xe    = x * int(x_err_per_strip)
        seedx = x*step + x_offset + xe
        seedy = y*step + y_offset + ye
        c     = img.At(seedx, seedy).(lab.Color)
      )
      superpixels[label] = &SuperPixel{label, c.L, c.A, c.B, float64(seedx), float64(seedy)}
      label++
    }
  }

  return slic
}

func (slic *SLIC) Run(iterations int) {
  if iterations <= 0 {
    iterations = 1
  }
  for i := 0; i < iterations; i++ {
    slic.resetDistances()
    slic.labelPixels()
    slic.recalculateCentroids()
  }

  label_count, new_labels := slic.enforceLabelConnectivity()
  slic.labelCount = label_count

  size := slic.image.Bounds().Size()
  sz := size.X * size.Y
  for i := 0; i < sz; i++ {
    slic.Labels[i] = new_labels[i]
  }
}

func (slic *SLIC) resetDistances() {
  for index := range slic.distvec {
    slic.distvec[index] = math.MaxFloat64
  }
}

func (slic *SLIC) labelPixels() {
  for n := range slic.Superpixels {
    slic.labelPixelsInSuperpixel(slic.Superpixels[n])
  }
}

func (slic *SLIC) labelPixelsInSuperpixel(s *SuperPixel) {
  fstep := float64(slic.step)
  invwt := 1.0 / ((fstep / slic.compactness) * (fstep / slic.compactness))

  size := slic.image.Bounds().Size()
  width, height := size.X, size.Y
  y1 := int(math.Max(0.0, s.Y-fstep))
  y2 := int(math.Min(float64(height), s.Y+fstep))
  x1 := int(math.Max(0.0, s.X-fstep))
  x2 := int(math.Min(float64(width), s.X+fstep))

  supL, supA, supB := s.L, s.A, s.B
  supX, supY := s.X, s.Y

  for y := y1; y < y2; y++ {
    for x := x1; x < x2; x++ {
      c := slic.image.At(x, y).(lab.Color)
      X, Y := float64(x), float64(y)
      var distc float64 = (c.L-supL)*(c.L-supL) + (c.A-supA)*(c.A-supA) + (c.B-supB)*(c.B-supB)
      var distxy float64 = (X-supX)*(X-supX) + (Y-supY)*(Y-supY)

      dist := math.Sqrt(distc) + math.Sqrt(distxy*invwt)

      i := y*width + x
      if dist < slic.distvec[i] {
        slic.distvec[i] = dist
        slic.Labels[i] = s.label
      }
    }
  }
}

func (slic *SLIC) AverageColors() (lvec, avec, bvec []float64) {
  lvec = make([]float64, slic.labelCount)
  avec = make([]float64, slic.labelCount)
  bvec = make([]float64, slic.labelCount)
  count := make([]int, slic.labelCount)

  size := slic.image.Bounds().Size()
  width, height := size.X, size.Y

  for y := 0; y < height; y++ {
    for x := 0; x < width; x++ {
      i := y*width + x
      label := slic.Labels[i]
      c := slic.image.At(x, y).(lab.Color)
      lvec[label] += c.L
      avec[label] += c.A
      bvec[label] += c.B
      count[label]++
    }
  }

  for i := 0; i < slic.labelCount; i++ {
    count := float64(count[i])
    lvec[i] = lvec[i] / count
    avec[i] = avec[i] / count
    bvec[i] = bvec[i] / count
  }

  return
}

func (slic *SLIC) recalculateCentroids() {
  supsz := len(slic.Superpixels)
  sigma_l := make([]float64, supsz)
  sigma_a := make([]float64, supsz)
  sigma_b := make([]float64, supsz)
  sigma_x := make([]float64, supsz)
  sigma_y := make([]float64, supsz)
  clustersize := make([]float64, supsz)

  size := slic.image.Bounds().Size()
  width, height := size.X, size.Y

  for y := 0; y < height; y++ {
    for x := 0; x < width; x++ {
      i := y*width + x
      label := slic.Labels[i]
      // This needs to be handled better...
      if label == -1 {
        continue
      }
      c := slic.image.At(x, y).(lab.Color)
      sigma_l[label] += c.L
      sigma_a[label] += c.A
      sigma_b[label] += c.B
      sigma_x[label] += float64(x)
      sigma_y[label] += float64(y)
      clustersize[label] += 1.0
    }
  }

  for n := 0; n < supsz; n++ {
    if clustersize[n] <= 0 {
      clustersize[n] = 1.0
    }

    superpixel := slic.Superpixels[n]
    superpixel.L = sigma_l[n] / clustersize[n]
    superpixel.A = sigma_a[n] / clustersize[n]
    superpixel.B = sigma_b[n] / clustersize[n]
    superpixel.X = sigma_x[n] / clustersize[n]
    superpixel.Y = sigma_y[n] / clustersize[n]
  }
}

func (slic *SLIC) enforceLabelConnectivity() (int, []int) {
  size := slic.image.Bounds().Size()
  width, height := size.X, size.Y
  sz := width * height
  target_supsz := sz / (slic.step * slic.step)
  SUPSZ := sz / target_supsz

  dx4 := [...]int{-1, 0, 1, 0}
  dy4 := [...]int{0, -1, 0, 1}

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

              if 0 > nlabels[nindex] && slic.Labels[oindex] == slic.Labels[nindex] {
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

func (slic *SLIC) DrawEdgesToImage(img image.Image) image.Image {
  // Create new RGBA image from source
  b := img.Bounds()
  canvas := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(canvas, canvas.Bounds(), img, b.Min, draw.Src)

  dx8 := []int{-1, -1, 0, 1, 1, 1, 0, -1}
  dy8 := []int{0, -1, -1, -1, 0, 1, 1, 1}

  size := slic.image.Bounds().Size()
  width, height := size.X, size.Y
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
      if slic.Labels[mainindex] == -1 {
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
            if slic.Labels[mainindex] != slic.Labels[index] {
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
