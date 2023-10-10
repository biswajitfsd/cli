package fingerprint

import (
	"fmt"
	"path/filepath"

	"github.com/debricked/cli/internal/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exclusions = file.DefaultExclusionsFingerprint()

const (
	ExclusionFlag = "exclusion-fingerprint"
)

func NewFingerprintCmd(fingerprinter file.IFingerprint) *cobra.Command {

	short := fmt.Sprintf("Fingerprint files for identification in a given path and writes it to %s. [beta feature]", file.OutputFileNameFingerprints)
	long := fmt.Sprintf("Fingerprint files for identification in a given path and writes it to %s. [beta feature]\nThis hashes all files and matches them against the Debricked knowledge base.", file.OutputFileNameFingerprints)
	cmd := &cobra.Command{
		Use:   "fingerprint [path]",
		Short: short,
		Long:  long,
		PreRun: func(cmd *cobra.Command, _ []string) {
			_ = viper.BindPFlags(cmd.Flags())
		},
		RunE: RunE(fingerprinter),
	}
	fileExclusionExample := filepath.Join("*", "**.pyc")
	dirExclusionExample := filepath.Join("**", "node_modules", "**")
	exampleFlags := fmt.Sprintf("-%s \"%s\" -%s \"%s\"", ExclusionFlag, fileExclusionExample, ExclusionFlag, dirExclusionExample)
	cmd.Flags().StringArrayVarP(&exclusions, ExclusionFlag, "", exclusions, `The following terms are supported to exclude paths:
Special Terms | Meaning
------------- | -------
"*"           | matches any sequence of non-Separator characters 
"/**/"        | matches zero or multiple directories
"?"           | matches any single non-Separator character
"[class]"     | matches any single non-Separator character against a class of characters ([see "character classes"])
"{alt1,...}"  | matches a sequence of characters if one of the comma-separated alternatives matches

Example: 
$ debricked files fingerprint . `+exampleFlags)

	viper.MustBindEnv(ExclusionFlag)

	return cmd
}

func RunE(f file.IFingerprint) func(_ *cobra.Command, args []string) error {
	return func(_ *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		output, err := f.FingerprintFiles(path, exclusions)

		if err != nil {
			return err
		}

		err = output.ToFile(file.OutputFileNameFingerprints)
		if err != nil {
			return err
		}

		return nil
	}
}
