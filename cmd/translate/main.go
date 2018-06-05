package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wzshiming/requests/util/translate"
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
	if *to == "" {
		flag.Usage()
		return
	}
	var reader *bufio.Reader
	if len(a) == 0 {
		buf, _ := ioutil.ReadAll(os.Stdin)
		reader = bufio.NewReader(bytes.NewBuffer(buf))
	} else {
		text := strings.Join(a, " ")
		reader = bufio.NewReader(bytes.NewBufferString(text))
	}
	if reader == nil {
		flag.Usage()
		return
	}

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println(err)
			return
		}
		ret, err := translate.GoogleTranslate(string(line), *from, *to)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ret)
	}
}
