package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aloder/tojen/gen"
	"github.com/spf13/cobra"
)

func Execute() {
	var packageName string
	var genMain bool
	var formating bool

	var cmdGen = &cobra.Command{
		Use:   "gen [path to file] [output path]",
		Short: "Generate code from file",
		Long: `echo is for echoing anything back.
Echo works a lot like print, except it has a child command.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			b, err := ioutil.ReadFile(args[0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if packageName == "" {
				packageName = "main"
			}
			retBytes, err := gen.GenerateFileBytes(b, packageName, genMain, formating)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if len(args) == 2 {
				osFile, err := os.Create(args[1])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = osFile.Write(retBytes)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println("Successfuly wrote file to " + args[1])
				err = osFile.Close()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				os.Exit(0)
			}
			fmt.Println(string(retBytes))
			os.Exit(0)
		},
	}

	var rootCmd = &cobra.Command{
		Use:   "tojen",
		Short: "Generate jennifer code from file",
		Long:  `Generate jennifer code from a file with the command gen`,
	}
	cmdGen.Flags().StringVarP(&packageName, "package", "p", "", "Name of package")
	cmdGen.Flags().BoolVarP(&genMain, "main", "m", false, "Generate main function")

	cmdGen.Flags().BoolVarP(&formating, "formatted", "f", false, "Format the generated code EXPERIMENTAL")

	rootCmd.AddCommand(cmdGen)
	rootCmd.Execute()

}
