package gen

import (
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Template struct {
	filename string
	dir      string
}

func (t *Template) path() string {
	return path.Join(t.dir, t.filename)
}

func (t *Template) dstFilename() string {
	i := strings.LastIndex(t.filename, ".tmpl")
	if i == -1 {
		return t.filename
	}

	return t.filename[:i]
}

var templates = []Template{
	{"startup.sh.tmpl", "bgp_4as/host"},
	{"Dockerfile.tmpl", "bgp_4as/host"},
}

type Host struct {
	Name    string `yaml:"name"`
	Gateway string `yaml:"gateway"`
}

func (h Host) RemovedGateway() string {
	i := strings.LastIndex(h.Gateway, ".")
	return h.Gateway[:i] + ".1"
}

func Gen(cfgPath string) {
	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		panic(err)
	}
	defer cfgFile.Close()

	b, err := io.ReadAll(cfgFile)
	if err != nil {
		panic(err)
	}

	var hosts struct {
		Hosts []*Host `yaml:"hosts"`
	}
	if err = yaml.Unmarshal(b, &hosts); err != nil {
		panic(err)
	}

	for _, h := range hosts.Hosts {
		for _, t := range templates {
			dir := path.Join(t.dir, h.Name)
			if err = os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}

			dstFile, err := os.Create(path.Join(dir, t.dstFilename()))
			if err != nil {
				panic(err)
			}

			tmpl := template.Must(template.ParseFiles(t.path()))
			if err := tmpl.Execute(dstFile, h); err != nil {
				panic(err)
			}
		}
	}
}
