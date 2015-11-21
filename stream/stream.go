package stream

import (
	// "bytes"
	// "log"
	"os/exec"
	"sync"
	// "time"
)

// ArtistNotification keeps track of all the data needed to show that a certain artist is online.
type ArtistNotification struct {
	Name string
}

// An Artist holds all the data needed to track an artist stream.
type Artist struct {
	Name string

	// Online is used to inform monitors that the artist has started streaming.
	Online chan ArtistNotification

	// Stream holds the RTMPDump info needed to get stream information.
	Stream *exec.Cmd
}

// An ArtistMap contains all the Artists being tracked. They can be looked up via their standard
// string representation.
type ArtistMap map[string]*Artist

// AddArtist adds a new instance of an artist to an ArtistMap after starting an RTMPDump instance for them.
func (a ArtistMap) AddArtist(name string) {

	cmd := exec.Command("rtmpdump", "-r", "rtmp://167.114.157.137:1935/golive/?/"+name)

	artist := &Artist{Name: name,
		Online: make(chan ArtistNotification),
		Stream: cmd}
	a[name] = artist

}

// RemoveArtist removes an artist from an ArtistMap, cleaning up
// after itself beforehand.
// TODO: Check if the artist name is in ArtistMap first
func (a ArtistMap) RemoveArtist(name string) {
	close(a[name].Online)
	a[name].Stream.Process.Kill()
	a[name].Stream.Process.Wait()
	delete(a, name)
}

// MakeAnnounceChan takes all the artists stored in the artist map,
// and returns a channel that combines all of their notification
// channels into one.
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

func monitor(a *exec.Cmd) {

}
