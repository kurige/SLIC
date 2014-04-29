package slic

import (
  "image"
  "math"
)

func rgb2xyz(r, g, b uint32) (X, Y, Z float64) {
  var (
    R = float64(r) / 255.0
    G = float64(g) / 255.0
    B = float64(b) / 255.0
  )

  if R > 0.04045 {
    R = math.Pow(((R + 0.055) / 1.055), 2.4)
  } else {
    R = R / 12.92
  }
  if G > 0.04045 {
    G = math.Pow(((G + 0.055) / 1.055), 2.4)
  } else {
    G = G / 12.92
  }
  if B > 0.04045 {
    B = math.Pow(((B + 0.055) / 1.055), 2.4)
  } else {
    B = B / 12.92
  }

  R = R * 100.0
  G = G * 100.0
  B = B * 100.0

  //Observer. = 2°, Illuminant = D65
  X = R*0.4124 + G*0.3576 + B*0.1805
  Y = R*0.2126 + G*0.7152 + B*0.0722
  Z = R*0.0193 + G*0.1192 + B*0.9505
  return
}

func xyz2lab(x, y, z float64) (L, A, B float64) {
  const epsilon float64 = 0.008856
  const kappa float64 = 7.787

  // Observer = 2°, Illuminant = D65
  const rX float64 = 95.047
  const rY float64 = 100.000
  const rZ float64 = 108.883

  var (
    X = x / rX
    Y = y / rY
    Z = z / rZ
  )

  if X > epsilon {
    X = math.Pow(X, (1.0 / 3.0))
  } else {
    X = (kappa * X) + (16.0 / 116.0)
  }
  if Y > epsilon {
    Y = math.Pow(Y, (1.0 / 3.0))
  } else {
    Y = (kappa * Y) + (16.0 / 116.0)
  }
  if Z > epsilon {
    Z = math.Pow(Z, (1.0 / 3.0))
  } else {
    Z = (kappa * Z) + (16.0 / 116.0)
  }

  L = (116.0 * Y) - 16.0
  A = 500.0 * (X - Y)
  B = 200.0 * (Y - Z)
  return
}

func rgb2lab(r, g, b uint32) (L, A, B float64) {
  x, y, z := rgb2xyz(r, g, b)
  return xyz2lab(x, y, z)
}

func imageToLab(img image.Image) (lvec, avec, bvec []float64) {
  var (
    w = img.Bounds().Size().X
    h = img.Bounds().Size().Y
  )

  lvec = make([]float64, w*h)
  avec = make([]float64, w*h)
  bvec = make([]float64, w*h)

  for y := 0; y < h; y++ {
    for x := 0; x < w; x++ {
      var (
        i          = y*w + x
        r, g, b, _ = img.At(x, y).RGBA()
        L, A, B    = rgb2lab(r, g, b)
      )
      lvec[i] = L
      avec[i] = A
      bvec[i] = B
    }
  }

  return
}
