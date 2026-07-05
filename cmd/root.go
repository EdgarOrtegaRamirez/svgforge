package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/svgforge/internal/convert"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/diff"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/models"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/optimizer"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/parser"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/query"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/stats"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/transform"
	"github.com/EdgarOrtegaRamirez/svgforge/internal/validate"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "svgforge",
	Short: "SVG Processing Toolkit",
	Long:  "A comprehensive SVG processing toolkit for parsing, optimizing, querying, diffing, converting, validating, and transforming SVG documents.",
}

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse and display SVG structure",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		printTree(doc, 0)
		return nil
	},
}

func printTree(doc *models.SVGDocument, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s<svg", prefix)
	if doc.Width != "" {
		fmt.Printf(" width=%q", doc.Width)
	}
	if doc.Height != "" {
		fmt.Printf(" height=%q", doc.Height)
	}
	if doc.ViewBox != nil {
		fmt.Printf(" viewBox=\"%.4g %.4g %.4g %.4g\"",
			doc.ViewBox.MinX, doc.ViewBox.MinY, doc.ViewBox.Width, doc.ViewBox.Height)
	}
	fmt.Println(">")
	if doc.Title != "" {
		fmt.Printf("%s  <title>%s</title>\n", prefix, doc.Title)
	}
	for _, el := range doc.Elements {
		printElement(el, indent+1)
	}
}

func printElement(el *models.Element, indent int) {
	if el == nil || el.Tag == "" {
		return
	}
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s<%s", prefix, el.Tag)
	for k, v := range el.Attributes {
		fmt.Printf(" %s=%q", k, v)
	}
	if el.Text != "" {
		fmt.Printf(">%s</%s>\n", el.Text, el.Tag)
	} else if len(el.Children) > 0 {
		fmt.Println(">")
		for _, child := range el.Children {
			printElement(child, indent+1)
		}
		fmt.Printf("%s</%s>\n", prefix, el.Tag)
	} else {
		fmt.Println(" />")
	}
}

var optimizeCmd = &cobra.Command{
	Use:   "optimize [file]",
	Short: "Optimize SVG by removing redundancy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		opts := optimizer.DefaultOptions()
		optimizer.Optimize(doc, opts)
		data, err := convert.ToBytes(doc)
		if err != nil {
			return fmt.Errorf("convert error: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:   "query [file] [selector]",
	Short: "Query SVG elements with CSS-like selectors",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		// For each element, check selector
		selector, err := query.Parse(args[1])
		if err != nil {
			return fmt.Errorf("selector parse error: %w", err)
		}
		for _, el := range doc.Elements {
			results := query.Query(el, selector)
			for _, r := range results {
				fmt.Printf("<%s", r.Tag)
				for k, v := range r.Attributes {
					fmt.Printf(" %s=%q", k, v)
				}
				fmt.Println(">")
			}
		}
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [file1] [file2]",
	Short: "Compare two SVG files structurally",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc1, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error on %s: %w", args[0], err)
		}
		doc2, err := p.ParseFile(args[1])
		if err != nil {
			return fmt.Errorf("parse error on %s: %w", args[1], err)
		}
		result := diff.Diff(doc1, doc2)
		fmt.Print(diff.FormatText(result))
		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [file]",
	Short: "Show SVG statistics",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		s := stats.Analyze(doc)
		fmt.Print(stats.FormatText(s))
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate SVG for issues",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		result := validate.Validate(doc)
		fmt.Print(validate.FormatText(result))
		return nil
	},
}

var convertCmd = &cobra.Command{
	Use:   "convert [file]",
	Short: "Convert SVG to other formats",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		to, _ := cmd.Flags().GetString("to")
		switch to {
		case "datauri":
			uri, err := convert.ToDataURI(doc)
			if err != nil {
				return err
			}
			fmt.Println(uri)
		case "html":
			html, err := convert.ToInlineHTML(doc)
			if err != nil {
				return err
			}
			fmt.Println(html)
		case "formatted":
			s, err := convert.ToFormatted(doc)
			if err != nil {
				return err
			}
			fmt.Print(s)
		default:
			data, err := convert.ToBytes(doc)
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
		return nil
	},
}

var transformCmd = &cobra.Command{
	Use:   "transform [file]",
	Short: "Apply transformations to SVG elements",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := parser.New()
		doc, err := p.ParseFile(args[0])
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		scale, _ := cmd.Flags().GetFloat64("scale")
		rotate, _ := cmd.Flags().GetFloat64("rotate")
		translateX, _ := cmd.Flags().GetFloat64("translate-x")
		translateY, _ := cmd.Flags().GetFloat64("translate-y")

		for _, el := range doc.Elements {
			if scale != 1.0 {
				transform.ScaleElement(el, scale, scale)
			}
			if rotate != 0 {
				transform.RotateElement(el, rotate)
			}
			if translateX != 0 || translateY != 0 {
				transform.TranslateElement(el, translateX, translateY)
			}
		}
		data, err := convert.ToBytes(doc)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	convertCmd.Flags().String("to", "bytes", "Output format: bytes, datauri, html, formatted")
	transformCmd.Flags().Float64("scale", 1.0, "Scale factor")
	transformCmd.Flags().Float64("rotate", 0, "Rotation angle in degrees")
	transformCmd.Flags().Float64("translate-x", 0, "X translation")
	transformCmd.Flags().Float64("translate-y", 0, "Y translation")

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(optimizeCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(transformCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
