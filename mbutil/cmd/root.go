package cmd

import "github.com/spf13/cobra"
import "os"
import "fmt"

import "github.com/spf13/viper"

var filename string
var keyvalue []string 
func init() {
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "Your filename")
	viper.BindPFlag("filename", rootCmd.PersistentFlags().Lookup("filename"))
	rootCmd.PersistentFlags().StringArrayVar(&keyvalue,"key-value-update" ,[]string{}, "Your filename")
	viper.BindPFlag("key-value-update", rootCmd.PersistentFlags().Lookup("key-value-update"))
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
