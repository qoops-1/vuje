package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bitly/go-simplejson"
	"github.com/nwidger/jsoncolor"

	termbox "github.com/nsf/termbox-go"
)

const version = "0.0.1"

func main() {
	var query string
	var separator string
	var pretty bool
	var ver bool
	finfo, err := os.Stdout.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	pipedOutput := (finfo.Mode() & os.ModeCharDevice) == 0
	flag.StringVar(&query, "s", "", "execute specified query in non-interactive mode and return results")
	flag.StringVar(&separator, "d", ".", "specify custom separator for the query. Default is \".\"")
	flag.BoolVar(&pretty, "p", !pipedOutput, "set to true if final output should be coloured. "+
		"By default the flag is set to true, but, if the output of the program is piped, it is set to false")
	flag.BoolVar(&ver, "v", false, "output version")
	flag.BoolVar(&ver, "version", false, "output version")
	flag.Parse()
	if ver {
		fmt.Println(version)
		return
	}
	if len([]rune(separator)) != 1 {
		log.Panicf("Separator must be a single character")
	}
	stdin := bufio.NewReader(os.Stdin)
	explorer := NewExplorer(stdin, []rune(separator)[0])
	var res *simplejson.Json
	if query != "" {
		res = explorer.ExecuteQuery(query)
	} else {
		res = explorer.Run()
	}
	printResult(res, pretty)
	fmt.Println()
}

func printResult(res *simplejson.Json, pretty bool) {
	if !pretty {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		enc.Encode(res.Interface())
		return
	}
	binJSON, _ := res.Encode()
	fmtr := jsoncolor.NewFormatter()
	fmtr.Indent = "  "
	err := fmtr.Format(os.Stdout, binJSON)
	if err != nil {
		log.Fatalln(err)
	}
}

func termboxFatalf(msg string, args ...interface{}) {
	termbox.Close()
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func termboxFatalln(vs ...interface{}) {
	termbox.Close()
	fmt.Fprintln(os.Stderr, vs...)
	os.Exit(1)
}
