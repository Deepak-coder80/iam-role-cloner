// internal/aws/client.go - AWS operations
package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Client struct {
	iam    *iam.Client
	sts    *sts.Client
	config aws.Config
}

type RoleInfo struct {
	RoleName        string
	Description     string
	TrustPolicy     string
	ManagedPolicies []string
	InlinePolicies  map[string]string
	Tags            map[string]string
}

// NewClient creates a new AWS client with the specified profile
func NewClient(profile string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for profile %s: %v", profile, err)
	}

	return &Client{
		iam:    iam.NewFromConfig(cfg),
		sts:    sts.NewFromConfig(cfg),
		config: cfg,
	}, nil
}

// ValidateCredentials checks if the AWS credentials are valid
func (c *Client) ValidateCredentials(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	return c.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}

// ListRoles lists all IAM roles, optionally filtered by prefix
func (c *Client) ListRoles(ctx context.Context, prefix string) ([]string, error) {
	var allRoles []string

	paginator := iam.NewListRolesPaginator(c.iam, &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list roles: %v", err)
		}

		for _, role := range output.Roles {
			roleName := *role.RoleName
			if prefix == "" || strings.HasPrefix(roleName, prefix) {
				allRoles = append(allRoles, roleName)
			}
		}
	}

	return allRoles, nil
}

// RoleExists checks if a role exists
func (c *Client) RoleExists(ctx context.Context, roleName string) bool {
	_, err := c.iam.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	return err == nil
}

// GetRoleInfo retrieves complete information about a role
func (c *Client) GetRoleInfo(ctx context.Context, roleName string) (*RoleInfo, error) {
	// Get basic role info
	roleOutput, err := c.iam.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role %s: %v", roleName, err)
	}

	role := roleOutput.Role
	roleInfo := &RoleInfo{
		RoleName:    *role.RoleName,
		Description: "",
		Tags:        make(map[string]string),
	}

	if role.Description != nil {
		roleInfo.Description = *role.Description
	}

	// Process trust policy properly
	trustPolicy, err := processPolicyDocument(role.AssumeRolePolicyDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to process trust policy: %v", err)
	}
	roleInfo.TrustPolicy = trustPolicy

	// Get managed policies
	managedPolicies, err := c.getManagedPolicies(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed policies: %v", err)
	}
	roleInfo.ManagedPolicies = managedPolicies

	// Get inline policies
	inlinePolicies, err := c.getInlinePolicies(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get inline policies: %v", err)
	}
	roleInfo.InlinePolicies = inlinePolicies

	// Get tags
	tags, err := c.getRoleTags(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	roleInfo.Tags = tags

	return roleInfo, nil
}

// CreateRole creates a new IAM role
func (c *Client) CreateRole(ctx context.Context, roleName, trustPolicy, description string) error {
	input := &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
	}

	if description != "" {
		input.Description = aws.String(description)
	}

	_, err := c.iam.CreateRole(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create role %s: %v", roleName, err)
	}

	return nil
}

// AttachManagedPolicy attaches a managed policy to a role
func (c *Client) AttachManagedPolicy(ctx context.Context, roleName, policyArn string) error {
	_, err := c.iam.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		return fmt.Errorf("failed to attach policy %s to role %s: %v", policyArn, roleName, err)
	}

	return nil
}

// CreateInlinePolicy creates an inline policy for a role
func (c *Client) CreateInlinePolicy(ctx context.Context, roleName, policyName, policyDocument string) error {
	_, err := c.iam.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})

	if err != nil {
		return fmt.Errorf("failed to create inline policy %s for role %s: %v", policyName, roleName, err)
	}

	return nil
}

// TagRole adds tags to a role
func (c *Client) TagRole(ctx context.Context, roleName string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}

	var iamTags []types.Tag
	for key, value := range tags {
		iamTags = append(iamTags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	_, err := c.iam.TagRole(ctx, &iam.TagRoleInput{
		RoleName: aws.String(roleName),
		Tags:     iamTags,
	})

	if err != nil {
		return fmt.Errorf("failed to tag role %s: %v", roleName, err)
	}

	return nil
}

// Helper function to get managed policies
func (c *Client) getManagedPolicies(ctx context.Context, roleName string) ([]string, error) {
	var policies []string

	paginator := iam.NewListAttachedRolePoliciesPaginator(c.iam, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, policy := range output.AttachedPolicies {
			policies = append(policies, *policy.PolicyArn)
		}
	}

	return policies, nil
}

// Helper function to get inline policies
func (c *Client) getInlinePolicies(ctx context.Context, roleName string) (map[string]string, error) {
	policies := make(map[string]string)

	// List policy names
	listOutput, err := c.iam.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, err
	}

	// Get each policy document
	for _, policyName := range listOutput.PolicyNames {
		getOutput, err := c.iam.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return nil, err
		}

		// Process policy document properly
		policyDoc, err := processPolicyDocument(getOutput.PolicyDocument)
		if err != nil {
			return nil, fmt.Errorf("failed to process inline policy %s: %v", policyName, err)
		}

		policies[policyName] = policyDoc
	}

	return policies, nil
}

// Helper function to properly process AWS policy documents
func processPolicyDocument(policyDoc interface{}) (string, error) {
	switch v := policyDoc.(type) {
	case string:
		return processStringPolicy(v)

	case *string:
		if v == nil {
			return "", fmt.Errorf("policy document is nil")
		}
		return processStringPolicy(*v)

	case map[string]interface{}:
		// If it's a map, convert to properly formatted JSON
		bytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal policy document: %v", err)
		}
		return string(bytes), nil

	default:
		// For any other type, try to marshal it
		bytes, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal policy document of type %T: %v", v, err)
		}
		return string(bytes), nil
	}
}

// Helper function to process string policies (handles both string and *string cases)
func processStringPolicy(policyStr string) (string, error) {
	// First, try to URL decode the string (AWS often returns URL-encoded JSON)
	decoded, err := url.QueryUnescape(policyStr)
	if err != nil {
		// If URL decoding fails, use the original string
		decoded = policyStr
	}

	// Validate it's proper JSON and format it nicely
	var temp interface{}
	if err := json.Unmarshal([]byte(decoded), &temp); err != nil {
		return "", fmt.Errorf("invalid JSON after decoding: %v (original: %s)", err, policyStr)
	}

	// Re-marshal to ensure consistent formatting
	bytes, err := json.MarshalIndent(temp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to re-marshal policy: %v", err)
	}

	return string(bytes), nil
}

// Helper function to get role tags
func (c *Client) getRoleTags(ctx context.Context, roleName string) (map[string]string, error) {
	tags := make(map[string]string)

	output, err := c.iam.ListRoleTags(ctx, &iam.ListRoleTagsInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, err
	}

	for _, tag := range output.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, nil
}

// ReplacePatternInJSON replaces patterns in JSON strings
func ReplacePatternInJSON(jsonStr, sourcePattern, destPattern string) string {
	return strings.ReplaceAll(jsonStr, sourcePattern, destPattern)
}

// GenerateNewRoleName generates new role name with pattern replacement
func GenerateNewRoleName(originalName, sourcePattern, destPattern string) string {
	return strings.ReplaceAll(originalName, sourcePattern, destPattern)
}
