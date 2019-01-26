package cmd

import (
	"testing"

	"gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	check.TestingT(t)
}

type CommonTestSuite struct {
}

var _ = check.Suite(&CommonTestSuite{})

func (s *CommonTestSuite) SetUpSuite(c *check.C) {

}

func (s *CommonTestSuite) TestParseClusterAndProjectID(c *check.C) {
	testParse(c, "local:p-12345", "local", "p-12345", false)
	testParse(c, "c-12345:p-12345", "c-12345", "p-12345", false)
	testParse(c, "cocal:p-12345", "", "", true)
	testParse(c, "c-123:p-123", "", "", true)
	testParse(c, "", "", "", true)
}

func testParse(c *check.C, testID, expectedCluster, expectedProject string, errorExpected bool) {
	actualCluster, actualProject, actualErr := parseClusterAndProjectID(testID)
	c.Assert(actualCluster, check.Equals, expectedCluster)
	c.Assert(actualProject, check.Equals, expectedProject)
	if errorExpected {
		c.Assert(actualErr, check.NotNil)
	} else {
		c.Assert(actualErr, check.IsNil)
	}
}
