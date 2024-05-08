package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/influxdb-alerter/internal/checker"
	"github.com/niktheblak/influxdb-alerter/pkg/notifier"
	"github.com/niktheblak/influxdb-alerter/pkg/notifier/console"
	"github.com/niktheblak/influxdb-alerter/pkg/notifier/email"
	"github.com/niktheblak/influxdb-alerter/pkg/notifier/http"
)

var notifiers []notifier.Notifier

var checkCmd = &cobra.Command{
	Use:          "check",
	Short:        "Check for recent measurements in InfluxDB",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("console.enabled") {
			notifiers = append(notifiers, &console.Console{Writer: os.Stdout})
		}
		if viper.GetBool("email.enabled") {
			e := email.New(email.Config{
				Host:     viper.GetString("email.host"),
				Port:     viper.GetInt("email.port"),
				Username: viper.GetString("email.username"),
				Password: viper.GetString("email.password"),
				From:     viper.GetString("email.from"),
				To:       []string{viper.GetString("email.to")},
			})
			notifiers = append(notifiers, e)
		}
		if viper.GetBool("http.enabled") {
			h := http.New(http.Config{
				URL:           viper.GetString("http.url"),
				Method:        "POST",
				Authorization: viper.GetString("http.authorization"),
				ContentType:   "text/plain",
			})
			notifiers = append(notifiers, h)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := checker.Config{
			Addr:        viper.GetString("influxdb.addr"),
			Org:         viper.GetString("influxdb.org"),
			Token:       viper.GetString("influxdb.token"),
			Bucket:      viper.GetString("influxdb.bucket"),
			Measurement: viper.GetString("influxdb.measurement"),
			Since:       viper.GetDuration("max_age"),
		}
		logger.Info("Using config", "config", cfg)
		c, err := checker.New(cfg, logger)
		if err != nil {
			return err
		}
		defer c.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err = c.Check(ctx)
		if errors.Is(err, &checker.ErrNoRecentData{}) {
			msg := err.Error()
			logger.Info(fmt.Sprintf("%s", msg))
			if err := notify(ctx, msg); err != nil {
				logger.Error("Failed to send notification", "error", err)
			}
			err = nil
		}
		if errors.Is(err, checker.ErrNoResults) {
			logger.Info("Query returned no results")
			// Database has likely just started or the wrong bucket is specified; ignore
			err = nil
		}
		return err
	},
}

func notify(ctx context.Context, message string) error {
	for _, n := range notifiers {
		if err := n.Notify(ctx, message); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	checkCmd.Flags().Duration("max_age", 0, "threshold measurement age for triggering notification")

	checkCmd.Flags().String("influxdb.addr", "", "InfluxDB server address")
	checkCmd.Flags().String("influxdb.org", "", "InfluxDB organization")
	checkCmd.Flags().String("influxdb.token", "", "InfluxDB token")
	checkCmd.Flags().String("influxdb.bucket", "", "InfluxDB bucket")
	checkCmd.Flags().String("influxdb.measurement", "", "InfluxDB measurement")

	checkCmd.Flags().Bool("email.enabled", false, "email notifications enabled")
	checkCmd.Flags().String("email.host", "", "SMTP host")
	checkCmd.Flags().Int("email.port", 0, "SMTP port")
	checkCmd.Flags().String("email.username", "", "SMTP username")
	checkCmd.Flags().String("email.password", "", "SMTP password")
	checkCmd.Flags().String("email.from", "", "email from field")
	checkCmd.Flags().String("email.to", "", "email to field")

	checkCmd.Flags().Bool("http.enabled", false, "HTTP notifications enabled")
	checkCmd.Flags().String("http.url", "", "HTTP notification URL")
	checkCmd.Flags().String("http.authorization", "", "HTTP Authorization header")

	checkCmd.Flags().Bool("console.enabled", false, "console notifications enabled")

	cobra.CheckErr(viper.BindPFlags(checkCmd.Flags()))

	viper.SetDefault("max_age", 30*time.Minute)
	viper.SetDefault("influxdb.addr", "http://127.0.0.1:8086")
	viper.SetDefault("influxdb.bucket", "ruuvitag")
	viper.SetDefault("influxdb.measurement", "ruuvitag")
	viper.SetDefault("email.port", 587)

	rootCmd.AddCommand(checkCmd)
}
