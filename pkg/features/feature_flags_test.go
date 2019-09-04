/*
Copyright 2019 the Velero contributors.

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

package features

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureFlags(t *testing.T) {
	newSet := NewFeatureFlagSet("feature1", "feature2")

	assert.True(t, newSet.Enabled("feature1"))
	assert.False(t, newSet.Enabled("feature3"))
	assert.Equal(t, []string{"feature1", "feature2"}, newSet.All())

	newSet.Enable("feature3")
	assert.True(t, newSet.Enabled("feature3"))
}
