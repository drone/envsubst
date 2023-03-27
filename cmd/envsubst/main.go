package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/drone/envsubst/v2"
)

var flagStrict bool = false

func main() {
	flag.Parse()
	stdin := bufio.NewScanner(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	for stdin.Scan() {
		line, err := envsubst.EvalEnv(stdin.Text(), flagStrict)
		if err != nil {
			log.Fatalf("Error while envsubst: %v", err)
		}
		_, err = fmt.Fprintln(stdout, line)
		if err != nil {
			log.Fatalf("Error while writing to stdout: %v", err)
		}
		stdout.Flush()
	}
}

func init() {
	flag.BoolVar(&flagStrict, "strict", false, "fail if variable is undefined.")
}