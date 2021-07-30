package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Configuration struct {
	Port            string `json:"port"`
	MaxReadBytes    int64  `json:"maxReadBytes"`
	DeadlineSeconds int    `json:"deadlineSeconds"`
}

func (c *Configuration) Init() {
	file := GetProgramPath() + ".json"

	ok, err := Exists(file)
	ChkErrFatal(err)

	if ok {
		bf, e := ioutil.ReadFile(file)
		ChkErrFatal(e)
		ChkErrFatal(json.Unmarshal(bf, c))
	} else {
		c.setStandartConfig()
		b, e := json.Marshal(c)
		ChkErrFatal(e)
		ChkErrFatal(ioutil.WriteFile(file, b, os.ModePerm))
	}
}

func (c *Configuration) setStandartConfig() {
	c.Port = "65534"
	c.DeadlineSeconds = 180
	c.MaxReadBytes = 2048
}
