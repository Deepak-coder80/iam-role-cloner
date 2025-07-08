// cmd/list.go - List roles command
package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	awsclient "iam-role-cloner/internal/aws"
	"iam-role-cloner/internal/logger"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List IAM roles in an AWS profile",
	Long: `List IAM roles in the specified AWS profile, optionally filtered by pattern.

This command helps you discover roles before cloning and verify what exists
in your source and destination accounts.

Examples:
  iam-role-cloner list --profile dev                    # List all roles in dev profile
  iam-role-cloner list -p prod --pattern "prod_"        # List roles starting with "prod_"
  iam-role-cloner list --profile staging --details      # List with detailed information
  iam-role-cloner list -p dev --pattern "app" --sort    # List and sort roles containing "app"`,

	Run: func(cmd *cobra.Command, args []string) {
		profile, _ := cmd.Flags().GetString("profile")
		pattern, _ := cmd.Flags().GetString("pattern")
		details, _ := cmd.Flags().GetBool("details")
		sortRoles, _ := cmd.Flags().GetBool("sort")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if profile == "" {
			fmt.Println("❌ Error: --profile flag is required")
			fmt.Println("Usage: iam-role-cloner list --profile <profile-name>")
			return
		}

		runListCommand(profile, pattern, details, sortRoles, verbose)
	},
}

func runListCommand(profile, pattern string, details, sortRoles, verbose bool) {
	// Initialize logger (no file logging for list command)
	log, err := logger.New(verbose, "")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer log.Close()

	log.Header(fmt.Sprintf("📋 IAM Roles in Profile: %s", profile))

	// Create AWS client
	log.Info(fmt.Sprintf("Connecting to AWS profile: %s", profile))
	client, err := awsclient.NewClient(profile)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create AWS client: %v", err))
		return
	}

	ctx := context.Background()

	// Validate credentials
	log.Debug("Validating AWS credentials...")
	identity, err := client.ValidateCredentials(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to validate credentials: %v", err))
		return
	}

	log.Success(fmt.Sprintf("Connected to AWS Account: %s", *identity.Account))
	log.Debug(fmt.Sprintf("User/Role ARN: %s", *identity.Arn))

	// List roles
	log.Info("Discovering IAM roles...")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Fetching roles..."
	s.Start()

	roles, err := client.ListRoles(ctx, "")
	s.Stop()

	if err != nil {
		log.Error(fmt.Sprintf("Failed to list roles: %v", err))
		return
	}

	// Filter roles by pattern if specified
	var filteredRoles []string
	if pattern != "" {
		log.Info(fmt.Sprintf("Filtering roles by pattern: '%s'", pattern))
		for _, role := range roles {
			if strings.Contains(strings.ToLower(role), strings.ToLower(pattern)) {
				filteredRoles = append(filteredRoles, role)
			}
		}
		roles = filteredRoles
		log.Info(fmt.Sprintf("Found %d roles matching pattern", len(roles)))
	} else {
		log.Info(fmt.Sprintf("Found %d total roles", len(roles)))
	}

	if len(roles) == 0 {
		log.Warning("No roles found")
		return
	}

	// Sort roles if requested
	if sortRoles {
		log.Debug("Sorting roles alphabetically...")
		sort.Strings(roles)
	}

	// Display roles
	log.Separator()
	if details {
		displayDetailedRoles(client, roles, log)
	} else {
		displaySimpleRoles(roles, pattern, log)
	}

	// Summary
	log.Separator()
	log.Success(fmt.Sprintf("Listed %d roles successfully", len(roles)))
}

func displaySimpleRoles(roles []string, pattern string, log *logger.Logger) {
	fmt.Printf("\n📝 Role Names:\n")
	fmt.Println("=" + strings.Repeat("=", 50))

	for i, role := range roles {
		// Highlight the pattern if specified
		displayName := role
		if pattern != "" {
			displayName = strings.ReplaceAll(role, pattern, fmt.Sprintf("**%s**", pattern))
		}

		fmt.Printf("%3d. %s\n", i+1, displayName)
	}
}

func displayDetailedRoles(client *awsclient.Client, roles []string, log *logger.Logger) {
	fmt.Printf("\n📋 Detailed Role Information:\n")
	fmt.Println("=" + strings.Repeat("=", 80))

	ctx := context.Background()

	for i, roleName := range roles {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(roles), roleName)
		fmt.Println(strings.Repeat("-", len(roleName)+10))

		// Get detailed role information
		log.Debug(fmt.Sprintf("Getting details for role: %s", roleName))

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Getting details for %s...", roleName)
		s.Start()

		roleInfo, err := client.GetRoleInfo(ctx, roleName)
		s.Stop()

		if err != nil {
			fmt.Printf("❌ Error getting role details: %v\n", err)
			continue
		}

		// Display role details
		fmt.Printf("📝 Description: %s\n", getDescription(roleInfo.Description))
		fmt.Printf("🔒 Managed Policies: %d\n", len(roleInfo.ManagedPolicies))

		if len(roleInfo.ManagedPolicies) > 0 && len(roleInfo.ManagedPolicies) <= 5 {
			for _, policy := range roleInfo.ManagedPolicies {
				policyName := extractPolicyName(policy)
				fmt.Printf("    • %s\n", policyName)
			}
		} else if len(roleInfo.ManagedPolicies) > 5 {
			for j, policy := range roleInfo.ManagedPolicies[:3] {
				policyName := extractPolicyName(policy)
				fmt.Printf("    • %s\n", policyName)
				if j == 2 {
					fmt.Printf("    • ... and %d more\n", len(roleInfo.ManagedPolicies)-3)
				}
			}
		}

		fmt.Printf("📄 Inline Policies: %d\n", len(roleInfo.InlinePolicies))
		if len(roleInfo.InlinePolicies) > 0 {
			for policyName := range roleInfo.InlinePolicies {
				fmt.Printf("    • %s\n", policyName)
			}
		}

		fmt.Printf("🏷️  Tags: %d\n", len(roleInfo.Tags))
		if len(roleInfo.Tags) > 0 {
			for key, value := range roleInfo.Tags {
				if len(value) > 30 {
					value = value[:30] + "..."
				}
				fmt.Printf("    • %s: %s\n", key, value)
			}
		}

		// Show trust relationship summary
		if strings.Contains(roleInfo.TrustPolicy, "ec2.amazonaws.com") {
			fmt.Printf("🖥️  Trust: EC2 Service Role\n")
		} else if strings.Contains(roleInfo.TrustPolicy, "lambda.amazonaws.com") {
			fmt.Printf("🚀 Trust: Lambda Service Role\n")
		} else if strings.Contains(roleInfo.TrustPolicy, "sts:AssumeRole") {
			fmt.Printf("👤 Trust: Cross-Account Role\n")
		} else {
			fmt.Printf("🔗 Trust: Custom Trust Policy\n")
		}
	}
}

func getDescription(desc string) string {
	if desc == "" {
		return "(No description)"
	}
	if len(desc) > 80 {
		return desc[:80] + "..."
	}
	return desc
}

func extractPolicyName(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return arn
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Required flags
	listCmd.Flags().StringP("profile", "p", "", "AWS profile to use (required)")
	listCmd.MarkFlagRequired("profile")

	// Optional flags
	listCmd.Flags().String("pattern", "", "Filter roles by pattern (case-insensitive)")
	listCmd.Flags().Bool("details", false, "Show detailed information for each role")
	listCmd.Flags().Bool("sort", false, "Sort roles alphabetically")
	listCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}
