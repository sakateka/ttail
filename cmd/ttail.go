package main

import (
	"flag"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/combaine/ttail"
	"github.com/labstack/gommon/log"
)

var timeFromLastLine bool

func init() {
	flag.Usage = func() {
		_, _ = os.Stderr.WriteString("Usage of " + os.Args[0] + " [options] file [file ...]:\n")
		flag.PrintDefaults()
	}
	flag.DurationVar(&ttail.FlagDuration, "num", 10*time.Second, "offset in time to start copy")
	flag.DurationVar(&ttail.FlagDuration, "n", 10*time.Second, "offset in time to start copy (shorthand)")
	flag.BoolVar(&timeFromLastLine, "l", false, "tail last N secconds from time in last line (default from time.Now())")
	flag.BoolVar(&timeFromLastLine, "last", false, "tail last N secconds from time in last line (default from time.Now())")
	flag.BoolVar(&ttail.FlagDebug, "d", false, "Debug mode")
	flag.BoolVar(&ttail.FlagDebug, "debug", false, "Debug mode")
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	if ttail.FlagDebug {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.DEBUG)
	}

	// TODO: regexp and timeLayout from config. (and flags?)
	re := regexp.MustCompile(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`)
	tLayout := "2006-01-02T15:04:05"

	var file *os.File
	var err error
	for _, fname := range flag.Args() {
		if file != nil {
			file.Close()
			file = nil
		}
		log.Debugf("[main]: process file %s", fname)

		_, err = os.Stat(fname)
		if os.IsNotExist(err) {
			log.Debugf("[main]: skip %s: %s", fname, err)
			continue
		}
		file, err = os.Open(fname)
		if err != nil {
			log.Debugf("[main]: skip %s: %s", fname, err)
			continue
		}
		fileInfo, err := file.Stat()
		if err != nil {
			log.Debugf("[main]: file stat %s: %s", fname, err)
			continue
		} else if fileInfo.IsDir() {
			log.Debugf("[main]: skip directory %s!", fname)
			continue
		}
		tfile := ttail.NewTimeFile(re, tLayout, file)

		if err := tfile.FindPosition(timeFromLastLine); err != nil {
			if err != io.EOF {
				log.Fatalf("[main]: error: %s", err)
			}
			log.Debugf("[main]: findPosition got EOF")
			return
		}
		_, _ = tfile.CopyTo(os.Stdout)
	}
}
