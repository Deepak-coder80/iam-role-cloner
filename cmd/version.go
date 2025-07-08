// cmd/version.go - Version command
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build information (these will be set at build time)
var (
	Version   = "1.0.0"
	GitCommit = "dev"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
	Platform  = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for the IAM Role Cloner CLI tool.

This includes the version number, build information, and system details.`,

	Run: func(cmd *cobra.Command, args []string) {
		detailed, _ := cmd.Flags().GetBool("detailed")

		if detailed {
			showDetailedVersion()
		} else {
			showSimpleVersion()
		}
	},
}

func showSimpleVersion() {
	fmt.Printf("IAM Role Cloner v%s\n", Version)
}

func showDetailedVersion() {
	fmt.Println("🚀 IAM Role Cloner")
	fmt.Println("==================")
	fmt.Printf("Version:      %s\n", Version)
	fmt.Printf("Git Commit:   %s\n", GitCommit)
	fmt.Printf("Build Date:   %s\n", BuildDate)
	fmt.Printf("Go Version:   %s\n", GoVersion)
	fmt.Printf("Platform:     %s\n", Platform)
	fmt.Println()
	fmt.Println("📋 Features:")
	fmt.Println("  ✅ Clone IAM roles between AWS accounts")
	fmt.Println("  ✅ Pattern replacement in names and policies")
	fmt.Println("  ✅ Dry-run mode for safe testing")
	fmt.Println("  ✅ Comprehensive logging and error handling")
	fmt.Println("  ✅ Interactive and command-line modes")
	fmt.Println("  ✅ Role discovery and validation")
	fmt.Println()
	fmt.Println("🔗 Repository: https://github.com/your-username/iam-role-cloner")
	fmt.Println("📚 Documentation: https://github.com/your-username/iam-role-cloner/README.md")
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Flags
	versionCmd.Flags().BoolP("explain", "e", false, "Show detailed version information")
}
