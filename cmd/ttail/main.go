package main

import (
	"flag"
	"io"
	"os"
	"time"

	stdLog "log"

	"github.com/sakateka/ttail"
	"github.com/sakateka/ttail/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var flagTimeFromLastLine bool
var flagLogType string
var flagDuration time.Duration
var configPath string

func init() {
	flag.Usage = func() {
		_, _ = os.Stderr.WriteString("Usage of " + os.Args[0] + " [options] file [file ...]:\n")
		flag.PrintDefaults()
	}
	flag.DurationVar(&flagDuration, "n", 10*time.Second, "offset in time to start copy")
	flag.BoolVar(&flagTimeFromLastLine, "l", false, "tail last N seconds from time in the last line (default from now)")
	flag.StringVar(&flagLogType, "t", "tskv", "use a type of log")
	flag.BoolVar(&ttail.FlagDebug, "d", false, "set Debug mode")
	flag.StringVar(&configPath, "c", config.DefaultConfigFile, "set config path")
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
	defer log.Sync()

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

		// Build options for this file
		opts := []ttail.TimeFileOptions{
			ttail.WithTimeFromLastLine(flagTimeFromLastLine),
			ttail.WithDuration(flagDuration),
		}

		if flagLogType != "" {
			logOpts, err := ttail.OptionsFromConfig(flagLogType, configPath)
			if err != nil {
				log.Fatal("Failed to get ttail options from config", zap.Error(err))
			}
			opts = append(opts, logOpts...)
		}

		tfile := ttail.NewTimeFile(file, opts...)

		if err := tfile.FindPosition(); err != nil {
			if err != io.EOF {
				log.Fatal("[main]: error", zap.Error(err))
			}
			log.Debug("[main]: findPosition got EOF")
			continue
		}

		_, _ = tfile.CopyTo(os.Stdout)
	}

	if file != nil {
		file.Close()
	}
}
