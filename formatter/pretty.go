package formatter

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

var colorBold = color.New(color.Bold).SprintfFunc()
var colorHighlight = color.New(color.Bold).Add(color.Underline).SprintFunc()
var colorError = color.New(color.FgRed).SprintFunc()
var colorWarning = color.New(color.FgYellow).SprintFunc()
var colorNotice = color.New(color.FgHiWhite).SprintFunc()

func (f *Formatter) prettyPrint(issues tflint.Issues, err *tflint.Error, sources map[string][]byte) {
	if len(issues) > 0 {
		fmt.Fprintf(f.Stdout, "%d issue(s) found:\n\n", len(issues))

		for _, issue := range issues.Sort() {
			f.printIssueWithSource(issue, sources)
		}
	}

	if err != nil {
		f.printErrors(err, sources)
	}
}

func (f *Formatter) printIssueWithSource(issue *tflint.Issue, sources map[string][]byte) {
	fmt.Fprintf(
		f.Stdout,
		"%s: %s (%s)\n\n",
		colorSeverity(issue.Rule.Type()), colorBold(issue.Message), issue.Rule.Name(),
	)
	fmt.Fprintf(f.Stdout, "  on %s line %d:\n", issue.Range.Filename, issue.Range.Start.Line)

	src := sources[issue.Range.Filename]

	if src == nil {
		fmt.Fprintf(f.Stdout, "   (source code not available)")
	} else {
		sc := hcl.NewRangeScanner(src, issue.Range.Filename, bufio.ScanLines)

		for sc.Scan() {
			lineRange := sc.Range()
			if !lineRange.Overlaps(issue.Range) {
				continue
			}

			beforeRange, highlightedRange, afterRange := lineRange.PartitionAround(issue.Range)
			if highlightedRange.Empty() {
				fmt.Fprintf(f.Stdout, "%4d: %s\n", lineRange.Start.Line, sc.Bytes())
			} else {
				before := beforeRange.SliceBytes(src)
				highlighted := highlightedRange.SliceBytes(src)
				after := afterRange.SliceBytes(src)
				fmt.Fprintf(
					f.Stdout,
					"%4d: %s%s%s\n",
					lineRange.Start.Line,
					before,
					colorHighlight(string(highlighted)),
					after,
				)
			}
		}
	}

	if len(issue.Callers) > 0 {
		fmt.Fprint(f.Stdout, "\nCallers:\n")
		for _, caller := range issue.Callers {
			fmt.Fprintf(f.Stdout, "   %s\n", caller)
		}
	}

	if issue.Rule.Link() != "" {
		fmt.Fprintf(f.Stdout, "\nReference: %s\n", issue.Rule.Link())
	}

	fmt.Fprint(f.Stdout, "\n\n")
}

func (f *Formatter) printErrors(err *tflint.Error, sources map[string][]byte) {
	if diags, ok := err.Cause.(hcl.Diagnostics); ok {
		fmt.Fprintf(f.Stderr, "%s. %d error(s) occurred:\n\n", err.Message, len(diags.Errs()))

		writer := hcl.NewDiagnosticTextWriter(f.Stderr, parseSources(sources), 0, true)
		writer.WriteDiagnostics(diags)
	} else {
		fmt.Fprintf(f.Stderr, "%s. An error occurred:\n\n", err.Message)
		fmt.Fprintf(f.Stderr, "%s: %s\n\n", colorError("Error"), err.Cause.Error())
	}
}

func parseSources(sources map[string][]byte) map[string]*hcl.File {
	ret := map[string]*hcl.File{}
	parser := hclparse.NewParser()

	var file *hcl.File
	var diags hcl.Diagnostics
	for filename, src := range sources {
		if strings.HasSuffix(filename, ".json") {
			file, diags = parser.ParseJSON(src, filename)
		} else {
			file, diags = parser.ParseHCL(src, filename)
		}

		if diags.HasErrors() {
			log.Printf("[WARN] Failed to parse %s. This file is not available in output. Reason: %s", filename, diags.Error())
		}
		ret[filename] = file
	}

	return ret
}

func colorSeverity(severity string) string {
	switch severity {
	case issue.ERROR:
		return colorError(severity)
	case issue.WARNING:
		return colorWarning(severity)
	case issue.NOTICE:
		return colorNotice(severity)
	default:
		panic("Unreachable")
	}
}
