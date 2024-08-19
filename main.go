package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	jsonParser "github.com/hashicorp/hcl/v2/json"
	"github.com/spf13/pflag"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func main() {
	reverse := pflag.BoolP("reverse", "r", false, "input JSON, output TFVARS")
	pflag.Parse()

	var err error
	if *reverse {
		err = toTFVARS()
	} else {
		err = toJSON()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func toJSON() error {
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read from stdin: %s", err)
	}

	ast, diags := hclsyntax.ParseConfig(src, "file.tfvars", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("unable to parse TFVARS: %s", diags.Error())
	}
	if ast.Body == nil {
		fmt.Println("{}")
		os.Exit(0)
	}

	attrs, diags := ast.Body.JustAttributes()
	if diags.HasErrors() {
		return fmt.Errorf("error evaluating: %s", diags.Error())
	}

	values := make(map[string]json.RawMessage, len(attrs))
	for name, attr := range attrs {
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("error evaluating %q: %s", name, diags.Error())
		}
		buf, err := ctyjson.Marshal(value, value.Type())
		if err != nil {
			return fmt.Errorf("error converting %q to JSON: %s", name, err)
		}
		values[name] = json.RawMessage(buf)
	}

	output, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %s", err)
	}
	fmt.Println(string(output))

	return nil
}

func toTFVARS() error {
	src, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to read from stdin: %s", err)
	}

	ast, diags := jsonParser.Parse([]byte(src), "file.json")
	if diags.HasErrors() {
		return fmt.Errorf("unable to parse JSON: %s", diags.Error())
	}

	attrs, diags := ast.Body.JustAttributes()
	if diags.HasErrors() {
		return fmt.Errorf("error evaluating: %s", diags.Error())
	}

	outF := hclwrite.NewEmptyFile()
	rootBody := outF.Body()
	for name, attr := range attrs {
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("error evaluating %q: %s", name, diags.Error())
		}
		rootBody.SetAttributeValue(name, value)
	}

	fmt.Fprintf(os.Stdout, "%s", hclwrite.Format(outF.Bytes()))

	return nil
}
