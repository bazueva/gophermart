package config

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerAddr struct {
	Host string
	Port int
}

func (s *ServerAddr) UnmarshalText(text []byte) error {
	return s.Set(string(text))
}

func (s *ServerAddr) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *ServerAddr) Set(addr string) error {
	value := strings.Split(addr, ":")
	if len(value) != 2 {
		return fmt.Errorf("wrong format")
	}

	var err error
	s.Port, err = strconv.Atoi(value[1])
	if err != nil {
		return fmt.Errorf("wrong port - %s", err.Error())
	}

	if value[0] != "" {
		s.Host = value[0]
	}

	return nil
}
