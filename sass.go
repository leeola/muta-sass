package sass

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/suapapa/go_sass"

	"github.com/leeola/muta"
	"github.com/leeola/muta/logging"
)

const defaultName string = "sass.SassStreamer"

type Options struct {
	// The name of this streamer, returned by Name()
	Name string
}

func NewSassStreamer(opts Options) *SassStreamer {
	if opts.Name == "" {
		opts.Name = defaultName
	}

	return &SassStreamer{
		Opts: opts,
	}
}

type SassStreamer struct {
	// The options for this streamer
	Opts Options

	// Stream buffer
	buffer bytes.Buffer
	// If the file is sass, then we set it to this to buffer it
	currentFi *muta.FileInfo
	// If the file is not sass, we set it to this to ignore it
	ignoreFi *muta.FileInfo
	// Compiler
	compiler sass.Compiler
}

func (s *SassStreamer) IsIgnoreFile(fi *muta.FileInfo) bool {
	if fi == s.ignoreFi {
		return true
	} else {
		return false
	}
}

func (s *SassStreamer) IsNewFile(fi *muta.FileInfo) bool {
	if fi != s.currentFi {
		s.currentFi = fi
		return true
	} else {
		return false
	}
}

func (s *SassStreamer) Name() string {
	return s.Opts.Name
}

func (s *SassStreamer) Stream(fi *muta.FileInfo, chunk []byte) (
	*muta.FileInfo, []byte, error) {
	switch {
	case fi == nil:
		return nil, nil, nil

	case s.IsIgnoreFile(fi):
		return fi, chunk, nil

	case s.IsNewFile(fi):
		switch filepath.Ext(fi.Name) {
		// I have yet to figure out how to apply the following option:
		// https://github.com/sass/libsass/blob/master/sass_context.cpp#L57
		// to the compiler. Thus, all indented sass syntax will fail.
		case ".sass":
			s.ignoreFi = fi
			logging.Warnf([]string{s.Opts.Name},
				"File '%s/%s' is being ignored. .sass syntax "+
					"is not currently supported", fi.Path, fi.Name)
		case ".scss":
			fi.Name = strings.TrimSuffix(fi.Name, ".scss") + ".css"
		default:
			s.ignoreFi = fi
		}
		return s.Stream(fi, chunk)

	// If chunk is nil, we're at EOF. Read our buffer, compile it,
	// and stream it.
	case chunk == nil:
		if s.buffer.Len() == 0 {
			return fi, nil, nil
		}
		source, _ := ioutil.ReadAll(&s.buffer)

		// We're setting the include paths for each file, so that
		// they can import relatively from their local directory
		s.compiler.IncludePaths = []string{fi.OriginalPath}

		var compiled string
		var err error
		compiled, err = s.compiler.Compile(string(source))
		return fi, []byte(compiled), err

	// File is sass, and chunk isn't nil. Buffer the data
	default:
		_, err := s.buffer.Write(chunk)
		return nil, nil, err
	}
}

func Sass() muta.Streamer {
	// This will likely be implemented into the muta core for ease of use,
	// in the future
	//return muta.FilterExtStream(NewSassStreamer(Options{}),
	//	".scss", ".scss")
	return NewSassStreamer(Options{})
}
