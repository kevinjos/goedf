package main

import (
  "fmt"
  "github.com/kevinjos/goedf"
)

func main() {

  h, _ := edf.NewHeader(8)
  om := h.GetOffsetMap()
  fmt.Printf("%+v\n", om)

  // var a [][16]byte
  a := make([][16]byte, 1)
  fmt.Printf("yes, this is the code that is running sucka: %d\n", len(a[200000]))
}
