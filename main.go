package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml"
)

const (
	kubeBranch     = "release-1.12"
	clientGoBranch = "release-9.0"

	boilerplate = "# Overrides below have been generated using https://github.com/ash2k/kubegodep2dep\n" + //
		"# Do not edit manually\n"
)

var (
	gopkginVersion = regexp.MustCompile("\\.v\\d+(\\.\\d+){0,2}$")
)

type Dependency struct {
	ImportPath string
	Rev        string // commit
}

type godeps struct {
	Deps []Dependency
}

type dep struct {
	revision string
	branch   string
}

type depManifest struct {
	Overrides []override `toml:"override"`
}

type override struct {
	Name     string `toml:"name"`
	Branch   string `toml:"branch,omitempty"`
	Revision string `toml:"revision,omitempty"`
}

func main() {
	godepsPath := flag.String("godep", "", "Path to Godeps.json file")
	flag.Parse()

	g, err := loadGodepsFile(*godepsPath)
	if err != nil {
		log.Fatal(err)
	}

	deps := predeclaredDeps()
	for _, d := range g.Deps {
		var depKey string
		var n int
		// k8s.io/kube-openapi/pkg/util/proto/validation -> k8s.io kube-openapi pkg util/proto/validation
		parts := strings.SplitN(d.ImportPath, "/", 4)
		switch { // This is not ideal, fix as needed
		case parts[0] == "github.com":
			n = 3
		case parts[0] == "bitbucket.org":
			n = 3
		case parts[0] == "golang.org" && parts[1] == "x":
			n = 3
		case parts[0] == "gopkg.in":
			if gopkginVersion.MatchString(parts[1]) { // gopkg.in/pkg.v3/BLABLA syntax
				n = 2
			} else if gopkginVersion.MatchString(parts[2]) { // gopkg.in/user/pkg.v3/BLABLA syntax
				n = 3
			} else {
				log.Fatalf("Unsupported syntax %s", d.ImportPath)
			}
		default:
			n = 2
		}
		depKey = path.Join(parts[:n]...) // join n first parts

		existingD, ok := deps[depKey]
		if ok {
			log.Printf("Already there: import key %s with import path %s", depKey, d.ImportPath)
			if existingD.revision != "" && d.Rev != existingD.revision {
				log.Fatalf("Revisions don't match for key %s: existing %s, new %s", depKey, existingD.revision, d.Rev)
			}
			continue
		}
		log.Printf("Adding: import key %s for import path %s", depKey, d.ImportPath)
		deps[depKey] = dep{
			revision: d.Rev,
		}
	}
	ordered := make([]string, 0, len(deps))
	for depKey := range deps {
		ordered = append(ordered, depKey)
	}
	sort.Strings(ordered)
	overrides := make([]override, 0, len(ordered))
	for _, depkey := range ordered {
		d := deps[depkey]
		var c override
		switch {
		case d.branch != "":
			c = override{
				Name:   depkey,
				Branch: d.branch,
			}
		case d.revision != "":
			c = override{
				Name:     depkey,
				Revision: d.revision,
			}
		default:
			panic(errors.New("unreachable"))
		}
		overrides = append(overrides, c)
	}
	m := depManifest{
		Overrides: overrides,
	}
	data, err := toml.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n%s\n", boilerplate, data)
}

func loadGodepsFile(path string) (*godeps, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var g godeps
	err = json.NewDecoder(f).Decode(&g)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s: %v", path, err)
	}
	return &g, err
}

// predeclared dependencies for a particular kubernetes version
func predeclaredDeps() map[string]dep {
	return map[string]dep{
		"k8s.io/apiextensions-apiserver": {
			branch: kubeBranch,
		},
		"k8s.io/apimachinery": {
			branch: kubeBranch,
		},
		"k8s.io/apiserver": {
			branch: kubeBranch,
		},
		"k8s.io/client-go": {
			branch: clientGoBranch,
		},
		"k8s.io/api": {
			branch: kubeBranch,
		},
		"k8s.io/code-generator": {
			branch: kubeBranch,
		},
	}
}
