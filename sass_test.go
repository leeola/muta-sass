package sass

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/leeola/muta"
	"github.com/leeola/muta/logging"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	logging.SetLevel(logging.ERROR)
}

func TestSass(t *testing.T) {
	tmpDir := filepath.Join("_test", "fixtures")

	Convey("Should not buffer files that aren't sass", t, func() {
		s := NewSassStreamer(Options{}).Stream
		oFi := muta.NewFileInfo("foo.bar")
		fi, chunk, err := s(oFi, []byte("some data"))
		So(err, ShouldBeNil)
		So(oFi, ShouldEqual, fi)
		So(string(chunk), ShouldEqual, "some data")
	})

	// Indented sass is currently not supported
	// See the comment around line ~64
	Convey("Should not buffer .sass files", t, func() {
		s := NewSassStreamer(Options{}).Stream
		oFi := muta.NewFileInfo("foo.sass")
		fi, chunk, err := s(oFi, []byte("some data"))
		So(err, ShouldBeNil)
		So(oFi, ShouldEqual, fi)
		So(string(chunk), ShouldEqual, "some data")
	})
	/*
			Convey("Should compile sass", t, func() {
				s := NewSassStreamer(Options{}).Stream
				oFi := muta.NewFileInfo("foo.sass")
				fi, chunk, err := s(oFi, []byte(`
		$font-stack:    Helvetica, sans-serif
		$primary-color: #333

		body
		  font: 100% $font-stack
		  color: $primary-color
		`))
				So(err, ShouldBeNil)
				So(fi, ShouldBeNil)
				So(chunk, ShouldBeNil)
				fi, chunk, err = s(oFi, nil)
				So(err, ShouldBeNil)
				So(fi, ShouldEqual, oFi)
				So(string(chunk), ShouldEqual, `body {
		  font: 100% Helvetica, sans-serif;
		  color: #333333; }
		`)
			})
	*/

	Convey("Should compile scss", t, func() {
		s := NewSassStreamer(Options{}).Stream
		oFi := muta.NewFileInfo("foo.scss")
		fi, chunk, err := s(oFi, []byte(`
$font-stack:    Helvetica, sans-serif;
$primary-color: #333;

body {
  font: 100% $font-stack;
  color: $primary-color;
}
`))
		So(err, ShouldBeNil)
		So(fi, ShouldBeNil)
		So(chunk, ShouldBeNil)
		fi, chunk, err = s(oFi, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldEqual, oFi)
		So(string(chunk), ShouldEqual, `body {
  font: 100% Helvetica, sans-serif;
  color: #333333; }
`)
	})

	Convey("Should compile scss from many chunks", t, func() {
		s := NewSassStreamer(Options{}).Stream
		oFi := muta.NewFileInfo("foo.scss")
		bs := bytes.Split([]byte(`
$font-stack:    Helvetica, sans-serif;
$primary-color: #333;

body {
  font: 100% $font-stack;
  color: $primary-color;
}
`), []byte{'\n'})
		for _, line := range bs {
			fi, chunk, err := s(oFi, append(line, '\n'))
			So(err, ShouldBeNil)
			So(fi, ShouldBeNil)
			So(chunk, ShouldBeNil)
		}
		fi, chunk, err := s(oFi, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldEqual, oFi)
		So(string(chunk), ShouldEqual, `body {
  font: 100% Helvetica, sans-serif;
  color: #333333; }
`)
	})

	Convey("Should return EOF when file is immediate EOF", t, func() {
		s := NewSassStreamer(Options{}).Stream
		oFi := muta.NewFileInfo("foo.scss")
		fi, chunk, err := s(oFi, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldEqual, oFi)
		So(chunk, ShouldBeNil)
	})

	Convey("Should change the file extension to css", t, func() {
		s := NewSassStreamer(Options{}).Stream
		fi, _, _ := s(muta.NewFileInfo("foo.scss"), nil)
		So(filepath.Ext(fi.Name), ShouldEqual, ".css")
	})

	Convey("Should allow sass to import relative files", t, func() {
		s := NewSassStreamer(Options{}).Stream
		// Note that these two files don't exist, but that doesn't
		// matter because this plugin doesn't load files. We're using
		// the path to ensure imports in tmpDir work
		oFi1 := muta.NewFileInfo(filepath.Join(tmpDir, "foo.scss"))
		oFi2 := muta.NewFileInfo(filepath.Join(tmpDir, "bar.scss"))
		s(oFi1, []byte(`
@import "style";
`))
		_, _, err := s(oFi1, nil)
		So(err, ShouldBeNil)
		// Now ensure that bad imports are properly returning errors
		// (If they are not, then this test may not be correct)
		s(oFi2, []byte(`
@import "thisdoesntexist";
`))
		_, _, err = s(oFi2, nil)
		So(err, ShouldNotBeNil)
	})
}
