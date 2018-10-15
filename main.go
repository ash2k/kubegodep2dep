package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
)

const (
	kubeBranch     = "release-1.12"
	clientGoBranch = "release-9.0"

	boilerplate = "# Constraints below have been generated using https://github.com/ash2k/kubegodep2dep\n" + //
		"# Do not edit manually\n"
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
	Constraints []constraint
}

type constraint struct {
	Name     string `json:"name"`
	Branch   string `json:"branch,omitempty"`
	Revision string `json:"revision,omitempty"`
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
		// k8s.io/kube-openapi/pkg/util/proto/validation -> k8s.io kube-openapi pkg util/proto/validation
		parts := strings.SplitN(d.ImportPath, "/", 4)
		switch {
		case parts[0] == "github.com":
			depKey = path.Join(parts[:3]...) // join 3 first parts
		case parts[0] == "golang.org" && parts[1] == "x":
			depKey = path.Join(parts[:3]...) // join 3 first parts
		default:
			// This is not ideal, fix as needed
			depKey = path.Join(parts[:2]...) // join 2 first parts

		}
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
	constraints := make([]constraint, 0, len(ordered))
	for _, depkey := range ordered {
		d := deps[depkey]
		var c constraint
		switch {
		case d.branch != "":
			c = constraint{
				Name:   depkey,
				Branch: d.branch,
			}
		case d.revision != "":
			c = constraint{
				Name:     depkey,
				Revision: d.revision,
			}
		default:
			panic(errors.New("unreachable"))
		}
		constraints = append(constraints, c)
	}
	m := depManifest{
		Constraints: constraints,
	}
	data, err := yaml.Marshal(&m)
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
