// Package stream manages Picarto streams, based on artist names. It currently
// wraps itself around rtmpdump and creates notifications if a stream is active.
package stream

import (
	"log"
	"sync"
)

// An ArtistMap contains all the Artists being tracked. They can be looked up via their standard
// string representation.
type ArtistMap map[string]*Artist

// AddArtist adds a new instance of an artist to an ArtistMap after starting an RTMPDump instance for them.
func (a ArtistMap) AddArtist(name string) {

	artist := &Artist{Name: name,
		Online: make(chan ArtistNotification),
	}
	a[name] = artist
	if err := artist.Poll(); err != nil {
		log.Fatal(err)
	}

}

// RemoveArtist removes an artist from an ArtistMap, cleaning up
// after itself beforehand.
// TODO: Check if the artist name is in ArtistMap first
func (a ArtistMap) RemoveArtist(name string) {
	artist := a[name]
	close(artist.Online)
	artist.Stop()
	delete(a, name)
}

// MakeAnnounceChan takes a channel of empty structs as a parameter.
// It takes all the artists stored in the artist map,
// and returns a channel that combines all of their notification
// channels into one.
//
// To indicate that the artist notification channel should be closed, just close the `done` channel.
//
// NOTE: Only one announcement channel should be run for a map at any time.
func (a ArtistMap) MakeAnnounceChan(done <-chan struct{}) <-chan ArtistNotification {
	var wg sync.WaitGroup
	out := make(chan ArtistNotification)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan ArtistNotification) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(a))
	for _, c := range a {
		go output(c.Online)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
