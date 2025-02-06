package utils

import (
	"compress/gzip"
	"context"
	"encoding/xml"
	"io"
	"os"
)

type Document struct {
	Title string `xml:"title"`
	URL   string `xml:"url"`
	Text  string `xml:"abstract"`
	ID    int    `xml:"-"`
}

func StreamDocuments(ctx context.Context, path string) (<-chan Document, <-chan error) {
	out := make(chan Document, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		f, err := os.Open(path)
		if err != nil {
			errCh <- err
			return
		}
		defer f.Close()

		gz, err := gzip.NewReader(f)
		if err != nil {
			errCh <- err
			return
		}
		defer gz.Close()

		dec := xml.NewDecoder(gz)
		var id int

		for {
			select {
			case <-ctx.Done():
				return
			default:
				tok, err := dec.Token()
				if err == io.EOF {
					return
				}
				if err != nil {
					errCh <- err
					return
				}

				if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "doc" {
					var doc Document
					if err := dec.DecodeElement(&doc, &se); err != nil {
						errCh <- err
						continue
					}
					doc.ID = id
					id++

					select {
					case out <- doc:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out, errCh
}
