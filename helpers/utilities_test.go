// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"anonymous-messaging/config"

	"github.com/stretchr/testify/assert"

	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
)

var mixes []config.MixConfig
var testDir string

// ById implements the sort interface and sorts based on the id of the nodes
type ById []config.MixConfig

func (v ById) Len() int           { return len(v) }
func (v ById) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ById) Less(i, j int) bool { return v[i].Id < v[j].Id }

func Setup() error {
	for i := 0; i < 10; i++ {
		mixes = append(mixes, config.MixConfig{Id: fmt.Sprintf("Mix%d", i),
			Host:   fmt.Sprintf("Host%d", i),
			Port:   fmt.Sprintf("Port%d", i),
			PubKey: nil})
	}

	currDir, err := os.Getwd()
	if err != nil {
		return err
	}
	testDir = currDir + "/test_path"
	return nil
}

func Clean() error {
	err := os.RemoveAll(testDir)
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	err := Setup()
	if err != nil {
		panic(m)
	}

	code := m.Run()

	err = Clean()
	if err != nil {
		panic(m)
	}
	os.Exit(code)
}

func TestDirExists_Pass(t *testing.T) {

	err := os.Mkdir(testDir, 755)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := DirExists(testDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, exists, " DirExists should return false for a non existing directory")
}

func TestDirExists_Fail(t *testing.T) {
	exists, err := DirExists("completely_random_directory/xxx")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, exists, " DirExists should return false for a non existing directory")
}

func TestPermute_Pass(t *testing.T) {
	permuted, err := Permute(mixes)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(mixes), len(permuted), " Permute should return a permutation of a given slice, hence the lengths should be equal")
	sort.Sort(ById(mixes))
	sort.Sort(ById(permuted))
	assert.True(t, reflect.DeepEqual(mixes, permuted))

}

func TestPermute_Fail(t *testing.T) {
	_, err := Permute([]config.MixConfig{})
	assert.EqualError(t, errors.New(" cannot permute an empty list of mixes"), err.Error(), " Permute should return an error for an empty slice")
}

func TestRandomExponential_Pass(t *testing.T) {
	val, err := RandomExponential(5.0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, reflect.Float64, reflect.TypeOf(val).Kind(), " RandomExponential should return a single float64 value")
}

func TestRandomExponential_Fail_ZeroParam(t *testing.T) {
	_, err := RandomExponential(0.0)
	assert.EqualError(t, errors.New("the parameter of exponential distribution has to be larger than zero"), err.Error(), " RandomExponential should return an error if the given parameter is non-positive")

}

func TestRandomExponential_Fail_NegativeParam(t *testing.T) {
	_, err := RandomExponential(-1.0)
	assert.EqualError(t, errors.New("the parameter of exponential distribution has to be larger than zero"), err.Error(), " RandomExponential should return an error if the given parameter is non-positive")
}

func TestRandomSample_Pass_SmallerLen(t *testing.T) {
	sample, err := RandomSample(mixes, 5)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 5, len(sample), " RandomSample should return a sample of given size")
}

func TestRandomSample_Pass_EqualLen(t *testing.T) {
	sample, err := RandomSample(mixes, 5)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 5, len(sample), " RandomSample should return a sample of given size")
}

func TestRandomSample_Fail(t *testing.T) {
	_, err := RandomSample(mixes, 20)
	assert.EqualError(t, errors.New(" cannot take a sample larger than the given list"), err.Error(), " RandomSample cannot take a sample larger than the given slice")
}

func TestResolveTCPAddress(t *testing.T) {
	// TO DO: How this should be tested ? And should it even be tested it if it uses a build in function?
}
