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

const (
	cfgFile = "config.yaml"
)

type tmplKind int

const (
	tmplHost tmplKind = iota
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
	dir, filename string
	kind          tmplKind
}

func NewTemplate(filename string, kind tmplKind) *Template {
	return &Template{
		dir:      "gen/templates",
		filename: filename,
		kind:     kind,
	}
}

func (t *Template) path() string {
	return path.Join(t.dir, t.kind.String(), t.filename)
}

func (t *Template) dstFilename() string {
	i := strings.LastIndex(t.filename, ".tmpl")
	if i == -1 {
		return t.filename
	}

	return t.filename[:i]
}

var tmplsHost = []*Template{
	NewTemplate("startup.sh.tmpl", tmplHost),
	NewTemplate("Dockerfile.tmpl", tmplHost),
}

var tmplsRouter = []*Template{
	NewTemplate("vtysh.conf.tmpl", tmplRouter),
	NewTemplate("daemons.tmpl", tmplRouter),
	NewTemplate("frr.conf.tmpl", tmplRouter),
	NewTemplate("Dockerfile.tmpl", tmplRouter),
}

type Host struct {
	Name    string `yaml:"name"`
	Gateway string `yaml:"gateway"`
}

func (h Host) RemovedGateway() string {
	i := strings.LastIndex(h.Gateway, ".")
	return h.Gateway[:i] + ".1"
}

type Router struct {
	Name          string          `yaml:"name"`
	Lo            string          `yaml:"lo"`
	IpPrefixLists []*IpPrefixList `yaml:"ip_prefix_lists,omitempty"`
	RouteMaps     []*RouteMap     `yaml:"route_maps,omitempty"`
	BGP           *BGP            `yaml:"bgp,omitempty"`
	//OSPF         *OSPF           `yaml:"ospf,omitempty"`
}

type IpPrefixList struct {
	Name string `yaml:"name"`
	Cidr string `yaml:"cidr"`
}

type RouteMap struct {
	Name            string `yaml:"name"`
	MatchPrefixList string `yaml:"match_prefix_list"`
}

type BGP struct {
	As        string      `yaml:"as"`
	Network   string      `yaml:"network"`
	Neighbors []*Neighbor `yaml:"neighbors"`
}

type Neighbor struct {
	Addr        string `yaml:"addr"`
	As          string `yaml:"as"`
	Weight      string `yaml:"weight,omitempty"`
	RouteMapIn  string `yaml:"route_map_in,omitempty"`
	RouteMapOut string `yaml:"route_map_out"`
}

func (r *Router) Bgpd() bool {
	return r.BGP != nil
}

func (r *Router) Ospfd() bool {
	//return r.OSPF != nil
	return false
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

	var cfg struct {
		Hosts   []*Host   `yaml:"hosts"`
		Routers []*Router `yaml:"routers"`
	}
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return err
	}

	for _, h := range cfg.Hosts {
		dir := path.Join(baseDir, tmplHost.String(), h.Name)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		for _, t := range tmplsHost {
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

	for _, r := range cfg.Routers {
		dir := path.Join(baseDir, tmplRouter.String(), r.Name)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		for _, t := range tmplsRouter {
			dstFile, err := os.Create(path.Join(dir, t.dstFilename()))
			if err != nil {
				return err
			}

			tmpl := template.Must(template.ParseFiles(t.path()))
			if err = tmpl.Execute(dstFile, r); err != nil {
				return err
			}
		}
	}

	return nil
}
