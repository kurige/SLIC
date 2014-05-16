package lab

import (
  "image"
  "image/color"
  "image/draw"
  "math"
)

// Observer = 2°, Illuminant = D65
const RefX float64 = 95.047
const RefY float64 = 100.000
const RefZ float64 = 108.883

func rgb2xyz(r, g, b uint8) (X, Y, Z float64) {
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

  //Observer. = 2°, Illuminant = D65
  X = R*0.4124 + G*0.3576 + B*0.1805
  Y = R*0.2126 + G*0.7152 + B*0.0722
  Z = R*0.0193 + G*0.1192 + B*0.9505

  X *= 100
  Y *= 100
  Z *= 100

  return
}

func xyz2rgb(x, y, z float64) (r, g, b uint8) {
  R := x*3.2406 + y*-1.5372 + z*-0.4986
  G := x*-0.9689 + y*1.8758 + z*0.0415
  B := x*0.0557 + y*-0.2040 + z*1.0570

  if R > 0.0031308 {
    R = 1.055*math.Pow(R, (1/2.4)) - 0.055
  } else {
    R = 12.92 * R
  }
  if G > 0.0031308 {
    G = 1.055*math.Pow(G, (1/2.4)) - 0.055
  } else {
    G = 12.92 * G
  }
  if B > 0.0031308 {
    B = 1.055*math.Pow(B, (1/2.4)) - 0.055
  } else {
    B = 12.92 * B
  }

  r = uint8(R * 255.0)
  g = uint8(G * 255.0)
  b = uint8(B * 255.0)

  return
}

func xyz2lab(x, y, z float64) (L, A, B float64) {
  const epsilon float64 = 0.008856
  const kappa float64 = 7.787

  var (
    X = x / RefX
    Y = y / RefY
    Z = z / RefZ
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

  L = ((116.0 * Y) - 16.0)
  A = (500.0 * (X - Y))
  B = (200.0 * (Y - Z))
  return
}

func lab2xyz(l, a, b float64) (x, y, z float64) {
  y = (l + 16.0) / 116.0
  x = a/500.0 + y
  z = y - b/200.0

  y_cube := math.Pow(y, 3)
  x_cube := math.Pow(x, 3)
  z_cube := math.Pow(z, 3)
  if y_cube > 0.008856 {
    y = y_cube
  } else {
    y = (y - 16.0/116.0) / 7.787
  }
  if x_cube > 0.008856 {
    x = x_cube
  } else {
    x = (x - 16.0/116.0) / 7.787
  }
  if z_cube > 0.008856 {
    z = z_cube
  } else {
    z = (z - 16.0/116.0) / 7.787
  }

  x = RefX * (x / 100)
  y = RefY * (y / 100)
  z = RefZ * (z / 100)
  return
}

func Rgb2lab(R, G, B uint8) (l, a, b float64) {
  x, y, z := rgb2xyz(R, G, B)
  return xyz2lab(x, y, z)
}

func Lab2rgb(l, a, b float64) (R, G, B uint8) {
  x, y, z := lab2xyz(l, a, b)
  return xyz2rgb(x, y, z)
}

type Color struct {
  L, A, B float64
}

func (p Color) RGBA() (uint32, uint32, uint32, uint32) {
  R, G, B := Lab2rgb(p.L, p.A, p.B)
  r := uint32(R)
  r |= r << 8
  g := uint32(G)
  g |= g << 8
  b := uint32(B)
  b |= b << 8
  return r, g, b, 0xffff
}

var ColorModel color.Model = color.ModelFunc(labModel)

func labModel(c color.Color) color.Color {
  if _, ok := c.(Color); ok {
    return c
  }
  var (
    r, g, b, _ = c.RGBA()
    L, A, B    = Rgb2lab(uint8(r>>8), uint8(g>>8), uint8(b>>8))
  )
  return Color{L, A, B}
}

type Image struct {
  Pix    []float64
  Stride int
  Rect   image.Rectangle
}

// NewRGBA returns a new RGBA with the given bounds.
func NewImage(r image.Rectangle) *Image {
  w, h := r.Dx(), r.Dy()
  buf := make([]float64, 3*w*h)
  return &Image{buf, 3 * w, r}
}

func (p *Image) ColorModel() color.Model { return ColorModel }

func (p *Image) Bounds() image.Rectangle { return p.Rect }

func (p *Image) At(x, y int) color.Color {
  if !(image.Point{x, y}.In(p.Rect)) {
    return Color{}
  }
  i := p.PixOffset(x, y)
  return Color{p.Pix[i+0], p.Pix[i+1], p.Pix[i+2]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *Image) PixOffset(x, y int) int {
  return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

func (p *Image) Set(x, y int, c color.Color) {
  if !(image.Point{x, y}.In(p.Rect)) {
    return
  }
  i := p.PixOffset(x, y)
  c1 := ColorModel.Convert(c).(Color)
  p.Pix[i+0] = c1.L
  p.Pix[i+1] = c1.A
  p.Pix[i+2] = c1.B
}

func (p *Image) SetLAB(x, y int, c Color) {
  if !(image.Point{x, y}.In(p.Rect)) {
    return
  }
  i := p.PixOffset(x, y)
  p.Pix[i+0] = c.L
  p.Pix[i+1] = c.A
  p.Pix[i+2] = c.B
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *Image) SubImage(r image.Rectangle) image.Image {
  r = r.Intersect(p.Rect)
  // If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
  // either r1 or r2 if the intersection is empty. Without explicitly checking for
  // this, the Pix[i:] expression below can panic.
  if r.Empty() {
    return &Image{}
  }
  i := p.PixOffset(r.Min.X, r.Min.Y)
  return &Image{
    Pix:    p.Pix[i:],
    Stride: p.Stride,
    Rect:   r,
  }
}

func ImageToLab(img image.Image) Image {
  b := img.Bounds()
  canvas := NewImage(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(canvas, canvas.Bounds(), img, b.Min, draw.Src)
  return *canvas
}
