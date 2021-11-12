package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danesparza/cloudjournal/data"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `The start command starts the log shipping server`,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {

	//	If we have a config file, report it:
	if viper.ConfigFileUsed() != "" {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Debug("No config file found.")
	}

	systemdb := viper.GetString("datastore.system")
	loglevel := viper.GetString("log.level")

	//	Emit what we know:
	log.WithFields(log.Fields{
		"systemdb": systemdb,
		"loglevel": loglevel,
	}).Info("Starting up")

	//	Create a DBManager object and associate with the api.Service
	db, err := data.NewManager(systemdb)
	if err != nil {
		log.WithError(err).Error("Error trying to open the system database")
		return
	}
	defer db.Close()

	//	Trap program exit appropriately
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go handleSignals(ctx, sigs, cancel, db)

	//	Get the comma-seperated list of units to check from configuration

	//	If there are no units specified, indicate that in the log

	//	Log that the system has started:
	log.Info("System started")

	/*
		entries := journal.GetJournalEntriesForUnitFromCursor("daydash", "s=152362fbd3cb491dac4b70a0eb7da4d7;i=230;b=378e82b47ba0454fad0b338e20aec7b0;m=235a86f;t=5d08066dc3ef4;x=c25b0ce5a0e74080")

		for _, entry := range entries {
			log.WithFields(log.Fields{
				"message": entry.Message,
			}).Info("Got an item")
		}
	*/

	t := time.Tick(1 * time.Minute)
	for {
		select {
		case <-t:
			log.Info("Something is happening...")
		case <-ctx.Done():
			return
		}
	}
}

func handleSignals(ctx context.Context, sigs <-chan os.Signal, cancel context.CancelFunc, db *data.Manager) {
	select {
	case <-ctx.Done():
	case sig := <-sigs:
		switch sig {
		case os.Interrupt:
			log.WithFields(log.Fields{
				"signal": "SIGINT",
			}).Info("Shutting down")
		case syscall.SIGTERM:
			log.WithFields(log.Fields{
				"signal": "SIGTERM",
			}).Info("Shutting down")
		}

		cancel()
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
}
