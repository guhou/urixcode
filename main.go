package main

import (
	"bufio"
	"fmt"
	flag "github.com/spf13/pflag"
	"net/url"
	"os"
)

type (
	exitCode   int
	transcoder func(string) string
	runner     func(transcoder)
)

// exit code constants
const (
	exitSuccess = 1<<iota - 1
	exitUsage
	exitIOError
	exitDecodeError
)

// set up program flags
var (
	flagDecode = flag.BoolP("decode", "D", false, "run in decode mode")
	flagEncode = flag.BoolP("encode", "E", false, "run in encode mode")
	flagHelp   = flag.BoolP("help", "h", false, "print usage and exit")
	flagStdin  = flag.Bool("stdin", false, "read URLs from stdin instead of args")
)

///////////////////////////////////////////////////////////////////////
// main
///////////////////////////////////////////////////////////////////////

func main() {
	var (
		transcoder transcoder
		runner     runner
	)

	// set up flag parsing
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = usage
	flag.Parse()

	// check for the help flag
	if *flagHelp {
		flag.Usage()
		exit(exitUsage)
	}

	// set mode switches- decode or encode mode
	switch {
	case *flagDecode && *flagEncode:
		stderrf("decode and encode are both set. I don't know what to do.\n")
		exit(exitUsage)
		break
	case *flagDecode:
		transcoder = decoder
		break
	case *flagDecode:
	default:
		transcoder = encoder
		break
	}

	// set input source- stdin or program args
	if *flagStdin {
		runner = runStdin
	} else {
		runner = runArgs
	}

	// invoke with the configured mode and input
	runner(transcoder)
}

///////////////////////////////////////////////////////////////////////
// runner functions
///////////////////////////////////////////////////////////////////////

func runArgs(transcoder transcoder) {
	for i := 0; i < flag.NArg(); i++ {
		input := flag.Arg(i)
		output := transcoder(input)
		stdoutf("%s\n", output)
	}
}

func runStdin(transcoder transcoder) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		output := transcoder(input)
		stdoutf(output + "\n")
	}
	if err := scanner.Err(); err != nil {
		stderrf("stdin read failed: %v\n", err)
		exit(exitIOError)
	}
}

///////////////////////////////////////////////////////////////////////
// transcoder functions
///////////////////////////////////////////////////////////////////////

func encoder(s string) string {
	encoded := url.QueryEscape(s)
	return encoded
}

func decoder(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		stderrf("bad encoding: %q\n", decoded)
		exit(exitDecodeError)
	}
	return decoded
}

///////////////////////////////////////////////////////////////////////
// utility output functions
///////////////////////////////////////////////////////////////////////

func usage() {
	programName := os.Args[0]
	stderrf("%s is an encoder and decoder for percent-encoding (a.k.a. URI-encoding)\n\n", programName)
	flag.PrintDefaults()
}

func stdoutf(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(os.Stdout, format, args...); err != nil {
		stderrf("stdout write failed: %v\n", err)
		exit(exitIOError)
	}
}

func stderrf(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format, args...); err != nil {
		// Since stderr just failed its write, assume that it is impossible to report
		// the error in any way other than crashing.
		exit(exitIOError)
	}
}

func exit(code exitCode) {
	os.Exit(int(code))
}
