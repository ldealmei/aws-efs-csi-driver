/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"k8s.io/kubernetes/test/e2e/framework"
	frameworkconfig "k8s.io/kubernetes/test/e2e/framework/config"
	"k8s.io/kubernetes/test/e2e/framework/testfiles"
)

const (
	kubeconfigEnvVar = "KUBECONFIG"

	commaDelim             = ","
	equalDelim             = "="
	expectedSelectorTokens = 2
)

// Combined label selectors in the form of key1=value1,key2=value2.
// Supplied by users.
var combinedEfsDriverLabelSelectors string

func init() {
	testing.Init()
	// k8s.io/kubernetes/test/e2e/framework requires env KUBECONFIG to be set
	// it does not fall back to defaults
	if os.Getenv(kubeconfigEnvVar) == "" {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		os.Setenv(kubeconfigEnvVar, kubeconfig)
	}

	framework.AfterReadingAllFlags(&framework.TestContext)
	// PWD is test/e2e inside the git repo
	testfiles.AddFileSource(testfiles.RootFileSource{Root: "../.."})

	frameworkconfig.CopyFlags(frameworkconfig.Flags, flag.CommandLine)
	framework.RegisterCommonFlags(flag.CommandLine)
	framework.RegisterClusterFlags(flag.CommandLine)

	flag.StringVar(&ClusterName, "cluster-name", "", "the cluster name")
	flag.StringVar(&Region, "region", "us-west-2", "the region")
	flag.StringVar(&FileSystemId, "file-system-id", "", "the id of an existing file system")
	flag.StringVar(&EfsDriverNamespace, "efs-driver-namespace", "kube-system", "namespace of EFS driver pods")
	flag.StringVar(&combinedEfsDriverLabelSelectors, "efs-driver-label-selectors", "app=efs-csi-node", "comma-separated label selectors for EFS driver pods, follows the form key1=value1,key2=value2")

	flag.Parse()

	var err error
	EfsDriverLabelSelectors, err = parseCombinedEfsDriverLabelSelectors()
	if err != nil {
		log.Fatalln(err)
	}
}

func parseCombinedEfsDriverLabelSelectors() (map[string]string, error) {
	selectors := map[string]string{}
	for _, combinedTokens := range strings.Split(combinedEfsDriverLabelSelectors, commaDelim) {
		selectorTokens := strings.Split(combinedTokens, equalDelim)
		if len(selectorTokens) != expectedSelectorTokens {
			return nil, errors.New("failed to parse combined EFS driver label selectors")
		}
		selectors[selectorTokens[0]] = selectorTokens[1]
	}
	return selectors, nil
}

func TestEFSCSI(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Run tests through the Ginkgo runner with output to console + JUnit for Jenkins
	var r []ginkgo.Reporter
	if framework.TestContext.ReportDir != "" {
		if err := os.MkdirAll(framework.TestContext.ReportDir, 0755); err != nil {
			log.Fatalf("Failed creating report directory: %v", err)
		} else {
			r = append(r, reporters.NewJUnitReporter(path.Join(framework.TestContext.ReportDir, fmt.Sprintf("junit_%v%02d.xml", framework.TestContext.ReportPrefix, config.GinkgoConfig.ParallelNode))))
		}
	}
	log.Printf("Starting e2e run %q on Ginkgo node %d", framework.RunID, config.GinkgoConfig.ParallelNode)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "EFS CSI Suite", r)
}
