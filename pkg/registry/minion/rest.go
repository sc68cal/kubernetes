/*
Copyright 2014 Google Inc. All rights reserved.

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

package minion

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
)

// REST implements the RESTStorage interface, backed by a MinionRegistry.
type REST struct {
	registry Registry
}

// NewREST returns a new REST.
func NewREST(m Registry) *REST {
	return &REST{
		registry: m,
	}
}

var ErrDoesNotExist = fmt.Errorf("The requested resource does not exist.")
var ErrNotHealty = fmt.Errorf("The requested minion is not healthy.")

func (rs *REST) Create(ctx api.Context, obj runtime.Object) (<-chan runtime.Object, error) {
	minion, ok := obj.(*api.Minion)
	if !ok {
		return nil, fmt.Errorf("not a minion: %#v", obj)
	}
	if minion.ID == "" {
		return nil, fmt.Errorf("ID should not be empty: %#v", minion)
	}

	minion.CreationTimestamp = util.Now()

	return apiserver.MakeAsync(func() (runtime.Object, error) {
		err := rs.registry.CreateMinion(ctx, minion)
		if err != nil {
			return nil, err
		}
		minion, err := rs.registry.GetMinion(ctx, minion.ID)
		if minion == nil {
			return nil, ErrDoesNotExist
		}
		if err != nil {
			return nil, err
		}
		return minion, nil
	}), nil
}

func (rs *REST) Delete(ctx api.Context, id string) (<-chan runtime.Object, error) {
	minion, err := rs.registry.GetMinion(ctx, id)
	if minion == nil {
		return nil, ErrDoesNotExist
	}
	if err != nil {
		return nil, err
	}
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		return &api.Status{Status: api.StatusSuccess}, rs.registry.DeleteMinion(ctx, id)
	}), nil
}

func (rs *REST) Get(ctx api.Context, id string) (runtime.Object, error) {
	minion, err := rs.registry.GetMinion(ctx, id)
	if minion == nil {
		return nil, ErrDoesNotExist
	}
	return minion, err
}

func (rs *REST) List(ctx api.Context, label, field labels.Selector) (runtime.Object, error) {
	return rs.registry.ListMinions(ctx)
}

func (rs *REST) New() runtime.Object {
	return &api.Minion{}
}

func (rs *REST) Update(ctx api.Context, minion runtime.Object) (<-chan runtime.Object, error) {
	return nil, fmt.Errorf("Minions can only be created (inserted) and deleted.")
}

func (rs *REST) toApiMinion(name string) *api.Minion {
	return &api.Minion{TypeMeta: api.TypeMeta{ID: name}}
}
