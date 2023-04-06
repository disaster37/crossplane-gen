package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/disaster37/crossplane-crd-generator/helper"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCrd(t *testing.T) {

	log.SetLevel(log.DebugLevel)

	var (
		err          error
		data         []byte
		expectedData []byte
		diff         string
	)

	tmpPath, err := ioutil.TempDir("", "test")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(tmpPath)

	// normal case
	err = generateCRD("./testdata/...", tmpPath, GenerateCrdOption{})
	assert.NoError(t, err)
	assert.FileExists(t, fmt.Sprintf("%s/test.crossplane.io_xtests.yaml", tmpPath))

	data, err = os.ReadFile(fmt.Sprintf("%s/test.crossplane.io_xtests.yaml", tmpPath))
	if err != nil {
		t.Fatal(err)
	}
	expectedData, err = os.ReadFile("testdata/default_xrd.yaml")
	if err != nil {
		t.Fatal(err)
	}
	diff = helper.Diff(string(expectedData), string(data))
	if diff != "" {
		assert.Fail(t, diff)
	}

	// With options
	err = generateCRD("./testdata/...", tmpPath, GenerateCrdOption{
		ClaimName:       "test",
		ClaimNamePlural: "tests",
		CrdOptions: []string{
			"generateEmbeddedObjectMeta=true",
		},
	})
	assert.NoError(t, err)
	assert.FileExists(t, fmt.Sprintf("%s/test.crossplane.io_xtests.yaml", tmpPath))

	data, err = os.ReadFile(fmt.Sprintf("%s/test.crossplane.io_xtests.yaml", tmpPath))
	if err != nil {
		t.Fatal(err)
	}
	expectedData, err = os.ReadFile("testdata/xrd_with_options.yaml")
	if err != nil {
		t.Fatal(err)
	}
	diff = helper.Diff(string(expectedData), string(data))
	if diff != "" {
		assert.Fail(t, diff)
	}

}
