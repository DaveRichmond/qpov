package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ThomasHabets/bsparse/bsp"
	"github.com/ThomasHabets/bsparse/pak"
)

var (
	pakFile = flag.String("pak", "", "Pakfile to use.")
)

func mkdirP(base, mf string) {
	var cparts []string
	for _, part := range strings.Split(mf, "/") {
		cparts = append(cparts, part)
		if err := os.Mkdir(path.Join(base, strings.Join(cparts, "/")), 0755); err != nil {
			//log.Printf("Creating model subdir: %v, continuing...", err)
		}
	}
}

func convert(p pak.MultiPak, args ...string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	outDir := fs.String("out", ".", "Output directory.")
	textures := fs.Bool("textures", false, "Use textures.")
	lights := fs.Bool("lights", true, "Export lights.")
	fs.Parse(args)

	//errors := []string{}
	os.Mkdir(*outDir, 0755)
	for _, mf := range p.List() {
		if path.Ext(mf) != ".bsp" {
			continue
		}
		func() {
			o, err := p.Get(mf)
			if err != nil {
				log.Fatalf("Getting %q: %v", mf, err)
			}

			b, err := bsp.Load(o)
			if err != nil {
				log.Fatalf("Loading %q: %v", mf, err)
			}

			mkdirP(*outDir, mf)
			fn := fmt.Sprintf(path.Join(mf, "level.inc"))
			of, err := os.Create(path.Join(*outDir, fn))
			if err != nil {
				log.Fatalf("Model create of %q fail: %v", fn, err)
			}
			defer of.Close()
			m, err := b.POVTriangleMesh(bsp.ModelPrefix(mf), *textures)
			if err != nil {
				log.Fatalf("Making mesh of %q: %v", mf, err)
			}
			fmt.Fprintln(of, m)
			if *lights {
				fmt.Fprintln(of, b.POVLights())
			}
		}()
	}
}

func info(p pak.MultiPak, args ...string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	//outDir := fs.String("out", ".", "Output directory.")
	//maps := fs.String("maps", ".*", "Regex of maps to convert.")
	fs.Parse(args)
	mapName := fs.Arg(0)
	b, err := p.Get(mapName)
	if err != nil {
		log.Fatalf("Finding map %q: %v", mapName, err)
	}

	m, err := bsp.Load(b)
	if err != nil {
		log.Fatalf("Loading map: %v", err)
	}
	fmt.Printf("Vertices: %v\n", len(m.Raw.Vertex))
	fmt.Printf("Faces: %v\n", len(m.Raw.Face))
	fmt.Printf("Edges: %v\n", len(m.Raw.Edge))
	fmt.Printf("LEdges: %v\n", len(m.Raw.LEdge))
	fmt.Printf("Vertices: %v\n", len(m.Raw.Entities))
	fmt.Printf("MipTexes: %v\n", len(m.Raw.MipTex))
	fmt.Printf("TexInfos: %v\n", len(m.Raw.TexInfo))

	fmt.Printf("Model  Faces\n")
	for n, mod := range m.Raw.Models {
		fmt.Printf("%7d %6v\n", n, mod.FaceNum)
	}
}

func pov(p pak.MultiPak, args ...string) {
	fs := flag.NewFlagSet("pov", flag.ExitOnError)
	//outDir := fs.String("out", ".", "Output directory.")
	//maps := fs.String("maps", ".*", "Regex of maps to convert.")
	fs.Parse(args)
	maps := fs.Arg(0)

	res, err := p.Get(maps)
	if err != nil {
		log.Fatalf("Finding %q: %v", maps, err)
	}

	m, err := bsp.Load(res)

	mesh, err := m.POVTriangleMesh(bsp.ModelPrefix(maps), true)
	if err != nil {
		log.Fatalf("Error getting mesh: %v", err)
	}
	fmt.Println(mesh)
}

func main() {
	flag.Parse()

	pakFile := flag.Arg(0)

	p, err := pak.MultiOpen(pakFile)
	if err != nil {
		log.Fatalf("Opening pakfile %q: %v", pakFile, err)
	}

	switch flag.Arg(1) {
	case "info":
		info(p, flag.Args()[2:]...)
	case "pov":
		pov(p, flag.Args()[2:]...)
	case "convert":
		convert(p, flag.Args()[2:]...)
	default:
		log.Fatalf("Unknown command %q", flag.Arg(1))
	}
}
