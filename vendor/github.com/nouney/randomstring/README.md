# Randomstring 

[![Go Report Card](https://goreportcard.com/badge/github.com/nouney/randomstring)](https://goreportcard.com/report/github.com/nouney/randomstring)
[![GoDoc](https://godoc.org/github.com/nouney/randomstring?status.svg)](https://godoc.org/github.com/nouney/randomstring)

A small Golang package to easely generate random strings.

The code is based on this awesome [StackOverflow answer](https://stackoverflow.com/a/31832326/2432477) by [icza](https://stackoverflow.com/users/1705598/icza).

## Installation

To install `randomstring`, run the following command:

 `$ go get github.com/nouney/randomstring`

## Example

```golang
package main

import (
    "fmt"
    "log"

    "github.com/nouney/randomstring"
)

func main() {
    // Generate an alphanum random string of 16 chars
    str := randomstring.Generate(4)
    fmt.Println("Alphanum:", str)

    // Create a generator using digits
    rsg, err := randomstring.NewGenerator(randomstring.CharsetNum)
    if err != nil {
        log.Panic(err)
    }
    str = rsg.Generate(3)
    fmt.Println("Num:", str)

    // Create a generator using digits and lowercase alphabet
    rsg, err = randomstring.NewGenerator(randomstring.CharsetNum, randomstring.CharsetAlphaLow)
    if err != nil {
        log.Panic(err)
    }
    str = rsg.Generate(8)
    fmt.Println("Alphanumlow:", str)

    // Create a generator with a custom charset
    rsg, err = randomstring.NewGenerator("AbCdEfGhIjKlMnOpQrStUvWxYz")
    if err != nil {
        log.Panic(err)
    }
    str = rsg.Generate(18)
    fmt.Println("Custom charset:", str)
}

```

## Default charsets

- `CharsetAlphaLow`: "abcdefghijklmnopqrstuvwxyz"
- `CharsetAlphaUp`: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
- `CharsetNum`: "0123456789"