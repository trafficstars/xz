// Commmand buzgen generates the hash table for a buzhash
// implementation. It uses the first argument as filename that should
// be generated.
package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"text/template"
)

type Table [256]uint32

func randomUint32() uint32 {
	p := make([]byte, 4)
	n, err := rand.Read(p)
	if err != nil {
		panic(fmt.Errorf("rand.Read error %s", err))
	}
	if n != 4 {
		panic(fmt.Errorf("rand.Read returned only %d bytes", n))
	}
	var u uint32
	u |= uint32(p[0])
	u |= uint32(p[1]) << 8
	u |= uint32(p[2]) << 16
	u |= uint32(p[3]) << 24
	return u
}

func generateTable() *Table {
	var tab Table
	for i := range tab {
		tab[i] = randomUint32()
	}
	return &tab
}

type Data struct {
	Package string
	Table   Table
}

func main() {
	log.SetPrefix("buzgen: ")
	log.SetFlags(0)

	out := os.Stdout
	if len(os.Args) > 1 {
		var err error
		if out, err = os.Create(os.Args[1]); err != nil {
			log.Fatal(err)
		}
		defer out.Close()
	}

	tab := generateTable()

	funcMap := template.FuncMap{"nlreq": nlreq, "tabreq": tabreq}
	tmpl, err := template.New("array").Funcs(funcMap).Parse(arrayTempl)
	if err != nil {
		log.Fatalf("template parse error %s", err)
	}

	pkg := os.Getenv("GOPACKAGE")
	if pkg == "" {
		pkg = "main"
	}

	err = tmpl.Execute(out, Data{Package: pkg, Table: *tab})
	if err != nil {
		log.Fatalf("template execution error %s", err)
	}
}

func nlreq(n int) bool {
	return (n+1)%4 == 0
}

func tabreq(n int) bool {
	return n%4 == 0
}

const arrayTempl = `package {{.Package}}

var table = [256]uint32{
{{range $i, $u := .Table -}}
	{{if tabreq $i}}{{"\t"}}{{end -}}
	{{$u | printf "%#08x" -}}
	{{if nlreq $i}}{{",\n"}}{{else}}, {{end -}}
{{end -}}
}
`
