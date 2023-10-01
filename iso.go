package podcast_cdr_manager

import (
	"bytes"
	"fmt"
	"github.com/kdomanski/iso9660"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

func (d *Disk) GenerateIso(fn string, includeProfileData bool, profile *Profile) error {

	casts, err := profile.GetCastsByDiskName(d.Name)
	if err != nil {
		return fmt.Errorf("failed to find casts by disk: %s", err)
	}

	if len(casts) == 0 {
		return fmt.Errorf("found 0 casts for disk")
	}

	writer, err := iso9660.NewWriter()
	if err != nil {
		return fmt.Errorf("create writer: %s", err)
	}
	defer func(writer *iso9660.ImageWriter) {
		err := writer.Cleanup()
		if err != nil {
			log.Printf("Cleanup error: %s", err)
		}
	}(writer)

	if includeProfileData && profile != nil {
		data, err := profile.ProfileData()
		if err != nil {
			return fmt.Errorf("create profile data: %s", err)
		}
		err = writer.AddFile(bytes.NewReader(data), profile.Name+".yaml")
		if err != nil {
			return fmt.Errorf("add file: %s", err)
		}
	}

	type Message struct {
		Err      error
		Filename string
		Bytes    []byte
	}

	mainThread := make(chan *Message)
	writeToDisk := make(chan *Message)
	concurrentLimit := make(chan struct{}, 4)
	go func() {
		for msg := range writeToDisk {
			if msg == nil {
				continue
			}
			if msg.Err != nil {
				mainThread <- msg
				break
			}
			err = writer.AddFile(bytes.NewReader(msg.Bytes), msg.Filename)
			if err != nil {
				mainThread <- &Message{Err: fmt.Errorf("add file: %s", err)}
				break
			}
			log.Printf("Written: %s", msg.Filename)
		}
		close(mainThread)
	}()
	wg := sync.WaitGroup{}
	wg.Add(len(casts))
	for i := range casts {
		go func(cast *Cast) {
			defer wg.Done()
			concurrentLimit <- struct{}{}
			defer func() {
				<-concurrentLimit
			}()
			log.Printf("Downloading %s", cast.MpegLink)
			r, err := http.Get(cast.MpegLink)
			if err != nil {
				log.Printf("Errpr Downloading %s %s", cast.MpegLink, err)
				writeToDisk <- &Message{
					Err: err,
				}
				return
			}
			defer func() {
				err := r.Body.Close()
				if err != nil {
					log.Printf("Error closing http connection: %s", err)
				}
			}()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Errpr Downloading %s %s", cast.MpegLink, err)
				writeToDisk <- &Message{
					Err: err,
				}
				return
			}
			u, err := url.Parse(cast.MpegLink)
			if err != nil {
				log.Printf("Errpr Downloading %s %s", cast.MpegLink, err)
				writeToDisk <- &Message{
					Err: fmt.Errorf("url parser: %w", err),
				}
				return
			}
			log.Printf("Downloaded %s", cast.MpegLink)
			writeToDisk <- &Message{
				Bytes:    b,
				Filename: path.Base(u.Path),
			}
		}(casts[i])
	}
	wg.Wait()
	close(writeToDisk)
	msg := <-mainThread

	if msg != nil && msg.Err != nil {
		return fmt.Errorf("downloading: %w", err)
	}

	outputFile, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("create file: %s", err)
	}

	err = writer.WriteTo(outputFile, d.Name)
	if err != nil {
		return fmt.Errorf("write ISO image: %s", err)
	}

	err = outputFile.Close()
	if err != nil {
		return fmt.Errorf("close output file: %s", err)
	}

	return nil
}
