package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uberswe/interval"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var (
	cfgFile string
	src     string
	dest    string
	repeat  string
	rootCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy copies files and directories recursively",
		Long:  "Copy copies files and directories recursively\n\nCreated by Markus Tenghamn (https://github.com/uberswe)",
		Run:   rootFunc,
	}
)

func rootFunc(cmd *cobra.Command, args []string) {
	viperSrc := viper.Get("source")
	viperDest := viper.Get("destination")
	viperRepeat := viper.Get("repeat")
	if (src == "" || dest == "") && viperDest != nil && viperSrc != nil {
		dest = fmt.Sprintf("%v", viperDest)
		src = fmt.Sprintf("%v", viperSrc)
	}
	if repeat == "" && viperRepeat != nil {
		repeat = fmt.Sprintf("%v", viperRepeat)
	}
	if src == "" || dest == "" {
		_ = cmd.Usage()
		return
	}
	if repeat != "" {
		err := interval.DoEvery(repeat, repeatFunc, -1)
		er(err)
	} else {
		err := Dir(src, dest)
		er(err)
	}
	fmt.Printf("%s copied to %s", src, dest)
}

func repeatFunc(interval time.Duration, time time.Time) {
	err := Dir(src, dest)
	er(err)
}

func init() {
	var err error
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./.copy.yaml)")
	rootCmd.PersistentFlags().StringVarP(&src, "source", "s", "", "path to the source file or directory")
	rootCmd.PersistentFlags().StringVarP(&dest, "destination", "d", "", "path to the destination file or directory")
	rootCmd.PersistentFlags().StringVarP(&repeat, "repeat", "e", "", "if specified the path will be copied at the provided interval (example 1m for 1 minute)")
	err = viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	er(err)
	err = viper.BindPFlag("destination", rootCmd.PersistentFlags().Lookup("destination"))
	er(err)
	err = viper.BindPFlag("repeat", rootCmd.PersistentFlags().Lookup("repeat"))
	er(err)
}

func main() {
	execute()
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Dir copies a whole directory recursively
func Dir(source string, destination string) error {
	var err error
	var fileInfos []os.FileInfo
	var sourceInfo os.FileInfo

	if sourceInfo, err = os.Stat(source); err != nil {
		return err
	}

	if err = os.MkdirAll(destination, sourceInfo.Mode()); err != nil {
		return err
	}

	if fileInfos, err = ioutil.ReadDir(source); err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		sourceFilePath := path.Join(source, fileInfo.Name())
		destinationFilePath := path.Join(destination, fileInfo.Name())

		if fileInfo.IsDir() {
			if err = Dir(sourceFilePath, destinationFilePath); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = File(sourceFilePath, destinationFilePath); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func File(src, dst string) error {
	var err error
	var sourceFile *os.File
	var destinationFile *os.File
	var sourceInfo os.FileInfo

	if sourceFile, err = os.Open(src); err != nil {
		return err
	}
	defer sourceFile.Close()

	if destinationFile, err = os.Create(dst); err != nil {
		return err
	}
	defer destinationFile.Close()

	if _, err = io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}
	if sourceInfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func er(msg interface{}) {
	if msg != nil {
		fmt.Println("Error:", msg)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigName(".copy")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
