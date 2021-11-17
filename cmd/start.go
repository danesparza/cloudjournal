package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danesparza/cloudjournal/cloudwatch"
	"github.com/danesparza/cloudjournal/data"
	"github.com/danesparza/cloudjournal/journal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tidwall/buntdb"
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
		"systemdb":           systemdb,
		"loglevel":           loglevel,
		"cloudwatch.group":   viper.GetString("cloudwatch.group"),
		"cloudwatch.profile": viper.GetString("cloudwatch.profile"),
		"cloudwatch.region":  viper.GetString("cloudwatch.region"),
	}).Info("Starting up")

	//	Create a DBManager object
	db, err := data.NewManager(systemdb)
	if err != nil {
		log.WithError(err).Error("Error trying to open the system database")
		return
	}
	defer db.Close()

	//	Associate the dbmanager object with the cloudwatch svc
	cloudService := cloudwatch.Service{
		DB:           db,
		LogGroupName: viper.GetString("cloudwatch.group"),
	}

	//	Trap program exit appropriately
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go handleSignals(ctx, sigs, cancel)

	//	Get the comma-seperated list of units to check from configuration
	monitoredUnits := strings.Split(viper.GetString("monitor.units"), ",")

	//	If there are no units specified, indicate that in the log
	if len(monitoredUnits) < 1 {
		log.WithFields(log.Fields{
			"monitor.units": viper.GetString("monitor.units"),
		}).Fatal("No monitor.units specified in config file.  There is nothing to monitor")
	}

	//	Log that the system has started:
	log.Info("System started")

	t := time.Tick(1 * time.Minute)
	for {
		select {
		case <-t:
			//	For each unit
			for _, unit := range monitoredUnits {
				unit = strings.TrimSpace(unit)

				//	Get the state for the unit
				unitState, err := cloudService.DB.GetLogStateForUnit(unit)
				if err != nil && err != buntdb.ErrNotFound {
					log.WithFields(log.Fields{
						"unit": unit,
					}).WithError(err).Error("problem trying to get state for unit")
				}

				//	Get the entries from the last cursor
				entries := journal.GetJournalEntriesForUnitFromCursor(unit, unitState.LastCursor)

				//	If we have entries ...
				if len(entries) > 0 {
					log.WithFields(log.Fields{
						"unit":       unit,
						"founditems": len(entries),
					}).Debug("found items to log")

					//	Log the entries:
					err = cloudService.WriteToLog(unit, entries)
					if err != nil {
						//	If we have an error, don't save state.  Just continue
						log.WithFields(log.Fields{
							"unit": unit,
						}).WithError(err).Error("problem writing to log.  Retrying with next batch")
						continue
					}

					//	Get the last cursor:
					lastCursor := entries[len(entries)-1].Cursor

					//	Save the state for the unit
					_, err = cloudService.DB.UpdateLogState(unit, lastCursor)
					if err != nil {
						log.WithFields(log.Fields{
							"unit": unit,
						}).WithError(err).Error("problem trying to save state for unit")
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func handleSignals(ctx context.Context, sigs <-chan os.Signal, cancel context.CancelFunc) {
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
