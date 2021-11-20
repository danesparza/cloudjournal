package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danesparza/cloudjournal/cloudwatch"
	"github.com/danesparza/cloudjournal/data"
	"github.com/danesparza/cloudjournal/journal"
	"github.com/danesparza/cloudjournal/system"
	"github.com/danesparza/cloudjournal/token"
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

	//	Start our token map:
	tokens := make(map[string]string)
	tokens["{machineid}"] = system.GetMachineID()
	tokens["{hostname}"] = system.GetHostname()

	//	Emit what we know:
	log.WithFields(log.Fields{
		"systemdb":           systemdb,
		"loglevel":           loglevel,
		"machineid":          tokens["{machineid}"],
		"hostname":           tokens["{hostname}"],
		"cloudwatch.group":   viper.GetString("cloudwatch.group"),
		"cloudwatch.stream":  viper.GetString("cloudwatch.stream"),
		"cloudwatch.profile": viper.GetString("cloudwatch.profile"),
		"cloudwatch.region":  viper.GetString("cloudwatch.region"),
		"monitor.units":      viper.GetString("monitor.units"),
		"monitor.interval":   viper.GetString("monitor.interval"),
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
		DB: db,
	}

	//	Convert interval to a duration
	monitorInterval, err := time.ParseDuration(fmt.Sprintf("%vm", viper.GetString("monitor.interval")))
	if err != nil {
		log.WithFields(log.Fields{
			"monitor.interval": viper.GetString("monitor.interval"),
		}).WithError(err).Error("problem converting interval to a duration")
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

	t := time.Tick(monitorInterval)
	for {
		select {
		case <-t:
			//	For each unit
			for _, unit := range monitoredUnits {
				unit = strings.TrimSpace(unit)

				//	Format our group and stream names
				tokens["{unit}"] = unit
				cloudwatchGroupname := token.Replace(viper.GetString("cloudwatch.group"), tokens)
				cloudwatchStreamname := token.Replace(viper.GetString("cloudwatch.stream"), tokens)

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

					//	Log the entries:
					err = cloudService.WriteToLog(cloudwatchGroupname, cloudwatchStreamname, entries)
					if err != nil {
						//	If we have an error, don't save state.  Just continue
						log.WithFields(log.Fields{
							"unit":       unit,
							"groupName":  cloudwatchGroupname,
							"streamName": cloudwatchStreamname,
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
