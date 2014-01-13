package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

var (
	RequiredConfigVarKeys = []string{"name", "hostname", "port", "username"}
	UniqueConfigVarKey    = RequiredConfigVarKeys[0]
)

// Config holds a giver configuration's key/value pairs
type Config struct {
	Vars map[string]interface{}
}

// NewConfig returns a reference to an initialized Config
func NewConfig() *Config {
	return &Config{Vars: make(map[string]interface{})}
}

// uniqueValue returns the value for the Config's unique key
func (c *Config) uniqueValue() (value string) {
	return c.Vars[UniqueConfigVarKey].(string)
}

// Len returns the number of key/value pairs in the Config
func (c *Config) Len() int {
	return len(c.Vars)
}

// MarshalJSON satisfies the json.Marshaler interface
func (c *Config) MarshalJSON() (data []byte, err error) {
	data, err = json.Marshal(c.Vars)
	if err != nil {
		return
	}

	return
}

// UnmarshalJSON satisfies the json.Unmarshaler interface
func (c *Config) UnmarshalJSON(data []byte) (err error) {
	if err = json.Unmarshal(data, &c.Vars); err != nil {
		return
	}

	for _, required := range RequiredConfigVarKeys {
		if _, ok := c.Vars[required]; !ok {
			for k := range c.Vars {
				delete(c.Vars, k)
			}
			return fmt.Errorf("missing required field: %s", required)
		}
	}

	return
}

// ConfigList facilitates listing and sorting of Configs
type ConfigList struct {
	Configs []*Config `json:"configurations"`
	// hide orderBy to prevent it from being marshalled/unmarshalled
	orderBy []string
}

// NewConfigList returns a reference to an initialized ConfigList.
// The default order for sorting is given by RequiredConfigVarKeys.
func NewConfigList() *ConfigList {
	return &ConfigList{Configs: []*Config{}, orderBy: RequiredConfigVarKeys}
}

// uniqueConfig returns the Config with the given unique value and it's index within
// the list.  If a Config is not found, nil and -1 will be returned.
func (cl *ConfigList) uniqueConfig(uniqueVar string) (config *Config, index int) {
	index = -1
	for i, c := range cl.Configs {
		if c.uniqueValue() == uniqueVar {
			return c, i
		}
	}
	return
}

// add appends the given Config to the ConfigList only if there is no other
// Config with the same unique value
func (cl *ConfigList) add(config *Config) (success bool) {
	if _, i := cl.uniqueConfig(config.uniqueValue()); i == -1 {
		cl.Configs = append(cl.Configs, config)
		return true
	}
	return
}

// delete removes a Config from the list based upon the given unique value
func (cl *ConfigList) delete(uniqueVar string) (success bool) {
	if _, i := cl.uniqueConfig(uniqueVar); i >= 0 {
		copy(cl.Configs[i:], cl.Configs[i+1:])
		// careful of memory leaks
		cl.Configs[len(cl.Configs)-1] = nil
		cl.Configs = cl.Configs[:len(cl.Configs)-1]
		return true
	}
	return
}

// Add appends the provided Configs (in ConfigList form) and returns a
// ConfigList of Configs that could not be appended
func (cl *ConfigList) Add(configList *ConfigList) (failedList *ConfigList) {
	failedList = NewConfigList()
	for _, config := range configList.Configs {
		if ok := cl.add(config); !ok {
			failedList.add(config)
		}
	}
	return
}

// Delete removes the provided Configs (in ConfigList form) and returns a
// ConfigList of Configs that could not be removed
// Note: all that is really checked is the UniqueConfigVarKey, so you can pass
// in a mock list, but all the RequiredConfigVarKeys must be stubbed.
func (cl *ConfigList) Delete(configList *ConfigList) (failedList *ConfigList) {
	failedList = NewConfigList()
	for _, config := range configList.Configs {
		if ok := cl.delete(config.uniqueValue()); !ok {
			failedList.add(config)
		}
	}
	return
}

// OrderBy changes the sorting order of the ConfigList
func (cl *ConfigList) OrderBy(order []string) {
	cl.orderBy = append(order, RequiredConfigVarKeys...)
}

// Len satisfies sort.Interface
func (cl *ConfigList) Len() int {
	return len(cl.Configs)
}

// Less satisfies sort.Interface
func (cl *ConfigList) Less(i, j int) bool {
	for _, key := range cl.orderBy {
		cI, cJ := cl.Configs[i].Vars[key], cl.Configs[j].Vars[key]

		if cI != cJ {

			// the required fields must be either strings, float64s or bools for now,
			// remember that JSON uses floats for all numbers, never ints
			if t, ok := cI.(string); ok {
				return t < cJ.(string)
			} else if t, ok := cI.(float64); ok {
				return t < cJ.(float64)
			} else {
				// false comes before true
				return !cI.(bool)
			}

		}

	}
	return false
}

// Swap satisfies sort.Interface
func (cl *ConfigList) Swap(i, j int) {
	cl.Configs[i], cl.Configs[j] = cl.Configs[j], cl.Configs[i]
}

// String returns a formatted and indented JSON string of the current ConfigList
func (cl *ConfigList) String() string {
	sort.Sort(cl)
	b, _ := json.MarshalIndent(cl, "", "    ")
	return string(b)
}
