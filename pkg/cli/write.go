package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Writer returns a writer based on the options, defaults to stdout
func Writer(opts *Options) io.Writer {
	var w io.Writer = os.Stdout

	if opts.Discard {
		w = ioutil.Discard
	}

	// Add suoport for writing directly to file with -v output for also writing to stdout
	// at the same time via a T writer

	return w
}

// WriteJSON writes the JSON output to the writer
func WriteJSON(w io.Writer, v interface{}, indent bool) error {
	var marshal func(interface{}) ([]byte, error) = json.Marshal
	if indent {
		marshal = func(interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		}
	}

	b, err := marshal(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

// WriteText writes the string output to the writer
func WriteText(w io.Writer, v interface{}) error {
	if _, err := fmt.Fprintln(w, v); err != nil {
		return err
	}

	return nil
}
