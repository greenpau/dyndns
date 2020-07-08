package main

import (
	"flag"
	"fmt"
	"github.com/greenpau/dyndns"
	"os"

	"github.com/greenpau/versioned"
	"go.uber.org/zap"
)

var (
	log        *zap.Logger
	app        *versioned.PackageManager
	appVersion string
	gitBranch  string
	gitCommit  string
	buildUser  string
	buildDate  string
)

func init() {
	app = versioned.NewPackageManager("dyndns")
	app.Description = "Dynamic DNS Registrator for Route 53"
	app.Documentation = "https://github.com/greenpau/dyndns/"
	app.SetVersion(appVersion, "")
	app.SetGitBranch(gitBranch, "")
	app.SetGitCommit(gitCommit, "")
	app.SetBuildUser(buildUser, "")
	app.SetBuildDate(buildDate, "")
}

func main() {
	var configFile string
	var logLevel string
	var isShowVersion bool
	var isValidate bool
	flag.StringVar(&configFile, "config", "", "path to configuration file")
	flag.StringVar(&logLevel, "log-level", "info", "logging severity level")
	flag.BoolVar(&isValidate, "validate", false, "validate configuration")
	flag.BoolVar(&isShowVersion, "version", false, "version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s - %s\n\n", app.Name, app.Description)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", app.Name)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: %s\n\n", app.Documentation)
	}
	flag.Parse()

	server := dyndns.NewServer()
	if err := server.SetLogLevel(logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed setting %s log level: %s\n", logLevel, err)
		os.Exit(1)
	}

	log = server.GetLogger()
	defer log.Sync()

	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s\n", app.Banner())
		os.Exit(0)
	}

	if configFile != "" {
		if err := server.LoadConfig(configFile); err != nil {
			log.Fatal("error reading configuration file", zap.String("error", err.Error()))
		}
		log.Debug("running configuration", zap.Any("config", server.GetConfig()))
	}

	if err := server.ValidateConfig(); err != nil {
		log.Fatal("invalid configuration", zap.String("error", err.Error()))
	}

	if isValidate {
		fmt.Fprintf(os.Stdout, "configuration is valid\n")
		os.Exit(0)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
