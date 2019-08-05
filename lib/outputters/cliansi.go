package outputters

import (
	"fmt"
	"github.com/aybabtme/rgbterm"
	"sort"
)

type ansiCLI struct {
	Output
}

func NewAnsiCLI() *ansiCLI {
	cli := new(ansiCLI)
	return cli
}

// Tree outputs sorted hashmap to the CLI with ANSI escapes
func (cli *ansiCLI) Tree(tree map[string][]string) {
	rootLabelIndex := []string{}
	for label := range tree {
		rootLabelIndex = append(rootLabelIndex, label)
	}
	sort.Strings(rootLabelIndex)

	fmt.Printf("%s\n", "\u2514\u2500\u2510")
	var branch string
	if len(rootLabelIndex) > 1 {
		branch = "\u251c"
	} else {
		branch = "\u2514"
	}
	branchSingle := branch + "\u2500\u2500"
	branchSingleEnd := "\u2514\u2500\u2500"

	for idx, label := range rootLabelIndex {
		idx++
		childLabels, exists := tree[label]
		var rootBranch string
		if idx < len(rootLabelIndex) {
			rootBranch = branchSingle
		} else {
			rootBranch = branchSingleEnd
		}
		cIdx := rgbterm.FgString(fmt.Sprintf("(%02d)", idx), 0xff, 0xff, 0) // Index of the root channel
		cLabel := rgbterm.FgString(label, 0xff, 0xff, 0xff)
		fmt.Printf("  %s%s %s\n", rootBranch, cIdx, cLabel)

		if exists && len(childLabels) > 0 {
			if idx < len(rootLabelIndex) {
				fmt.Printf("  %s ", "\u2502")
			} else {
				fmt.Printf("    ")
			}
			sort.Strings(childLabels)
			for cidx, childLabel := range childLabels {
				if cidx == 0 {
					if len(childLabels) == 1 {
						fmt.Printf("      %s %s\n", branchSingleEnd, childLabel)
					} else {
						fmt.Printf("      %s %s\n", branchSingle, childLabel)
					}
				} else if cidx < len(childLabels)-1 {
					fmt.Printf("  %s       %s %s\n", "\u2502", branchSingle, childLabel)
				} else {
					fmt.Printf("  %s       %s %s\n", "\u2502", branchSingleEnd, childLabel)
				}
			}
		}
		if idx < len(rootLabelIndex) {
			fmt.Printf("  %s\n", "\u2502")
		} else {
			fmt.Println("")
		}
	}
}
