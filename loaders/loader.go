package loaders

import (
	"fmt"
	"net/url"
	"os"

	"github.com/rehabstudio/oneill/containerdefs"
)

// GetLoader parses a given URI and returns an appropriate loader. For now
// this always returns our default (and only) loader, but could be easily
// expanded to load container definitions from a remote location, or from a
// single file.
func GetLoader(uriStr string) (containerdefs.DefinitionLoader, error) {

	// parse uri so we can decide which loader to use
	uri, err := url.Parse(uriStr)
	if err != nil {
		return &LoaderDirectory{rootDirectory: ""}, err
	}

	// return an appropriate file or directory loader
	if uri.Scheme == "file" {
		// check if path exists
		src, err := os.Stat(uri.Path)
		if err != nil {
			return &LoaderDirectory{rootDirectory: ""}, err
		}
		// check if path is a single file or a directory
		if src.IsDir() {
			return &LoaderDirectory{rootDirectory: uri.Path}, nil
		} else {
			return &LoaderFile{path: uri.Path}, nil
		}
	}

	// return the http loader
	if uri.Scheme == "http" || uri.Scheme == "https" {
		return &LoaderURL{url: uriStr}, nil
	}

	// return the http loader
	if uri.Scheme == "stdin" {
		return &LoaderStdin{}, nil
	}

	// couldn't find a loader, return an error :(
	err = fmt.Errorf("Unable find matching loader for definitions uri: %s", uriStr)
	return &LoaderDirectory{rootDirectory: ""}, err
}
