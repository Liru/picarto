package stream

import (
	"errors"
	"log"
	"os/exec"
	"time"
)

// ArtistNotification keeps track of all the data needed to show that a certain artist is online.
type ArtistNotification struct {
	Name string
	Time time.Time
}

// An Artist holds all the data needed to track an artist stream.
type Artist struct {
	Name string

	// Online is used to inform monitors that the artist has started streaming.
	Online chan ArtistNotification

	// Stream holds the RTMPDump info needed to get stream information.
	Stream *exec.Cmd

	remove bool
}

// NewArtist makes a new Artist struct with all the necessary information filled out.
// It is mostly used for single artist management. If multiple Artists need to be managed, use an ArtistMap instead.
func NewArtist(n string) *Artist {
	a := &Artist{Name: n,
		Online: make(chan ArtistNotification),
	}

	return a
}

// Poll spawns an instance of rtmpdump to listen on a particular user's stream.
// When it runs, it will wait for rtmpdump to return any output at all, which indicates
// that a stream is active on the particular user account. It will then send a notification
// on the Artist's Online channel.
//
// If rtmpdump is disconnected from the art stream, it will automatically try to reconnect.
//
// Poll will continue to loop and send notifications until Stop is called on the Artist, or RemoveArtist is called on the map containing the artist.
//
// rtmpdump MUST be installed on the system for Poll to work, at least until a native Go solution is made. If it isn't installed, then an error is returned. Otherwise, Poll returns nil.
func (a *Artist) Poll() error {
	if _, err := exec.LookPath("rtmpdump"); err != nil {
		return errors.New("stream: rtmpdump not installed")
	}

	// TODO: Add check to see if already polling.

	go func() {
		for {
			// BUG(Liru): Poll checks if a.remove is set to true in multiple places, possibly allowing bugs. This could probably be cleaned up.
			if a.remove {
				break
			}

			cmd := exec.Command("rtmpdump", "-r", "rtmp://167.114.157.137:1935/golive/?/"+a.Name)

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				// Usually a race condition if this one is called. Most likely that stdout was already set.
				log.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				// The error here usually means that rtmpdump isn't installed.
				log.Fatal(err)
			}

			defer func() {
				cmd.Process.Kill()
				cmd.Process.Wait()
			}()

			a.Stream = cmd

			_, err = stdout.Read(make([]byte, 1))

			// err == nil when something is successfully read from the rtmpdump output.
			// err != nil when we lose a connection [EOF] or rtmpdump is stopped.
			if err != nil {
				log.Println(a.Name, err)
				continue
			}

			if a.remove {
				break
			}

			go func() {
				a.Online <- ArtistNotification{
					Name: a.Name,
					Time: time.Now(),
				}
			}()

			time.Sleep(time.Second * 2) // TODO: Make the Sleep() call in Poll configurable?
		}
	}()

	return nil
}

// Stop stops the rtmpdump instance associated with the Artist.
// It does not close the Online channel that the Artist contains.
func (a *Artist) Stop() {
	a.remove = true
	a.Stream.Process.Kill()
	a.Stream.Process.Wait()
}

func (a *Artist) String() string {
	return a.Name
}
