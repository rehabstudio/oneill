package dockerclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Env represents a list of key-pair represented in the form KEY=VALUE.
type Env []string

// UnmarshalYAML converts the YAML map into a slice of strings containing `k=v`
// pairs.
func (m *Env) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	if m == nil {
		return errors.New("Env: UnmarshalYAML on nil pointer")
	}

	// unmarshall yaml data to struct
	interfaceMap := make(map[string]interface{})
	err := unmarshal(&interfaceMap)
	if err != nil {
		return err
	}

	// jsonify all map values so we can treat them as raw strings instead of
	// rich data structures
	stringMap := make(map[string]string)
	for k, v := range interfaceMap {
		switch v := v.(type) {
		case string:
			stringMap[k] = v
		default:
			_tmp, err := json.Marshal(v)
			if err != nil {
				return err
			}
			stringMap[k] = string(_tmp)
		}
	}

	// convert newly created struct into a simple slice of k=v environment
	// variable pairs
	*m = envMapToSlice(stringMap)

	return nil
}

// envSliceToMap returns the map representation of a slice of strings
// containing environment variables.
func envSliceToMap(env []string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, kv := range env {
		parts := strings.SplitN(kv, "=", 2)
		m[parts[0]] = parts[1]
	}
	return m
}

// envMapToslice flattens a map[string]string into a flat slice of strings
// where keys are seperated from values by `=`.
func envMapToSlice(envMap map[string]string) []string {
	var envSlice []string
	for k, v := range envMap {
		envString := fmt.Sprintf("%s=%s", k, v)
		envSlice = append(envSlice, envString)
	}
	return envSlice
}

// mergeEnvs merges two slices of strings containing environment variables.
// All values in newEnv will be appended to origEnv, except if a value with
// the given key already exists, in which case the value in origEnv will be
// overwritten.
func mergeEnvs(origEnv, newEnv []string) []string {

	// convert both slices into maps to make them easier to work with
	origEnvMap := envSliceToMap(origEnv)
	newEnvMap := envSliceToMap(newEnv)

	// merge newEnv into origEnv, overwriting keys as necessary
	for k, v := range newEnvMap {
		origEnvMap[k] = v
	}

	// convert map back into a slice and return it
	return envMapToSlice(origEnvMap)
}

// EnvsMatch checks if a running container's environment matches the one
// defined in a container definition. The variables defined in the container
// definition are added to those defined in the base image before comparing
// with those read from the running container.
func EnvsMatch(env0, env1, fromImage []string) bool {

	env0 = mergeEnvs(fromImage, env0)

	if len(env0) != len(env1) {
		return false
	}

	sort.Strings(env0)
	sort.Strings(env1)

	for i := range env0 {
		if env0[i] != env1[i] {
			return false
		}
	}

	return true
}
