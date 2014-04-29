package slic

import (
  "image"
  _ "image/jpeg"
  "os"
  "sync"
  "testing"
)

const C float64 = 2000.0 // Compactness
const S int = 40         // Superpixel Size

var (
  inputImageName string      = "test_images/clown_fish.jpg"
  inputImage     image.Image = nil
  s              *SLIC       = nil
)

func loadInputImage() image.Image {
  if inputImage != nil {
    return inputImage
  }

  file, err := os.Open(inputImageName)
  if err != nil {
    panic(err)
  }

  src_img, _, err := image.Decode(file)
  if err != nil {
    panic(err)
  }

  return src_img
}

func BenchmarkInitial(b *testing.B) {
  img := loadInputImage()

  for n := 0; n < b.N; n++ {
    s = MakeSlic(img, C, S)
  }
}

func BenchmarkImageToLab(b *testing.B) {
  img := loadInputImage()

  if s == nil {
    s = MakeSlic(img, C, S)
  }

  for n := 0; n < b.N; n++ {
    s.lvec, s.avec, s.bvec = imageToLab(img)
  }
}

func BenchmarkLabeling(b *testing.B) {
  img := loadInputImage()

  if s == nil {
    s = MakeSlic(img, C, S)
  }

  for n := 0; n < b.N; n++ {
    s.resetDistances()
    var wg sync.WaitGroup
    for n := range s.superpixels {
      superpixel := s.superpixels[n]
      go s.labelPixelsInSuperpixel(superpixel, &wg)
    }
    wg.Wait()
  }
}

func BenchmarkCentroids(b *testing.B) {
  img := loadInputImage()

  if s == nil {
    s = MakeSlic(img, C, S)

    // Just run one dummy iteration
    s.resetDistances()
    var wg sync.WaitGroup
    for i := range s.superpixels {
      superpixel := s.superpixels[i]
      go s.labelPixelsInSuperpixel(superpixel, &wg)
    }
    wg.Wait()
  }

  for n := 0; n < b.N; n++ {
    s.recalculateCentroids()
  }
}

func BenchmarkConnect(b *testing.B) {
  img := loadInputImage()

  if s == nil {
    s = MakeSlic(img, C, S)

    // Just run one dummy iteration
    s.resetDistances()
    var wg sync.WaitGroup
    for i := range s.superpixels {
      superpixel := s.superpixels[i]
      go s.labelPixelsInSuperpixel(superpixel, &wg)
    }
    wg.Wait()
  }

  for n := 0; n < b.N; n++ {
    new_labels := s.enforceLabelConnectivity()
    for i := 0; i < s.sz; i++ {
      s.Labels[i] = new_labels[i]
    }
  }
}
