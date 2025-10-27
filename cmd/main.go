package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/noders-team/go-daml/internal/codegen"
	"github.com/noders-team/go-daml/internal/codegen/model"
	"github.com/spf13/cobra"
)

var (
	dar    string
	output string
	debug  bool
	pkg    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "godaml --dar <path> --output <dir> --go_package <name> [--debug]",
		Short: "Go DAML codegen tool",
		Long: `A command-line interface tool for generating Go code from DAML (.dar) files.

This tool extracts DAML definitions from .dar archives and generates corresponding Go structs and types.`,
		Example: `  godaml --dar ./test.dar --output ./generated --go_package main
  godaml --dar /path/to/contracts.dar --output ./src/daml --go_package contracts --debug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dar == "" {
				return fmt.Errorf("--dar parameter is required")
			}
			if output == "" {
				return fmt.Errorf("--output parameter is required")
			}
			if pkg == "" {
				return fmt.Errorf("--go_package parameter is required")
			}

			return runCodeGen(dar, output, pkg, debug)
		},
	}

	rootCmd.Flags().StringVar(&dar, "dar", "", "path to the DAR file (required)")
	rootCmd.Flags().StringVar(&output, "output", "", "output directory where generated Go files will be saved (required)")
	rootCmd.Flags().StringVar(&pkg, "go_package", "", "Go package name for generated code (required)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "enable debug logging")

	rootCmd.MarkFlagRequired("dar")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("go_package")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func removePackageID(filename string) string {
	lastHyphen := strings.LastIndex(filename, "-")
	if lastHyphen == -1 {
		return filename
	}

	potentialHash := filename[lastHyphen+1:]
	if len(potentialHash) == 64 {
		allHex := true
		for _, ch := range potentialHash {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				allHex = false
				break
			}
		}
		if allHex {
			return filename[:lastHyphen]
		}
	}

	return filename
}

func processDalf(dalfRelPath, unzippedPath, pkgName, sdkVersion, outputDir string, isMainDalf bool) error {
	dalfFullPath := filepath.Join(unzippedPath, dalfRelPath)
	dalfContent, err := os.ReadFile(dalfFullPath)
	if err != nil {
		return fmt.Errorf("failed to read dalf file '%s': %w", dalfFullPath, err)
	}

	manifest := &model.Manifest{
		SdkVersion: sdkVersion,
		MainDalf:   dalfRelPath,
	}

	pkg, err := codegen.GetAST(dalfContent, manifest)
	if err != nil {
		return fmt.Errorf("failed to generate AST: %w", err)
	}

	res, err := codegen.Bind(pkgName, pkg.PackageID, sdkVersion, pkg.Structs, isMainDalf)
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %w", err)
	}

	parts := strings.Split(dalfRelPath, "/")
	var baseFileName string
	if len(parts) > 1 {
		dalfFileName := parts[len(parts)-1]
		baseFileName = strings.TrimSuffix(dalfFileName, ".dalf")
	} else {
		baseFileName = strings.TrimSuffix(dalfRelPath, ".dalf")
	}

	baseFileName = removePackageID(baseFileName)

	sanitizedFileName := strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(baseFileName), ".", "_"), "-", "_")
	fileName := filepath.Join(outputDir, sanitizedFileName+".go")

	err = os.WriteFile(fileName, []byte(res), 0o644)
	if err != nil {
		return fmt.Errorf("failed to save generated file '%s': %w", fileName, err)
	}

	log.Info().Msgf("successfully generated Go code: %s", fileName)
	return nil
}

func runCodeGen(dar, outputDir, pkgFile string, debugMode bool) error {
	if debugMode {
		log.Info().Msg("debug mode enabled")
	}

	unzippedPath, err := codegen.UnzipDar(dar, nil)
	if err != nil {
		return fmt.Errorf("failed to unzip dar file '%s': %w", dar, err)
	}
	defer os.RemoveAll(unzippedPath)

	manifest, err := codegen.GetManifest(unzippedPath)
	if err != nil {
		return fmt.Errorf("failed to get manifest from '%s': %w", unzippedPath, err)
	}

	err = os.MkdirAll(outputDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	log.Info().Msgf("processing MainDalf: %s", manifest.MainDalf)
	err = processDalf(manifest.MainDalf, unzippedPath, pkgFile, manifest.SdkVersion, outputDir, true)
	if err != nil {
		return fmt.Errorf("failed to process MainDalf: %w", err)
	}

	successCount := 1
	failedCount := 0
	skippedCount := 0

	for _, dalf := range manifest.Dalfs {
		if dalf == manifest.MainDalf {
			log.Debug().Msgf("skipping MainDalf (already processed): %s", dalf)
			skippedCount++
			continue
		}

		dalfLower := strings.ToLower(dalf)
		if strings.Contains(dalfLower, "prim") || strings.Contains(dalfLower, "stdlib") {
			log.Debug().Msgf("skipping dalf (prim/stdlib): %s", dalf)
			skippedCount++
			continue
		}

		log.Info().Msgf("processing dependency dalf: %s", dalf)
		err = processDalf(dalf, unzippedPath, pkgFile, manifest.SdkVersion, outputDir, false)
		if err != nil {
			log.Error().Err(err).Msgf("failed to process dalf: %s", dalf)
			failedCount++
			continue
		}
		successCount++
	}

	log.Info().Msgf("code generation summary: %d succeeded, %d failed, %d skipped", successCount, failedCount, skippedCount)

	if successCount == 0 {
		return fmt.Errorf("all dalf files failed to process")
	}

	return nil
}
