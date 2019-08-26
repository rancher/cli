package main

import (
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	check.TestingT(t)
}

type MainTestSuite struct {
}

var _ = check.Suite(&MainTestSuite{})

func (m *MainTestSuite) SetUpSuite(c *check.C) {
}

func (m *MainTestSuite) TestParseArgs(c *check.C) {
	input := [][]string{
		{"rancher", "run", "--debug", "-itd"},
		{"rancher", "run", "--debug", "-itf=b"},
		{"rancher", "run", "--debug", "-itd#"},
		{"rancher", "run", "--debug", "-f=b"},
		{"rancher", "run", "--debug", "-=b"},
		{"rancher", "run", "--debug", "-"},
	}
	r0, err := parseArgs(input[0])
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(r0, check.DeepEquals, []string{"rancher", "run", "--debug", "-i", "-t", "-d"})

	r1, err := parseArgs(input[1])
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(r1, check.DeepEquals, []string{"rancher", "run", "--debug", "-i", "-t", "-f=b"})

	_, err = parseArgs(input[2])
	if err == nil {
		c.Fatal("should raise error")
	}

	r3, err := parseArgs(input[3])
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(r3, check.DeepEquals, []string{"rancher", "run", "--debug", "-f=b"})

	_, err = parseArgs(input[4])
	if err == nil {
		c.Fatal("should raise error")
	}

	r5, err := parseArgs(input[5])
	if err != nil {
		c.Fatal(err)
	}
	c.Assert(r5, check.DeepEquals, []string{"rancher", "run", "--debug", "-"})
}
