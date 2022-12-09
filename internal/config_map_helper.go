package internal

import (
	"sort"
)

const (
	EKuiperToolConfig string = "ekuiper-tool-config"
)

type ConfigMapInfo struct {
	// pod.volumeMount.name
	MountName string
	// pod.volumeMount.mountPath
	MountPath string
	// spec.volume.name the full config map name is {hdb.name}-suffix
	MapNameSuffix string
	MapKey        string
	MapPath       string
}

type ConfigmapSet struct {
	cms map[string]*ConfigMapInfo
}

var ConfigMaps = ConfigmapSet{
	cms: map[string]*ConfigMapInfo{
		EKuiperToolConfig: {
			MountName:     EKuiperToolConfig,
			MountPath:     "/kuiper-kubernetes-tool/sample",
			MapNameSuffix: "ekuiper-tool-config",
			MapKey:        "neuronStream.json",
		},
	},
}

// Visit visits the config map in lexicographical order, calling fn for each.
func (c *ConfigmapSet) Visit(fn func(m ConfigMapInfo)) {
	for _, flag := range sortConfigMaps(c.cms) {
		fn(*c.cms[flag])
	}
}

// Get returns the config map of given name
func (c *ConfigmapSet) Get(name string) (ConfigMapInfo, bool) {
	if m, ok := c.cms[name]; ok {
		return *m, true
	}
	return ConfigMapInfo{}, false
}

// sortConfigMaps returns the flags as a slice in lexicographical sorted order.
func sortConfigMaps(cms map[string]*ConfigMapInfo) []string {
	result := make([]string, len(cms))
	i := 0
	for name := range cms {
		result[i] = name
		i++
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func GetEKuiperToolConfig() map[string]any {
	return map[string]any{
		"command": map[string]interface{}{
			"url":         "/streams",
			"description": "create neuronStream",
			"method":      "post",
			"data": map[string]string{
				"sql": `create stream neuronStream() WITH (TYPE="neuron",FORMAT="json",SHARED="true");`,
			},
		},
	}
}
