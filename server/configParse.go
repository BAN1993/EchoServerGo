package main

import (
	"fmt"
	"github.com/ini"
	"strings"
)

type Configs struct {
	fd          *ini.File
	listenaddr  string
	loglevel    uint32
	logfilename string
}

func (c* Configs) Load(path string) error {
	conf, err := ini.Load("ServerConfig.ini")
	if err != nil {
		return err
	}
	c.fd = conf

	listen, err := c.getString("server", "listen")
	if err != nil {
		return err
	}
	c.listenaddr = listen

	level, err := c.getUInt32("log", "level")
	if err != nil {
		return err
	}
	c.loglevel = level

	filename, err := c.getString("log", "filename")
	if err != nil {
		return err
	}
	c.logfilename = filename

	return nil
}

func (c* Configs) getString(section string, key string) (string, error) {
	s := c.fd.Section(section)
	if s == nil {
		return "", fmt.Errorf("can not get section:%s", section)
	}

	d := s.Key(key).String()
	if len(d) <= 0 {
		return "", fmt.Errorf("can not get key:%s", key)
	}

	return d, nil
}

func (c* Configs) getUInt32(section string, key string) (uint32, error) {
	s := c.fd.Section(section)
	if s == nil {
		return 0, fmt.Errorf("can not get section:%s", section)
	}

	d, err := s.Key(key).Int()
	if err != nil {
		return 0, fmt.Errorf("can not get key:%s:%s", key, err)
	}

	return uint32(d), nil
}

func (c* Configs) getBool(section string, key string) (bool, error) {
	s := c.fd.Section(section)
	if s == nil {
		return false, fmt.Errorf("can not get section:%s", section)
	}

	d, err := s.Key(key).Bool()
	if err != nil {
		return false, fmt.Errorf("can not get key:%s:%s", key, err)
	}

	return d, nil
}

func PreString(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\r", "", -1)
	return str
}
