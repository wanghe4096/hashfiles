package HashFiles

import (
	"testing"
	"path/filepath"
	"regexp"
	"github.com/stretchr/testify/assert"
)

func Test_walkFiles(t *testing.T) {
	list, err := filepath.Glob("/Users/wanghe/work/src/github.com/wanghe4096/")
	 assert.NotNil(t, list)
	assert.Nil(t, err)
}

func TestScanDirectoryForLogfiles(t *testing.T) {
	reg, _ := regexp.Compile(".idea/*")
	s := ScanDirectoryForLogfiles("./", reg)
	for _, v := range s{
		t.Log(v.Name())
	}
	assert.NotZero(t, len(s))
}

func TestRun(t *testing.T) {
	Run("./", ".idea/*")
}