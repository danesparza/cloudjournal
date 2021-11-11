package cmd

import (
	"github.com/danesparza/cloudjournal/journal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		entries := journal.GetJournalEntriesForUnitFromCursor("daydash", "s=152362fbd3cb491dac4b70a0eb7da4d7;i=230;b=378e82b47ba0454fad0b338e20aec7b0;m=235a86f;t=5d08066dc3ef4;x=c25b0ce5a0e74080")

		for _, entry := range entries {
			log.WithFields(log.Fields{
				"message": entry.Message,
			}).Info("Got an item")
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
