package tags

import (
	"fmt"
	"strings"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsNeptuneClusterInstanceTagsRule checks whether the resource is tagged correctly
type AwsNeptuneClusterInstanceTagsRule struct {
	resourceType  string
	attributeName string
}


// NewAwsNeptuneClusterInstanceTagsRule returns new tags rule with default attributes
func NewAwsNeptuneClusterInstanceTagsRule() *AwsNeptuneClusterInstanceTagsRule {
	return &AwsNeptuneClusterInstanceTagsRule{
		resourceType:  "aws_neptune_cluster_instance",
		attributeName: "tags",
	}
}

// Name returns the rule name
func (r *AwsNeptuneClusterInstanceTagsRule) Name() string {
	return "aws_resource_tags_aws_neptune_cluster_instance"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsNeptuneClusterInstanceTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsNeptuneClusterInstanceTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsNeptuneClusterInstanceTagsRule) Link() string {
	return ""
}

// Check checks for matching tags
func (r *AwsNeptuneClusterInstanceTagsRule) Check(runner *tflint.Runner) error {
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var resourceTags map[string]string
		err := runner.EvaluateExpr(attribute.Expr, &resourceTags)
		tags := []string{}
		for k, _ := range resourceTags {
			tags = append(tags, k)
		}

		return runner.EnsureNoError(err, func() error {
			configTags := runner.GetConfigTags()
			hash := make(map[string]bool)
			for _, k := range tags {
				hash[k] = true
			}
			var found []string
			for _, tag := range configTags {
				if _, ok := hash[tag]; ok {
					found = append(found, tag)
				}
			}
			if len(found) != len(configTags) {
				wanted := strings.Join(sort.StringSlice(configTags), ",")
				found := strings.Join(sort.StringSlice(tags), ",")
				runner.EmitIssue(r, fmt.Sprintf("Wanted tags: %v, found: %v\n", wanted, found), attribute.Expr.Range())
			}
			return nil
		})
	})
}
