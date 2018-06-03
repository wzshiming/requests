package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wzshiming/requests/examples/translate"
)

var (
	from = flag.String("from", "auto", "from lang")
	to   = flag.String("to", "", "to lang")
)

func init() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Translate:\n")
		for i := 0; ; i++ {
			v, ok := translate.GoogleCodeMap[translate.GoogleCode(i)]
			if !ok {
				break
			}
			fmt.Fprintf(w, "   %s\n", v[1])
		}
		fmt.Fprintf(w, "Examples:\n")
		fmt.Fprintf(w, "    %s [Options] {text}\n", os.Args[0])
		fmt.Fprintf(w, "    %s -to zh-CN hello\n", os.Args[0])
		fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	a := flag.Args()
	if len(a) == 0 || *to == "" {
		flag.Usage()
		return
	}
	text := strings.Join(a, " ")
	ret, err := translate.GoogleTranslate(text, *from, *to)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
