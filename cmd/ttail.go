package main

import (
	"flag"
	"io"
	"os"
	"regexp"
	"time"

	stdLog "log"

	"github.com/sakateka/ttail"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
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

	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zapcore.ErrorLevel)
	if ttail.FlagDebug {
		cfg.Level.SetLevel(zapcore.DebugLevel)
	}
	var err error
	log, err = cfg.Build()
	if err != nil {
		stdLog.Fatalf("can't initialize zap logger: %v", err)
	}

	// TODO: regexp and timeLayout from config. (and flags?)
	re := regexp.MustCompile(`\ttimestamp=(\d{4}-\d{2}-\d{2}T\d\d:\d\d:\d\d)\t`)
	tLayout := "2006-01-02T15:04:05"

	var file *os.File
	var fileInfo os.FileInfo
	for _, fname := range flag.Args() {
		if file != nil {
			file.Close()
			file = nil
		}
		log.Debug("[main]: process file", zap.String("fileName", fname))

		fileInfo, err = os.Stat(fname)
		if err != nil {
			log.Error("[main]: file stat", zap.String("logname", fname), zap.Error(err))
			continue
		} else if fileInfo.IsDir() {
			log.Error("[main]: skip directory!", zap.String("name", fname))
			continue
		}
		file, err = os.Open(fname)
		if err != nil {
			log.Error("[main]: skip", zap.String("logname", fname), zap.Error(err))
			continue
		}
		tfile := ttail.NewTimeFile(re, tLayout, file)

		if err := tfile.FindPosition(timeFromLastLine); err != nil {
			if err != io.EOF {
				log.Fatal("[main]: error", zap.Error(err))
			}
			log.Debug("[main]: findPosition got EOF")
			return
		}
		_, _ = tfile.CopyTo(os.Stdout)
	}
}
