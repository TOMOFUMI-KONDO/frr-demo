package gen

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

const cfgFile = "config.yaml"
const tmplDir = "gen/templates"

type tmplKind int

const (
	tmplHost = iota
	tmplRouter
)

func (t tmplKind) String() string {
	switch t {
	case tmplHost:
		return "host"
	case tmplRouter:
		return "router"
	default:
		panic(fmt.Errorf("unknown tmplKind: %d", t))
	}
}

type Template struct {
	filename string
	kind     tmplKind
}

func (t *Template) path() string {
	return path.Join(tmplDir, t.kind.String(), t.filename)
}

func (t *Template) dstFilename() string {
	i := strings.LastIndex(t.filename, ".tmpl")
	if i == -1 {
		return t.filename
	}

	return t.filename[:i]
}

var templates = []Template{
	{"startup.sh.tmpl", tmplHost},
	{"Dockerfile.tmpl", tmplHost},
}

type Host struct {
	Name    string `yaml:"name"`
	Gateway string `yaml:"gateway"`
}

func (h Host) RemovedGateway() string {
	i := strings.LastIndex(h.Gateway, ".")
	return h.Gateway[:i] + ".1"
}

func Gen(baseDir string) error {
	f, err := os.Open(path.Join(baseDir, cfgFile))
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var hosts struct {
		Hosts []*Host `yaml:"hosts"`
	}
	if err = yaml.Unmarshal(b, &hosts); err != nil {
		return err
	}

	for _, h := range hosts.Hosts {
		dir := path.Join(baseDir, "host", h.Name)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		for _, t := range templates {
			dstFile, err := os.Create(path.Join(dir, t.dstFilename()))
			if err != nil {
				return err
			}

			tmpl := template.Must(template.ParseFiles(t.path()))
			if err = tmpl.Execute(dstFile, h); err != nil {
				return err
			}
		}
	}

	return nil
}
