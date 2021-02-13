package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uberswe/interval"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"time"
)

var (
	cfgFile   string
	src       string
	dest      string
	repeat    string
	destSlice []string
	srcSlice  []string
	rootCmd   = &cobra.Command{
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
		ds, destOk := viperDest.([]interface{})
		ss, srcOk := viperSrc.([]interface{})
		if !destOk {
			log.Println("dest conversion failed")
			log.Println(reflect.TypeOf(viperDest))
			dest = fmt.Sprintf("%v", viperDest)
		} else {
			destSlice = interfaceSliceToStringSlice(ds)
		}
		if !srcOk {
			log.Println("src conversion failed")
			log.Println(reflect.TypeOf(viperSrc))
			src = fmt.Sprintf("%v", viperSrc)
		} else {
			srcSlice = interfaceSliceToStringSlice(ss)
		}
	}
	if repeat == "" && viperRepeat != nil {
		repeat = fmt.Sprintf("%v", viperRepeat)
	}
	if (src == "" || dest == "") && (len(srcSlice) == 0 || len(destSlice) == 0) {
		_ = cmd.Usage()
		return
	}
	if repeat != "" {
		if len(destSlice) > 0 && len(srcSlice) > 0 && len(srcSlice) == len(destSlice) {
			err := interval.DoEvery(repeat, nil, repeatFuncMulti, -1)
			er(err)
		} else {
			err := interval.DoEvery(repeat, nil, repeatFunc, -1)
			er(err)
			fmt.Printf("%s copied to %s", src, dest)
		}
	} else {
		if len(destSlice) > 0 && len(srcSlice) > 0 && len(srcSlice) == len(destSlice) {
			err := copyMultipleDirectories(srcSlice, destSlice)
			er(err)
		} else {
			err := copyDir(src, dest)
			er(err)
			fmt.Printf("%s copied to %s", src, dest)
		}
	}
}

func interfaceSliceToStringSlice(i []interface{}) []string {
	var r []string
	for _, s := range i {
		log.Printf("%v\n", s)
		if str, ok := s.(string); ok {
			r = append(r, str)
		} else {
			panic(fmt.Sprintf("expected string in slice but got %s", reflect.TypeOf(s)))
		}
	}
	return r
}

func copyMultipleDirectories(s []string, d []string) error {
	for i := 0; i < len(s); i++ {
		log.Printf("%s copied to %s\n", s[i], d[i])
		err := copyDir(s[i], d[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func repeatFuncMulti(interval time.Duration, time time.Time, extra interface{}) {
	for i := 0; i < len(src); i++ {
		log.Printf("%s copied to %s\n", srcSlice[i], destSlice[i])
		err := copyDir(srcSlice[i], destSlice[i])
		er(err)
	}
}

func repeatFunc(interval time.Duration, time time.Time, extra interface{}) {
	err := copyDir(src, dest)
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

// copyDir copies a whole directory recursively
func copyDir(source string, destination string) error {
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
			if err = copyDir(sourceFilePath, destinationFilePath); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(sourceFilePath, destinationFilePath); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
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
