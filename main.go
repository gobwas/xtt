package main

import (
	"bufio"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gobwas/flagutil"
	"github.com/gobwas/flagutil/parse/pargs"
)

var (
	input   string
	output  string
	verbose bool

	attrMod = make(map[xml.Name][]Mod)
)

type Mod interface {
	Apply(string) (string, error)
	Name() string
}

func main() {
	log.SetFlags(0)

	flag.StringVar(&input,
		"input", "",
		"input file",
	)
	flag.StringVar(&output,
		"output", "",
		"output file",
	)
	flag.BoolVar(&verbose,
		"verbose", false,
		"be verobse",
	)
	flag.Var((mods)(attrMod),
		"mod",
		"attribute modifiers in form of `attribute-name:mod1,mod2,mod3`",
	)
	_ = flagutil.Parse(context.Background(), flag.CommandLine,
		flagutil.WithCustomUsage(),
		flagutil.WithParser(&pargs.Parser{
			Shorthand: true,
			Args:      os.Args[1:],
		}),
	)
	if input == "" || output == "" {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flagutil.PrintDefaults(context.Background(), flag.CommandLine,
			flagutil.WithCustomUsage(),
			flagutil.WithParser(&pargs.Parser{
				Shorthand: true,
			}),
		)
		os.Exit(1)
	}
	src, err := os.Open(input)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	dst, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	var (
		br = bufio.NewReader(src)
		bw = bufio.NewWriter(dst)
	)

	dec := xml.NewDecoder(br)
	enc := xml.NewEncoder(bw)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		switch v := tok.(type) {
		case xml.StartElement:
			cp := v.Copy()
			for i, attr := range v.Attr {
				curr := attr.Value
				for _, mod := range attrMod[attr.Name] {
					next, err := mod.Apply(curr)
					if err != nil {
						log.Fatal(err)
					}
					if verbose {
						log.Printf(
							"applied %s to %s: %q -> %q",
							mod.Name(), name(attr.Name),
							curr, next,
						)
					}
					curr = next
				}
				x := attr
				x.Value = curr
				cp.Attr[i] = x
			}
			tok = cp
		}
		if err != nil {
			log.Fatal(err)
		}
		if err := enc.EncodeToken(tok); err != nil {
			log.Fatal(err)
		}
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}
}

func name(n xml.Name) string {
	if n.Space == "" {
		return n.Local
	}
	return n.Local + "@" + n.Space
}

type mods map[xml.Name][]Mod

func (m mods) Set(s string) error {
	attr, mods := split2(s, ':')
	if mods == "" {
		return fmt.Errorf("malformed mod pair: %q", s)
	}
	attr, space := split2(attr, '@')
	name := xml.Name{
		Local: attr,
		Space: space,
	}
	for _, mod := range strings.Split(mods, ",") {
		mod = strings.TrimSpace(mod)
		switch mod {
		case "fraction":
			ms := append(m[name], new(FractionMod))
			m[name] = ms

		default:
			return fmt.Errorf("unknown mod: %q", mod)
		}
	}
	return nil
}

func (m mods) String() string {
	var sb strings.Builder
	for attr, mods := range m {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		fmt.Fprintf(&sb, "%s:", attr)
		for i, mod := range mods {
			if i > 0 {
				sb.WriteByte(',')
			}
			fmt.Fprintf(&sb, "%s", mod.Name())
		}
	}
	return sb.String()
}

func split2(s string, c byte) (a, b string) {
	i := strings.IndexByte(s, c)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}
