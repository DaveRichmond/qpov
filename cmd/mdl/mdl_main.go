package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ThomasHabets/bsparse/mdl"
	"github.com/ThomasHabets/bsparse/pak"
)

var (
	model   = flag.String("model", "progs/ogre.mdl", "Model to read.")
	command = flag.String("c", "show", "Command (convert, show)")
	outDir  = flag.String("out", ".", "Output directory.")
)

func frameName(mf string, frame int) string {
	re := regexp.MustCompile(`[/.-]`)
	return fmt.Sprintf("demprefix_%s_%d", re.ReplaceAllString(mf, "_"), frame)
}

func convert(p pak.MultiPak) {
	errors := []string{}
	os.Mkdir(*outDir, 0755)
	for _, mf := range p.List() {
		if path.Ext(mf) != ".mdl" {
			continue
		}
		func() {
			o, err := p.Get(mf)
			if err != nil {
				log.Fatalf("Getting %q: %v", mf, err)
			}

			m, err := mdl.Load(o)
			if err != nil {
				log.Printf("Loading %q: %v", mf, err)
				errors = append(errors, mf)
				return
			}
			var cparts []string
			for _, part := range strings.Split(mf, "/") {
				cparts = append(cparts, part)
				if err := os.Mkdir(path.Join(*outDir, strings.Join(cparts, "/")), 0755); err != nil {
					//log.Printf("Creating model subdir: %v, continuing...", err)
				}
			}
			fn := fmt.Sprintf(path.Join(mf, "model.inc"))
			of, err := os.Create(path.Join(*outDir, fn))
			if err != nil {
				log.Fatalf("Model create of %q fail: %v", fn, err)
			}
			defer of.Close()
			for n := range m.Frames {
				fmt.Fprintf(of, "#macro %s(pos, rot, skin)\n%s\n#end\n", frameName(mf, n), m.POVFrameID(n, "skin"))
			}

			for n, skin := range m.Skins {
				of, err := os.Create(path.Join(*outDir, mf, fmt.Sprintf("skin_%d.png", n)))
				if err != nil {
					log.Fatalf("Skin create of %q fail: %v", fn, err)
				}
				defer of.Close()
				if err := (&png.Encoder{}).Encode(of, skin); err != nil {
					log.Fatalf("Encoding skin to png: %v", err)
				}
			}
		}()
	}
	fmt.Printf("Failed to convert %d models:\n  %s\n", len(errors), strings.Join(errors, "\n  "))
}

func show(p pak.MultiPak) {
	h, err := p.Get(*model)
	if err != nil {
		log.Fatalf("Unable to get %q: %v", *model, err)
	}

	m, err := mdl.Load(h)
	if err != nil {
		log.Fatalf("Unable to load %q: %v", *model, err)
	}

	fmt.Printf("Filename: %s\n  Triangles: %v\n", *model, len(m.Triangles))
	fmt.Printf("Skins: %v\n", len(m.Skins))
	fmt.Printf("  %6s %16s\n", "Frame#", "Name")
	for n, f := range m.Frames {
		fmt.Printf("  %6d %16s\n", n, f.Name)
	}
}

func triangles(p pak.MultiPak) {
	h, err := p.Get(*model)
	if err != nil {
		log.Fatalf("Unable to get %q: %v", *model, err)
	}

	m, err := mdl.Load(h)
	if err != nil {
		log.Fatalf("Unable to load %q: %v", *model, err)
	}

	for n, _ := range m.Frames {
		fmt.Printf("#macro %s(pos, rot)\n%s\n#end\n", frameName(*model, n), m.POVFrameID(n, "blaha"))
	}
}

func main() {
	flag.Parse()
	p, err := pak.MultiOpen(flag.Args()...)
	if err != nil {
		log.Fatalf("Failed to open pakfiles %q: %v", flag.Args(), err)
	}

	switch *command {
	case "convert":
		convert(p)
	case "pov-tri":
		triangles(p)
	case "show":
		show(p)
	}
}

var randColorState int

func randColor() string {
	randColorState++
	colors := []string{
		"Green",
		"White",
		"Blue",
		"Red",
		"Yellow",
	}
	return colors[randColorState%len(colors)]
}
