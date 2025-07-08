// cmd/clone.go - Enhanced clone command with real AWS integration
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	awsclient "iam-role-cloner/internal/aws"
	"iam-role-cloner/internal/logger"
)

// Enhanced configuration struct
type CloneConfig struct {
	SourceProfile string
	DestProfile   string
	SourcePattern string
	DestPattern   string
	Roles         []string
	Verbose       bool
	DryRun        bool
	LogFile       string
}

// Enhanced cloneCmd with real AWS functionality
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone IAM roles between AWS profiles",
	Long: `Clone IAM roles from source to destination AWS profile with pattern replacement.

This command will:
1. Validate AWS profiles and credentials
2. Get pattern replacement rules (e.g., dev_ -> prod_)
3. Select roles to clone (with auto-discovery)
4. Clone roles with all policies and tags
5. Apply pattern replacement to names and policy content

Examples:
  iam-role-cloner clone                                    # Interactive mode
  iam-role-cloner clone -s dev -d prod                     # With profiles
  iam-role-cloner clone --dry-run --verbose                # Dry run with details
  iam-role-cloner clone --source-pattern "dev_" --dest-pattern "prod_"`,

	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		sourceProfile, _ := cmd.Flags().GetString("source-profile")
		destProfile, _ := cmd.Flags().GetString("dest-profile")
		sourcePattern, _ := cmd.Flags().GetString("source-pattern")
		destPattern, _ := cmd.Flags().GetString("dest-pattern")
		logFile, _ := cmd.Flags().GetString("log-file")

		// Default log file name
		if logFile == "" {
			logFile = fmt.Sprintf("iam-clone-%s.log", time.Now().Format("20060102-150405"))
		}

		config := &CloneConfig{
			SourceProfile: sourceProfile,
			DestProfile:   destProfile,
			SourcePattern: sourcePattern,
			DestPattern:   destPattern,
			Verbose:       verbose,
			DryRun:        dryRun,
			LogFile:       logFile,
		}

		runEnhancedClone(config)
	},
}

func runEnhancedClone(config *CloneConfig) {
	// Initialize logger
	log, err := logger.New(config.Verbose, config.LogFile)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Header("🚀 IAM Role Cloning Wizard")

	if config.DryRun {
		log.Warning("Running in DRY-RUN mode - no actual changes will be made")
	}

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Get and validate profiles
	if err := getAndValidateProfiles(config, log, reader); err != nil {
		log.Error(fmt.Sprintf("Profile validation failed: %v", err))
		os.Exit(1)
	}

	// Step 2: Get pattern configuration
	if err := getPatternConfiguration(config, log, reader); err != nil {
		log.Error(fmt.Sprintf("Pattern configuration failed: %v", err))
		os.Exit(1)
	}

	// Step 3: Discover and select roles
	if err := discoverAndSelectRoles(config, log, reader); err != nil {
		log.Error(fmt.Sprintf("Role selection failed: %v", err))
		os.Exit(1)
	}

	// Step 4: Show summary and confirm
	if !showSummaryAndConfirm(config, log, reader) {
		log.Info("Operation cancelled by user")
		return
	}

	// Step 5: Perform the cloning
	if err := performCloning(config, log); err != nil {
		log.Error(fmt.Sprintf("Cloning failed: %v", err))
		os.Exit(1)
	}

	log.Success("🎉 Role cloning completed successfully!")
	log.Info(fmt.Sprintf("Log file saved: %s", config.LogFile))
}

func getAndValidateProfiles(config *CloneConfig, log *logger.Logger, reader *bufio.Reader) error {
	log.Info("Step 1: Profile Configuration and Validation")
	log.Separator()

	// Get source profile
	if config.SourceProfile == "" {
		fmt.Print("Enter source AWS profile: ")
		profile, _ := reader.ReadString('\n')
		config.SourceProfile = strings.TrimSpace(profile)
	}

	// Get destination profile
	if config.DestProfile == "" {
		fmt.Print("Enter destination AWS profile: ")
		profile, _ := reader.ReadString('\n')
		config.DestProfile = strings.TrimSpace(profile)
	}

	// Validate source profile
	log.Info(fmt.Sprintf("Validating source profile: %s", config.SourceProfile))
	sourceClient, err := awsclient.NewClient(config.SourceProfile)
	if err != nil {
		return fmt.Errorf("failed to create source client: %v", err)
	}

	ctx := context.Background()
	sourceIdentity, err := sourceClient.ValidateCredentials(ctx)
	if err != nil {
		return fmt.Errorf("source profile validation failed: %v", err)
	}

	log.Success(fmt.Sprintf("Source profile validated - Account: %s", *sourceIdentity.Account))
	log.Debug(fmt.Sprintf("Source ARN: %s", *sourceIdentity.Arn))

	// Validate destination profile
	log.Info(fmt.Sprintf("Validating destination profile: %s", config.DestProfile))
	destClient, err := awsclient.NewClient(config.DestProfile)
	if err != nil {
		return fmt.Errorf("failed to create destination client: %v", err)
	}

	destIdentity, err := destClient.ValidateCredentials(ctx)
	if err != nil {
		return fmt.Errorf("destination profile validation failed: %v", err)
	}

	log.Success(fmt.Sprintf("Destination profile validated - Account: %s", *destIdentity.Account))
	log.Debug(fmt.Sprintf("Destination ARN: %s", *destIdentity.Arn))

	if *sourceIdentity.Account == *destIdentity.Account {
		log.Warning("Source and destination are the same AWS account")
		fmt.Print("Continue anyway? (y/n): ")
		confirm, _ := reader.ReadString('\n')
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
			return fmt.Errorf("operation cancelled - same account")
		}
	}

	return nil
}

func getPatternConfiguration(config *CloneConfig, log *logger.Logger, reader *bufio.Reader) error {
	log.Info("Step 2: Pattern Configuration")
	log.Separator()

	if config.SourcePattern == "" {
		fmt.Print("Enter source pattern (e.g., 'dev_', 'staging-'): ")
		pattern, _ := reader.ReadString('\n')
		config.SourcePattern = strings.TrimSpace(pattern)
	}

	if config.DestPattern == "" {
		fmt.Print("Enter destination pattern (e.g., 'prod_', 'live-'): ")
		pattern, _ := reader.ReadString('\n')
		config.DestPattern = strings.TrimSpace(pattern)
	}

	log.Success(fmt.Sprintf("Pattern replacement: '%s' → '%s'", config.SourcePattern, config.DestPattern))

	// Show examples
	exampleRole := fmt.Sprintf("%sexample_role", config.SourcePattern)
	newExampleRole := awsclient.GenerateNewRoleName(exampleRole, config.SourcePattern, config.DestPattern)
	log.Info(fmt.Sprintf("Example transformation: %s → %s", exampleRole, newExampleRole))

	return nil
}

func discoverAndSelectRoles(config *CloneConfig, log *logger.Logger, reader *bufio.Reader) error {
	log.Info("Step 3: Role Discovery and Selection")
	log.Separator()

	// Create source client for role discovery
	sourceClient, err := awsclient.NewClient(config.SourceProfile)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Discover roles with source pattern
	log.Info("Discovering roles in source account...")
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()

	allRoles, err := sourceClient.ListRoles(ctx, config.SourcePattern)
	s.Stop()

	if err != nil {
		return fmt.Errorf("failed to discover roles: %v", err)
	}

	if len(allRoles) == 0 {
		log.Warning(fmt.Sprintf("No roles found with pattern '%s'", config.SourcePattern))
		return getRolesManually(config, log, reader)
	}

	log.Success(fmt.Sprintf("Found %d roles with pattern '%s'", len(allRoles), config.SourcePattern))

	// Show discovered roles
	fmt.Println("\nDiscovered roles:")
	for i, role := range allRoles {
		newRole := awsclient.GenerateNewRoleName(role, config.SourcePattern, config.DestPattern)
		fmt.Printf("  %d. %s → %s\n", i+1, role, newRole)
	}

	// Let user select roles
	fmt.Print("\nEnter role numbers to clone (e.g., 1,3,5 or 'all'): ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if strings.ToLower(selection) == "all" {
		config.Roles = allRoles
		log.Success(fmt.Sprintf("Selected all %d roles", len(allRoles)))
	} else {
		selectedRoles, err := parseRoleSelection(selection, allRoles)
		if err != nil {
			return fmt.Errorf("invalid selection: %v", err)
		}
		config.Roles = selectedRoles
		log.Success(fmt.Sprintf("Selected %d roles", len(selectedRoles)))
	}

	return nil
}

func getRolesManually(config *CloneConfig, log *logger.Logger, reader *bufio.Reader) error {
	log.Info("Manual role entry mode")

	fmt.Print("How many roles do you want to clone? (1-20): ")
	countStr, _ := reader.ReadString('\n')
	countStr = strings.TrimSpace(countStr)

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 20 {
		return fmt.Errorf("invalid role count: %s", countStr)
	}

	config.Roles = make([]string, 0, count)
	for i := 0; i < count; i++ {
		fmt.Printf("Enter role name #%d: ", i+1)
		role, _ := reader.ReadString('\n')
		role = strings.TrimSpace(role)

		if role != "" {
			config.Roles = append(config.Roles, role)
		} else {
			i-- // Retry
		}
	}

	return nil
}

func parseRoleSelection(selection string, allRoles []string) ([]string, error) {
	parts := strings.Split(selection, ",")
	var selectedRoles []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		index, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}

		if index < 1 || index > len(allRoles) {
			return nil, fmt.Errorf("number out of range: %d", index)
		}

		selectedRoles = append(selectedRoles, allRoles[index-1])
	}

	return selectedRoles, nil
}

func showSummaryAndConfirm(config *CloneConfig, log *logger.Logger, reader *bufio.Reader) bool {
	log.Info("Step 4: Configuration Summary")
	log.Separator()

	fmt.Printf("Source Profile:      %s\n", config.SourceProfile)
	fmt.Printf("Destination Profile: %s\n", config.DestProfile)
	fmt.Printf("Pattern Replacement: '%s' → '%s'\n", config.SourcePattern, config.DestPattern)
	fmt.Printf("Dry Run:            %v\n", config.DryRun)
	fmt.Printf("Verbose Logging:    %v\n", config.Verbose)
	fmt.Printf("Log File:           %s\n", config.LogFile)
	fmt.Println("\nRoles to clone:")

	for i, role := range config.Roles {
		newRole := awsclient.GenerateNewRoleName(role, config.SourcePattern, config.DestPattern)
		fmt.Printf("  %d. %s → %s\n", i+1, role, newRole)
	}

	fmt.Print("\nProceed with cloning? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	return confirm == "y" || confirm == "yes"
}

func performCloning(config *CloneConfig, log *logger.Logger) error {
	log.Info("Step 5: Role Cloning Process")
	log.Separator()

	// Create AWS clients
	sourceClient, err := awsclient.NewClient(config.SourceProfile)
	if err != nil {
		return fmt.Errorf("failed to create source client: %v", err)
	}

	var destClient *awsclient.Client
	if !config.DryRun {
		destClient, err = awsclient.NewClient(config.DestProfile)
		if err != nil {
			return fmt.Errorf("failed to create destination client: %v", err)
		}
	}

	ctx := context.Background()
	successCount := 0

	for i, role := range config.Roles {
		newRole := awsclient.GenerateNewRoleName(role, config.SourcePattern, config.DestPattern)
		log.Progress(i+1, len(config.Roles), fmt.Sprintf("Cloning: %s → %s", role, newRole))

		if err := cloneSingleRole(ctx, sourceClient, destClient, role, newRole, config, log); err != nil {
			log.Error(fmt.Sprintf("Failed to clone %s: %v", role, err))
			continue
		}

		successCount++
		log.Success(fmt.Sprintf("Successfully cloned: %s → %s", role, newRole))
	}

	log.Separator()
	log.Success(fmt.Sprintf("Cloning completed: %d/%d roles successful", successCount, len(config.Roles)))

	if config.DryRun {
		log.Info("This was a dry run. Use without --dry-run to perform actual cloning.")
	}

	return nil
}

func cloneSingleRole(ctx context.Context, sourceClient, destClient *awsclient.Client,
	sourceRole, destRole string, config *CloneConfig, log *logger.Logger) error {

	// Step 1: Get role information
	log.Debug(fmt.Sprintf("  Getting role information for: %s", sourceRole))
	roleInfo, err := sourceClient.GetRoleInfo(ctx, sourceRole)
	if err != nil {
		return fmt.Errorf("failed to get role info: %v", err)
	}

	log.Debug(fmt.Sprintf("  Retrieved role info: %d managed policies, %d inline policies, %d tags",
		len(roleInfo.ManagedPolicies), len(roleInfo.InlinePolicies), len(roleInfo.Tags)))

	if config.DryRun {
		log.Info("  [DRY RUN] Would create role and copy policies/tags")

		// Process the trust policy to show what would actually be sent to AWS
		processedTrustPolicy := awsclient.ReplacePatternInJSON(
			roleInfo.TrustPolicy, config.SourcePattern, config.DestPattern)

		if config.Verbose {
			log.Debug(fmt.Sprintf("  [DRY RUN] Original trust policy: %s", roleInfo.TrustPolicy))
			log.Debug(fmt.Sprintf("  [DRY RUN] Processed trust policy: %s", processedTrustPolicy))

			// Show managed policies that would be attached
			if len(roleInfo.ManagedPolicies) > 0 {
				log.Debug(fmt.Sprintf("  [DRY RUN] Would attach %d managed policies:", len(roleInfo.ManagedPolicies)))
				for _, policy := range roleInfo.ManagedPolicies {
					log.Debug(fmt.Sprintf("    - %s", policy))
				}
			}

			// Show inline policies that would be created
			if len(roleInfo.InlinePolicies) > 0 {
				log.Debug(fmt.Sprintf("  [DRY RUN] Would create %d inline policies:", len(roleInfo.InlinePolicies)))
				for policyName := range roleInfo.InlinePolicies {
					newPolicyName := awsclient.GenerateNewRoleName(policyName, config.SourcePattern, config.DestPattern)
					log.Debug(fmt.Sprintf("    - %s → %s", policyName, newPolicyName))
				}
			}

			// Show tags that would be copied
			if len(roleInfo.Tags) > 0 {
				log.Debug(fmt.Sprintf("  [DRY RUN] Would copy %d tags:", len(roleInfo.Tags)))
				for key, value := range roleInfo.Tags {
					if key == "Environment" {
						envValue := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(
							config.DestPattern, "_"), "-"), ".")
						log.Debug(fmt.Sprintf("    - %s: %s → %s", key, value, envValue))
					} else {
						newValue := awsclient.ReplacePatternInJSON(value, config.SourcePattern, config.DestPattern)
						if newValue != value {
							log.Debug(fmt.Sprintf("    - %s: %s → %s", key, value, newValue))
						} else {
							log.Debug(fmt.Sprintf("    - %s: %s", key, value))
						}
					}
				}
			}
		}
		return nil
	}

	// Check if destination role already exists
	if destClient.RoleExists(ctx, destRole) {
		return fmt.Errorf("destination role already exists: %s", destRole)
	}

	// Step 2: Create the role with pattern-replaced trust policy
	log.Debug("  Creating new role...")
	processedTrustPolicy := awsclient.ReplacePatternInJSON(
		roleInfo.TrustPolicy, config.SourcePattern, config.DestPattern)

	// Debug: Show the processed trust policy if verbose
	if config.Verbose {
		log.Debug(fmt.Sprintf("  Original trust policy: %s", roleInfo.TrustPolicy))
		log.Debug(fmt.Sprintf("  Processed trust policy: %s", processedTrustPolicy))
	}

	description := fmt.Sprintf("Cloned from %s on %s", sourceRole, time.Now().Format("2006-01-02 15:04:05"))
	if err := destClient.CreateRole(ctx, destRole, processedTrustPolicy, description); err != nil {
		// Enhanced error message with policy content
		if config.Verbose {
			log.Error(fmt.Sprintf("  Failed trust policy content: %s", processedTrustPolicy))
		}
		return fmt.Errorf("failed to create role: %v", err)
	}

	log.Debug("  Role created successfully")

	// Step 3: Attach managed policies
	log.Debug(fmt.Sprintf("  Attaching %d managed policies...", len(roleInfo.ManagedPolicies)))
	for _, policyArn := range roleInfo.ManagedPolicies {
		if err := destClient.AttachManagedPolicy(ctx, destRole, policyArn); err != nil {
			log.Warning(fmt.Sprintf("    Failed to attach managed policy %s: %v", policyArn, err))
		} else {
			log.Debug(fmt.Sprintf("    Attached: %s", policyArn))
		}
	}

	// Step 4: Create inline policies with pattern replacement
	log.Debug(fmt.Sprintf("  Creating %d inline policies...", len(roleInfo.InlinePolicies)))
	for policyName, policyDocument := range roleInfo.InlinePolicies {
		newPolicyName := awsclient.GenerateNewRoleName(policyName, config.SourcePattern, config.DestPattern)
		processedDocument := awsclient.ReplacePatternInJSON(
			policyDocument, config.SourcePattern, config.DestPattern)

		if config.Verbose {
			log.Debug(fmt.Sprintf("    Creating inline policy: %s", newPolicyName))
			log.Debug(fmt.Sprintf("    Policy document preview: %.100s...", processedDocument))
		}

		if err := destClient.CreateInlinePolicy(ctx, destRole, newPolicyName, processedDocument); err != nil {
			log.Warning(fmt.Sprintf("    Failed to create inline policy %s: %v", newPolicyName, err))
		} else {
			log.Debug(fmt.Sprintf("    Created inline policy: %s", newPolicyName))
		}
	}

	// Step 5: Copy and update tags
	if len(roleInfo.Tags) > 0 {
		log.Debug(fmt.Sprintf("  Copying %d tags...", len(roleInfo.Tags)))
		processedTags := make(map[string]string)

		for key, value := range roleInfo.Tags {
			// Replace patterns in tag values and update Environment tag
			if key == "Environment" {
				// Set environment to destination pattern (cleaned)
				envValue := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(
					config.DestPattern, "_"), "-"), ".")
				processedTags[key] = envValue
			} else {
				processedTags[key] = awsclient.ReplacePatternInJSON(value, config.SourcePattern, config.DestPattern)
			}
		}

		if config.Verbose {
			log.Debug(fmt.Sprintf("    Processed tags: %+v", processedTags))
		}

		if err := destClient.TagRole(ctx, destRole, processedTags); err != nil {
			log.Warning(fmt.Sprintf("    Failed to copy tags: %v", err))
		} else {
			log.Debug("    Tags copied successfully")
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	// Command-specific flags
	cloneCmd.Flags().StringP("source-profile", "s", "", "Source AWS profile")
	cloneCmd.Flags().StringP("dest-profile", "d", "", "Destination AWS profile")
	cloneCmd.Flags().String("source-pattern", "", "Source environment pattern (e.g., 'dev_')")
	cloneCmd.Flags().String("dest-pattern", "", "Destination environment pattern (e.g., 'prod_')")
	cloneCmd.Flags().String("log-file", "", "Log file path (default: auto-generated)")

	// Global flags
	cloneCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cloneCmd.Flags().Bool("dry-run", false, "Show what would be done without actually doing it")
}
