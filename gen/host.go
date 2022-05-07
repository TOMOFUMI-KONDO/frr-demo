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
	{"startup.sh.tmpl", "gen/templates"},
	{"Dockerfile.tmpl", "gen/templates"},
}

type Host struct {
	Name    string `yaml:"name"`
	Gateway string `yaml:"gateway"`
}

func (h Host) RemovedGateway() string {
	i := strings.LastIndex(h.Gateway, ".")
	return h.Gateway[:i] + ".1"
}

func GenHost(baseDir, cfgPath string) error {
	cfgFile, err := os.Open(path.Join(baseDir, cfgPath))
	if err != nil {
		return err
	}
	defer cfgFile.Close()

	b, err := io.ReadAll(cfgFile)
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
		for _, t := range templates {
			dir := path.Join(baseDir, h.Name)
			if err = os.MkdirAll(dir, 0755); err != nil {
				return err
			}

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
