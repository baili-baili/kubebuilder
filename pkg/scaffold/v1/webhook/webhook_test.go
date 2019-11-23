/*
Copyright 2020 The Kubernetes Authors.

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

package webhook_test

import (
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kubebuilder/pkg/model"
	"sigs.k8s.io/kubebuilder/pkg/model/config"
	"sigs.k8s.io/kubebuilder/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/scaffoldtest"
	. "sigs.k8s.io/kubebuilder/pkg/scaffold/v1/webhook"
)

var _ = Describe("Webhook", func() {
	type webhookTestcase struct {
		resource.Options
		Config
	}

	serverName := "default"
	inputs := []*webhookTestcase{
		{
			Options: resource.Options{
				Group:                      "crew",
				Version:                    "v1",
				Kind:                       "FirstMate",
				Namespaced:                 true,
				CreateExampleReconcileBody: true,
			},
			Config: Config{
				Type:       "mutating",
				Operations: []string{"create", "update"},
				Server:     serverName,
			},
		},
		{
			Options: resource.Options{
				Group:                      "crew",
				Version:                    "v1",
				Kind:                       "FirstMate",
				Namespaced:                 true,
				CreateExampleReconcileBody: true,
			},
			Config: Config{
				Type:       "mutating",
				Operations: []string{"delete"},
				Server:     serverName,
			},
		},
		{
			Options: resource.Options{
				Group:                      "ship",
				Version:                    "v1beta1",
				Kind:                       "Frigate",
				Namespaced:                 true,
				CreateExampleReconcileBody: false,
			},
			Config: Config{
				Type:       "validating",
				Operations: []string{"update"},
				Server:     serverName,
			},
		},
		{
			Options: resource.Options{
				Group:                      "creatures",
				Version:                    "v2alpha1",
				Kind:                       "Kraken",
				Namespaced:                 false,
				CreateExampleReconcileBody: false,
			},
			Config: Config{
				Type:       "validating",
				Operations: []string{"create"},
				Server:     serverName,
			},
		},
		{
			Options: resource.Options{
				Group:                      "core",
				Version:                    "v1",
				Kind:                       "Namespace",
				Namespaced:                 false,
				CreateExampleReconcileBody: false,
			},
			Config: Config{
				Type:       "mutating",
				Operations: []string{"update"},
				Server:     serverName,
			},
		},
	}

	for i := range inputs {
		in := inputs[i]
		res := in.Options.NewV1Resource(
			&config.Config{
				Version: config.Version1,
				Domain:  "testproject.org",
				Repo:    "project",
			},
			false,
		)

		Describe(fmt.Sprintf("scaffolding webhook %s", in.Kind), func() {
			files := []struct {
				instance input.File
				file     string
			}{
				{
					file: filepath.Join("pkg", "webhook", "add_default_server.go"),
					instance: &AddServer{
						Config: in.Config,
					},
				},
				{
					file: filepath.Join("pkg", "webhook", "default_server", "server.go"),
					instance: &Server{
						Config: in.Config,
					},
				},
				{
					file: filepath.Join("pkg", "webhook", "default_server",
						fmt.Sprintf("add_%s_%s.go", strings.ToLower(in.Type), strings.ToLower(in.Kind))),
					instance: &AddAdmissionWebhookBuilderHandler{
						Resource: res,
						Config:   in.Config,
					},
				},
				{
					file: filepath.Join("pkg", "webhook", "default_server",
						strings.ToLower(in.Kind), strings.ToLower(in.Type),
						"webhooks.go"),
					instance: &AdmissionWebhooks{
						Resource: res,
						Config:   in.Config,
					},
				},
				{
					file: filepath.Join("pkg", "webhook", "default_server",
						strings.ToLower(in.Kind), strings.ToLower(in.Type),
						fmt.Sprintf("%s_webhook.go", strings.Join(in.Operations, "_"))),
					instance: &AdmissionWebhookBuilder{
						Resource: res,
						Config:   in.Config,
					},
				},
				{
					file: filepath.Join("pkg", "webhook", "default_server",
						strings.ToLower(in.Kind), strings.ToLower(in.Type),
						fmt.Sprintf("%s_%s_handler.go", strings.ToLower(in.Kind), strings.Join(in.Operations, "_"))),
					instance: &AdmissionHandler{
						Resource: res,
						Config:   in.Config,
					},
				},
			}

			for j := range files {
				f := files[j]
				Context(f.file, func() {
					It(fmt.Sprintf("should write a file matching the golden file %s", f.file), func() {
						s, result := scaffoldtest.NewTestScaffold(f.file, f.file)
						Expect(s.Execute(&model.Universe{}, scaffoldtest.Options(), f.instance)).To(Succeed())
						Expect(result.Actual.String()).To(Equal(result.Golden), result.Actual.String())
					})
				})
			}
		})
	}
})
