// generates

package main

import (
  "github.com/agoravoting/agora-http-go/middleware"
  "fmt"
  "flag"
)

func main() {
  var secret = flag.String("secret", "elpastelestaenelhorno", "secret")
  var msg = flag.String("msg", "whatever", "message to sign")
  flag.Parse()

  fmt.Printf("secret used: '%s'\n", *secret)
  fmt.Printf("msg used: '%s'\n", *msg)
  fmt.Println(middleware.AuthHeader(*msg, *secret))
}