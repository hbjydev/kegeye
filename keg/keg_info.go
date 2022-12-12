package keg

import (
	"time"

	"gopkg.in/yaml.v3"
)

const (
	KegTimestampLayout = "2006-01-02 15:04:05Z"
)

type KegTimestamp struct {
	Time time.Time
}

func (kt *KegTimestamp) UnmarshalYAML(value *yaml.Node) error {
	t, err := time.Parse(KegTimestampLayout, value.Value)
	if err != nil {
		return err
	}

	kt.Time = t
	return nil
}

func (kt *KegTimestamp) MarshalYAML() (interface{}, error) {
  t := kt.Time.Format(KegTimestampLayout)
  return t, nil
}

type KegIndex struct {
	File    string  `yaml:"file" json:"file"`
	Summary *string `yaml:"summary" json:"summary"`
}

type KegInfo struct {
	Title   string       `yaml:"title" json:"title"`
	Kegv    *string      `yaml:"kegv" json:"kegv"`
	Creator *string      `yaml:"creator" json:"creator"`
	State   *string      `yaml:"state" json:"state"`
	Updated KegTimestamp `yaml:"updated" json:"updated"`

	Summary *string    `yaml:"summary" json:"summary"`
	Urls    []string   `yaml:"urls" json:"urls"`
	Indexes []KegIndex `yaml:"indexes" json:"indexes"`
}

func KegInfoFromYaml(in string) (KegInfo, error) {
	info := KegInfo{}

	inBytes := []byte(in)
	if err := yaml.Unmarshal(inBytes, &info); err != nil {
		return KegInfo{}, err
	}

	return info, nil
}
