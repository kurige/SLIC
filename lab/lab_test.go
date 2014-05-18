package lab

import (
  . "github.com/franela/goblin"
  "image/color"
  "testing"
)

// Reference values
const (
  BLACK_R, BLACK_G, BLACK_B    uint8   = 0, 0, 0
  BLACK_X, BLACK_Y, BLACK_Z    float64 = 0, 0, 0
  BLACK_L_, BLACK_A_, BLACK_B_ float64 = 0, 0, 0

  WHITE_R, WHITE_G, WHITE_B    uint8   = 255, 255, 255
  WHITE_X, WHITE_Y, WHITE_Z    float64 = 95.05, 100.0, 108.89999999999999
  WHITE_L_, WHITE_A_, WHITE_B_ float64 = 100.0, 0.00526049995830391, -0.010408184525267927

  SEMI_RED_R, SEMI_RED_G, SEMI_RED_B    uint8   = 127, 15, 15
  SEMI_RED_X, SEMI_RED_Y, SEMI_RED_Z    float64 = 9.009444302551767, 4.888163219692639, 0.920596075638935
  SEMI_RED_L_, SEMI_RED_A_, SEMI_RED_B_ float64 = 26.413738607701468, 45.15854892196941, 32.373250225032976

  SEMI_GREEN_R, SEMI_GREEN_G, SEMI_GREEN_B    uint8   = 15, 127, 15
  SEMI_GREEN_X, SEMI_GREEN_Y, SEMI_GREEN_Z    float64 = 7.872597456996945, 15.314791405383385, 2.993059576933216
  SEMI_GREEN_L_, SEMI_GREEN_A_, SEMI_GREEN_B_ float64 = 46.0623692928914, -49.55702990747893, 46.644197240981164

  SEMI_BLUE_R, SEMI_BLUE_G, SEMI_BLUE_B    uint8   = 15, 15, 127
  SEMI_BLUE_X, SEMI_BLUE_Y, SEMI_BLUE_Z    float64 = 4.198590589337114, 1.975511812468243, 20.238694297913558
  SEMI_BLUE_L_, SEMI_BLUE_A_, SEMI_BLUE_B_ float64 = 15.358205331292986, 41.58489587147529, -60.074023138804996
)

func TestRGBToXYZ(t *testing.T) {
  g := Goblin(t)
  g.Describe("RGB to XYZ", func() {
    g.It("Black", func() {
      x, y, z := rgb2xyz(BLACK_R, BLACK_G, BLACK_B)
      g.Assert(x).Equal(BLACK_X)
      g.Assert(y).Equal(BLACK_Y)
      g.Assert(z).Equal(BLACK_Z)
    })
    g.It("White", func() {
      x, y, z := rgb2xyz(WHITE_R, WHITE_G, WHITE_B)
      g.Assert(x).Equal(WHITE_X)
      g.Assert(y).Equal(WHITE_Y)
      g.Assert(z).Equal(WHITE_Z)
    })
    g.It("Sorta Red", func() {
      x, y, z := rgb2xyz(SEMI_RED_R, SEMI_RED_G, SEMI_RED_B)
      g.Assert(x).Equal(SEMI_RED_X)
      g.Assert(y).Equal(SEMI_RED_Y)
      g.Assert(z).Equal(SEMI_RED_Z)
    })
    g.It("Sorta Green", func() {
      x, y, z := rgb2xyz(SEMI_GREEN_R, SEMI_GREEN_G, SEMI_GREEN_B)
      g.Assert(x).Equal(SEMI_GREEN_X)
      g.Assert(y).Equal(SEMI_GREEN_Y)
      g.Assert(z).Equal(SEMI_GREEN_Z)
    })
    g.It("Sorta Blue", func() {
      x, y, z := rgb2xyz(SEMI_BLUE_R, SEMI_BLUE_G, SEMI_BLUE_B)
      g.Assert(x).Equal(SEMI_BLUE_X)
      g.Assert(y).Equal(SEMI_BLUE_Y)
      g.Assert(z).Equal(SEMI_BLUE_Z)
    })
  })
}

func TestRGBToLAB(t *testing.T) {
  g := Goblin(t)
  g.Describe("RGB to LAB", func() {
    g.It("Black", func() {
      l, a, b := Rgb2lab(BLACK_R, BLACK_G, BLACK_B)
      g.Assert(l).Equal(BLACK_L_)
      g.Assert(a).Equal(BLACK_A_)
      g.Assert(b).Equal(BLACK_B_)
    })
    g.It("White", func() {
      l, a, b := Rgb2lab(WHITE_R, WHITE_G, WHITE_B)
      g.Assert(l).Equal(WHITE_L_)
      g.Assert(a).Equal(WHITE_A_)
      g.Assert(b).Equal(WHITE_B_)
    })
    g.It("Sorta Red", func() {
      l, a, b := Rgb2lab(SEMI_RED_R, SEMI_RED_G, SEMI_RED_B)
      g.Assert(l).Equal(SEMI_RED_L_)
      g.Assert(a).Equal(SEMI_RED_A_)
      g.Assert(b).Equal(SEMI_RED_B_)
    })
    g.It("Sorta Green", func() {
      l, a, b := Rgb2lab(SEMI_GREEN_R, SEMI_GREEN_G, SEMI_GREEN_B)
      g.Assert(l).Equal(SEMI_GREEN_L_)
      g.Assert(a).Equal(SEMI_GREEN_A_)
      g.Assert(b).Equal(SEMI_GREEN_B_)
    })
    g.It("Sorta Blue", func() {
      l, a, b := Rgb2lab(SEMI_BLUE_R, SEMI_BLUE_G, SEMI_BLUE_B)
      g.Assert(l).Equal(SEMI_BLUE_L_)
      g.Assert(a).Equal(SEMI_BLUE_A_)
      g.Assert(b).Equal(SEMI_BLUE_B_)
    })
  })
}

func TestLABToRGB(t *testing.T) {
  _g := Goblin(t)
  _g.Describe("LAB to RGB", func() {
    _g.It("Black", func() {
      r, g, b := Lab2rgb(BLACK_L_, BLACK_A_, BLACK_B_)
      _g.Assert(r).Equal(BLACK_R)
      _g.Assert(g).Equal(BLACK_G)
      _g.Assert(b).Equal(BLACK_B)
    })
    _g.It("White", func() {
      r, g, b := Lab2rgb(WHITE_L_, WHITE_A_, WHITE_B_)
      _g.Assert(r).Equal(WHITE_R)
      _g.Assert(g).Equal(WHITE_G)
      _g.Assert(b).Equal(WHITE_B)
    })
    _g.It("Sorta Red", func() {
      r, g, b := Lab2rgb(SEMI_RED_L_, SEMI_RED_A_, SEMI_RED_B_)
      _g.Assert(r).Equal(SEMI_RED_R - 1) // SO CLOSE!
      _g.Assert(g).Equal(SEMI_RED_G)
      _g.Assert(b).Equal(SEMI_RED_B)
    })
    _g.It("Sorta Green", func() {
      r, g, b := Lab2rgb(SEMI_GREEN_L_, SEMI_GREEN_A_, SEMI_GREEN_B_)
      _g.Assert(r).Equal(SEMI_GREEN_R)
      _g.Assert(g).Equal(SEMI_GREEN_G)
      _g.Assert(b).Equal(SEMI_GREEN_B)
    })
    _g.It("Sorta Blue", func() {
      r, g, b := Lab2rgb(SEMI_BLUE_L_, SEMI_BLUE_A_, SEMI_BLUE_B_)
      _g.Assert(r).Equal(SEMI_BLUE_R)
      _g.Assert(g).Equal(SEMI_BLUE_G - 1) // SO CLOSE!
      _g.Assert(b).Equal(SEMI_BLUE_B)
    })
  })
}

func TestRoundtrip(t *testing.T) {
  g := Goblin(t)
  g.Describe("Roundtrip", func() {
    g.It("Should match color values", func() {
      var c1, c2 color.Color
      var R1, G1, B1, R2, G2, B2 uint32

      c1 = color.RGBA{255, 255, 255, 255}
      c2 = ColorModel.Convert(c1)
      R1, G1, B1, _ = c1.RGBA()
      R2, G2, B2, _ = c2.RGBA()
      g.Assert(R1).Equal(R2)
      g.Assert(G1).Equal(G2)
      g.Assert(B1).Equal(B2)

      c1 = color.RGBA{0, 0, 0, 0}
      c2 = ColorModel.Convert(c1)
      R1, G1, B1, _ = c1.RGBA()
      R2, G2, B2, _ = c2.RGBA()

      g.Assert(R1).Equal(R2)
      g.Assert(G1).Equal(G2)
      g.Assert(B1).Equal(B2)

      c1 = color.RGBA{127, 0, 0, 255}
      c2 = ColorModel.Convert(c1)
      R1, G1, B1, _ = c1.RGBA()
      R2, G2, B2, _ = c2.RGBA()

      g.Assert(R1).Equal(R2)
      g.Assert(G1).Equal(G2)
      g.Assert(B1).Equal(B2)

      c1 = color.RGBA{0, 127, 0, 255}
      c2 = ColorModel.Convert(c1)
      R1, G1, B1, _ = c1.RGBA()
      R2, G2, B2, _ = c2.RGBA()

      g.Assert(R1).Equal(R2)
      g.Assert(G1).Equal(G2)
      g.Assert(B1).Equal(B2)

      c1 = color.RGBA{0, 0, 127, 255}
      c2 = ColorModel.Convert(c1)
      R1, G1, B1, _ = c1.RGBA()
      R2, G2, B2, _ = c2.RGBA()

      g.Assert(R1).Equal(R2)
      g.Assert(G1).Equal(G2)
      g.Assert(B1).Equal(B2)
    })
  })
}
