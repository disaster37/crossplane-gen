package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/rbac"
	"sigs.k8s.io/controller-tools/pkg/schemapatcher"
	"sigs.k8s.io/controller-tools/pkg/webhook"
	"sigs.k8s.io/yaml"
)

var (
	optionsRegistry = &markers.Registry{}
)

type GenerateCrdOption struct {
	CrdOptions         []string
	SchemaPatchOptions []string
	ClaimName          string
	ClaimNamePlural    string
}

func init() {

	allGenerators := map[string]genall.Generator{
		"crd":         crd.Generator{},
		"rbac":        rbac.Generator{},
		"object":      deepcopy.Generator{},
		"webhook":     webhook.Generator{},
		"schemapatch": schemapatcher.Generator{},
	}

	// allOutputRules defines the list of all known output rules, giving
	// them names for use on the command line.
	// Each output rule turns into two command line options:
	// - output:<generator>:<form> (per-generator output)
	// - output:<form> (default output)
	allOutputRules := map[string]genall.OutputRule{
		"dir":       genall.OutputToDirectory(""),
		"none":      genall.OutputToNothing,
		"stdout":    genall.OutputToStdout,
		"artifacts": genall.OutputArtifacts{},
	}

	for genName, gen := range allGenerators {
		// make the generator options marker itself
		defn := markers.Must(markers.MakeDefinition(genName, markers.DescribesPackage, gen))
		if err := optionsRegistry.Register(defn); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := gen.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(defn, help)
			}
		}

		// make per-generation output rule markers
		for ruleName, rule := range allOutputRules {
			ruleMarker := markers.Must(markers.MakeDefinition(fmt.Sprintf("output:%s:%s", genName, ruleName), markers.DescribesPackage, rule))
			if err := optionsRegistry.Register(ruleMarker); err != nil {
				panic(err)
			}
			if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
				if help := helpGiver.Help(); help != nil {
					optionsRegistry.AddHelp(ruleMarker, help)
				}
			}
		}
	}

	// make "default output" output rule markers
	for ruleName, rule := range allOutputRules {
		ruleMarker := markers.Must(markers.MakeDefinition("output:"+ruleName, markers.DescribesPackage, rule))
		if err := optionsRegistry.Register(ruleMarker); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(ruleMarker, help)
			}
		}
	}

	// add in the common options markers
	if err := genall.RegisterOptionsMarkers(optionsRegistry); err != nil {
		panic(err)
	}

}

func GenerateCRD(c *cli.Context) error {

	options := GenerateCrdOption{
		CrdOptions:         c.StringSlice("crd-options"),
		SchemaPatchOptions: c.StringSlice("schemapatch-options"),
		ClaimName:          c.String("claim-name"),
		ClaimNamePlural:    c.String("claim-plural-name"),
	}

	return generateCRD(c.String("source-path"), c.String("target-path"), options)

}

func generateCRD(sourcePath string, targetPath string, options GenerateCrdOption) (err error) {

	// Create temporary folder to generate initial CRD with gen-controller
	tmpPath, err := ioutil.TempDir("", "crossplaneG")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(tmpPath)

	opts := []string{
		fmt.Sprintf("paths=\"%s\"", sourcePath),
		fmt.Sprintf("output:crd:artifacts:config=\"%s\"", tmpPath),
	}

	if len(options.CrdOptions) > 0 {
		for _, crdOption := range options.CrdOptions {
			opts = append(opts, fmt.Sprintf("crd:%s", crdOption))
		}
	} else {
		opts = append(opts, "crd")
	}

	for _, schemaPatchOption := range options.SchemaPatchOptions {
		opts = append(opts, fmt.Sprintf("schemapatch:%s", schemaPatchOption))
	}

	log.Debugf("Options: %s", spew.Sdump(opts))

	rt, err := genall.FromOptions(optionsRegistry, opts)
	if err != nil {
		return errors.Wrap(err, "Error when init controller-gen with options that you provide")
	}
	if len(rt.Generators) == 0 {
		return fmt.Errorf("no generators specified")
	}

	if hadErrs := rt.Run(); hadErrs {
		// don't obscure the actual error with a bunch of usage
		return errors.New("not all generators ran successfully")
	}

	log.Debug("CRD generated with controller-gen")

	// Loop over all generated CRD to clean them
	filepath.Walk(tmpPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "Error when open file %s", info.Name())
		}

		if info.IsDir() {
			return nil
		}

		log.Debugf("Process file %s", info.Name())

		// Read and convert file to map
		f, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Error when read file %s", path)
		}
		obj := map[string]any{}
		if err = yaml.Unmarshal(f, &obj); err != nil {
			return errors.Wrapf(err, "Error when convert file %s to map", path)
		}

		// Clean map and write final file
		obj["apiVersion"] = "apiextensions.crossplane.io/v1"
		obj["kind"] = "CompositeResourceDefinition"
		delete(obj["metadata"].(map[string]any), "creationTimestamp")
		delete(obj["metadata"].(map[string]any), "annotations")
		delete(obj["spec"].(map[string]any), "scope")

		spec := obj["spec"].(map[string]any)

		if options.ClaimName != "" {
			m := map[string]any{
				"kind": options.ClaimName,
			}
			if options.ClaimNamePlural != "" {
				m["plural"] = options.ClaimNamePlural
			}
			spec["claimNames"] = m
		}

		for i, version := range spec["versions"].([]any) {
			version := version.(map[string]any)
			delete(version, "storage")
			version["referenceable"] = true

			p := version["schema"].(map[string]any)["openAPIV3Schema"].(map[string]any)["properties"].(map[string]any)
			delete(p, "apiVersion")
			delete(p, "kind")
			delete(p, "metadata")

			spec["versions"].([]any)[i] = version

		}

		// Write final file
		data, err := yaml.Marshal(obj)
		if err != nil {
			return errors.Wrap(err, "Error when convert object to Yaml")
		}
		os.WriteFile(fmt.Sprintf("%s/%s", targetPath, info.Name()), data, 0644)

		return nil
	})

	return nil
}
