package cmd

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func writeDataIntoFileIfChanged(filename string, data interface{}) error {
	err := createFileIfNotExist(filename, []byte(`[]`))
	if err != nil {
		return errors.WithStack(err)
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.WithStack(err)
	}

	h := md5.New()
	fileHash := h.Sum(buf)
	h.Reset()

	buf, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	dataHash := h.Sum(buf)
	if bytes.Equal(fileHash, dataHash) {
		logrus.Debugf("has no change")
		return nil
	}

	logrus.Infof("service discovery configuration saved")
	err = ioutil.WriteFile(filename, buf, 0)
	return errors.WithStack(err)
}

func createFileIfNotExist(filename string, data []byte) error {
	if _, err := os.Stat(filename); os.IsExist(err) {
		return nil
	}

	logrus.Debugf("create file %s with: %s", filename, data)
	err := ioutil.WriteFile(filename, data, os.ModePerm)
	return errors.WithStack(err)
}
