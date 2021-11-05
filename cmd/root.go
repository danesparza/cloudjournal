package cmd

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nimbostratus",
	Short: "A journald to cloudwatch log shipper",
	Long:  `A journald to cloudwatch log shipper`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/nimbostratus.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.WithError(err).Fatal("Couldn't find home directory")
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(home)           // adding home directory as first search path
		viper.AddConfigPath(".")            // also look in the working directory
		viper.SetConfigName("nimbostratus") // name the config file (without extension)
	}

	viper.AutomaticEnv() // read in environment variables that match

	//	Set our defaults
	viper.SetDefault("datastore.system", path.Join(home, "nimbostratus", "db", "system.db"))
	viper.SetDefault("datastore.retentiondays", 30)
	viper.SetDefault("server.port", "2005")
	viper.SetDefault("server.allowed-origins", "*")
	viper.SetDefault("log.level", "info")

	// If a config file is found, read it in
	viper.ReadInConfig()

	//	Set the log level based on configuration:
	loglevel := viper.GetString("log.level")
	switch loglevel {
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

}