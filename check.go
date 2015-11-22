// This is a sample program to demonstrate how to use the Stream library to check if an artist is online.
// To run it, try:
//   go run check.go <artist>
// and you should have an answer within 5 seconds.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/liru/picarto/stream"
)

func main() {

	if len(os.Args) == 1 {
		os.Stderr.WriteString("You need to specify an artist to check if they're online.\n")
		os.Stderr.WriteString("Try:\t" + os.Args[0] + " <artist>\n")
		return
	}

	a := os.Args[1]

	artist := stream.NewArtist(a)

	fmt.Println("Checking if", artist, "is online...")
	artist.Poll()

	select {
	case <-artist.Online:
		fmt.Println(artist, "is online!")
	case <-time.After(time.Second * 5):
		fmt.Println(artist, "is not online :(")
	}

}
