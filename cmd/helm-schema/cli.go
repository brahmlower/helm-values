package main

import (
	"helmschema/cmd/helm-schema/internal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GenerateCommand() *cobra.Command {
	v := viper.New()

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the worker server",
		RunE: func(cmd *cobra.Command, args []string) error {
			plan := &internal.Plan{
				Logger:         logger,
				ValuesPath:     v.GetString("values"),
				OutputPath:     v.GetString("output"),
				Stdout:         v.GetBool("stdout"),
				StrictComments: v.GetBool("strict"),
			}

			g := internal.NewGenerator(logger, plan)
			_, err := g.Generate()
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("values", "", "path to the values file")
	v.BindPFlag("values", cmd.Flags().Lookup("values"))
	v.BindEnv("values")

	cmd.Flags().String("output", "", "path to the output file")
	v.BindPFlag("output", cmd.Flags().Lookup("output"))
	v.BindEnv("output")

	cmd.Flags().Bool("stdout", false, "write to stdout")
	v.BindPFlag("stdout", cmd.Flags().Lookup("stdout"))
	v.BindEnv("stdout")

	cmd.Flags().Bool("strict", false, "fail on doc comment parsing errors")
	v.BindPFlag("strict", cmd.Flags().Lookup("strict"))
	v.BindEnv("strict")

	return cmd
}
