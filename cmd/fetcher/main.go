/*
Copyright 2018 The Fission Authors.

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

package main

import (
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hanzoai/functions/cmd/fetcher/app"
	fv1 "github.com/hanzoai/functions/pkg/apis/core/v1"
	"github.com/hanzoai/functions/pkg/crd"
	"github.com/hanzoai/functions/pkg/utils/loggerfactory"
	"github.com/hanzoai/functions/pkg/utils/manager"
	"github.com/hanzoai/functions/pkg/utils/profile"
)

const fetcherPort = "8000"

// Usage: fetcher <shared volume path>
func main() {

	mgr := manager.New()
	defer mgr.Wait()

	logger := loggerfactory.GetLogger()
	defer logger.Sync()

	ctx := signals.SetupSignalHandler()
	profile.ProfileIfEnabled(ctx, logger, mgr)

	app.Run(ctx, crd.NewClientGenerator(), logger, mgr, fetcherPort, fv1.PodInfoMount)
}
