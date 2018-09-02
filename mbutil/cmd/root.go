package cmd

import "github.com/spf13/cobra"
import "os"
import "fmt"

import "github.com/spf13/viper"

var filename string

func init() {
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "your filename", "Author name for copyright attribution")
	viper.BindPFlag("filename", rootCmd.PersistentFlags().Lookup("filename"))

}

var rootCmd = &cobra.Command{
	Use:   "mbutil",
	Short: "mbutil is tool for viewing / manipulating mapbox mbtiles",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
