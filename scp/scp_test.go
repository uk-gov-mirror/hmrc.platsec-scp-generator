package scp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	rc := m.Run()
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < .99 {
			fmt.Println("Tests passed but coverage failed at ", c)
			rc = -1
		}
	}
	os.Exit(rc)
}

//TestGenerateServiceName tests a service name can be
//created from the incoming scanner event_source
func TestGenerateServiceName(t *testing.T) {
	eventSource := "s3.amazonaws.com"
	serviceName := ServiceName(eventSource)
	assert.Equal(t, "s3", serviceName)
}

//TestLoadScannerReport tests that a scanner report can
//be loaded
func TestLoadScannerReport(t *testing.T) {
	scannerFileName := "./testdata/s3_scanner_report.json"
	scannerFileData, err := LoadScannerFile(scannerFileName)

	if err != nil {
		t.Fatalf("cannot read: %s, got %v",
			scannerFileName, err)
	}

	if len(scannerFileData) <= 0 {
		t.Fatalf("fixture %s is empty file", scannerFileData)
	}
}

//TestDirectorCheckTrue tests directoryCheck returns true for
//existing directory
func TestDirectoryCheckTrue(t *testing.T) {
	directory := "../scp/"
	actual, _ := directoryCheck(directory)

	assert.True(t, true, actual)
}

//TestDirectoryCheckFalse test directoryCheck returns false for
//a non existent directory
func TestDirectoryCheckFalse(t *testing.T) {
	directory := "../scpfalse/"
	expected := false

	actual, _ := directoryCheck(directory)

	assert.False(t, expected, actual)

}

//TestDecodeFile decodes the file to a map
func TestDecodeFile(t *testing.T) {
	jsonData := getScannerMessage()
	testStub := jsonFileStub{inputData: jsonData}
	testData := testStub.getData()
	reports, _ := GenerateReport(testData)
	report := *reports

	assert.NotNil(t, report)
	assert.Equal(t, 10, len(report[0].Results.ServiceUsage))
}

//TestDecodeFileError returns an error
func TestDecodeFileError(t *testing.T) {
	jsonData := getCorruptedScannerMessage()
	testStub := jsonFileStub{inputData: jsonData}
	testData := testStub.getData()
	_, err := GenerateReport(testData)
	assert.Error(t, err)
}

//TestGenerateAllowListData tests that
//API actions above a threshold are mapped to
//A new data structure
func TestGenerateAllowListData(t *testing.T) {
	testData := getTestReport()
	r := *testData
    apiFn := GreaterThan

	cases := []struct {
		threshold int64
		report    Report
		expected  int64
	}{
		{
			threshold: 10,
			report:    r[0],
			expected:  8,
		},
		{
			threshold: 100,
			report:    r[0],
			expected:  3,
		},
		{
			threshold: 3000,
			report:    r[0],
			expected:  0,
		},
	}

	for _, c := range cases {
		allowList, _ := GenerateList(c.threshold, &c.report, apiFn)
		assert.NotNil(t, allowList)
		assert.Equal(t, c.expected, int64(len(allowList)))
	}
}

//TestGenerateDenyListData tests that
//API actions above a threshold are mapped to
//A new data structure
func TestGenerateDenyListData(t *testing.T) {
	testData := getTestReport()
	r := *testData
	apiFn := LessThan

	cases := []struct {
		threshold int64
		report    Report
		expected  int64
	}{
		{
			threshold: 10,
			report:    r[0],
			expected:  2,
		},
		{
			threshold: 100,
			report:    r[0],
			expected:  7,
		},
		{
			threshold: 3000,
			report:    r[0],
			expected:  10,
		},
	}

	for _, c := range cases {
		denyList, _ := GenerateList(c.threshold, &c.report, apiFn)
		assert.NotNil(t, denyList)
		assert.Equal(t, c.expected, int64(len(denyList)))
	}
}

//TestGenerateAllowListGeneratesError tests for
//An error being returned for zero and negative
//Thresholds
func TestGenerateAllowListGeneratesError(t *testing.T) {
	testReports := getTestReport()
	testReport := *testReports
	apiFn := GreaterThan

	cases := []struct {
		threshold int64
		report    Report
		expected  error
	}{
		{
			threshold: 0,
			report:    testReport[0],
			expected:  ErrInvalidParameters,
		},
		{
			threshold: -1,
			report:    testReport[0],
			expected:  ErrInvalidParameters,
		},
	}

	for _, c := range cases {
		_, err := GenerateList(c.threshold, &c.report, apiFn)
		assert.Error(t, err)
	}
}

//TestGenerateAllowSCP test that we can
//generate an SCP from an Allow List
func TestGenerateAllowSCP(t *testing.T) {
	allowList := getTestAllowListFilteredData()
	scpType := "Allow"
	awsService := "s3"
	generated := GenerateSCP(scpType, awsService, allowList)

	assert.Equal(t, "2012-10-17", generated.Version)
}

//TestSaveSCP tests that we can save an SCP report
func TestSaveSCP(t *testing.T) {
	testSCP := getTestSCP("Allow", "S3")

	SCPSaved := SaveSCP(testSCP)

	assert.Nil(t, SCPSaved)
}

//TestGetSCPType test that the SCPType is returned
func TestGetSCPType(t *testing.T) {
	testConfig := SCPConfig{SCPType:"Allow",ScannerFile: "TestFile", Threshold: 34}
	actual := testConfig.ServiceType()
	assert.Equal(t, "Allow",*actual)
}

//TestGetScannerFilename test that the SCPType is returned
func TestGetScannerFilename(t *testing.T) {
	testConfig := SCPConfig{SCPType:"Allow",ScannerFile: "TestFile", Threshold: 34}
	actual := testConfig.ScannerFilename()
	assert.Equal(t, "TestFile",*actual)
}

//TestGetThreshold test that the SCPType is returned
func TestGetThreshold(t *testing.T) {
	testConfig := SCPConfig{SCPType:"Allow",ScannerFile: "TestFile", Threshold: 34}
	actual := testConfig.ThresholdLimit()
	assert.Equal(t, 34,int(*actual))
}

//TestLoadScannerFileReturnsError test that an error is
//returned
func TestLoadScannerFileReturnsError(t *testing.T) {
	testFile := "testFile"
	fileData, err := LoadScannerFile(testFile)

	assert.NotNil(t, err)
	assert.Nil(t, fileData)
}

//TestSCPTypeParameterPass tests that we do not
//fail when we pass the correct parameter types
func TestSCPTypeParameterPass (t *testing.T) {
	cases := []struct {
		value string
		expected bool
	}{
		{
			value: "Allow",
			expected: true,
		},
		{
			value: "Deny",
			expected: true,
		},
		{
			value: "deny",
			expected: true,
		},
		{
			value: "allow",
			expected: true,
		},
	}

	for _, c := range cases {
		actual := CheckSCPParameter(c.value)
		assert.Equal(t, c.expected,actual)
	}
}


//TestSCPTypeParameterReturnsFalse tests that we do not
//fail when we pass the correct parameter types
func TestSCPTypeParameterReturnsFalse (t *testing.T) {
	cases := []struct {
		value string
		expected bool
	}{
		{
			value: "Allowime",
			expected: false,
		},
		{
			value: "Denyme",
			expected: false,
		},
		{
			value: "denyme",
			expected: false,
		},
		{
			value: "allowme",
			expected: false,
		},
	}

	for _, c := range cases {
		actual := CheckSCPParameter(c.value)
		assert.Equal(t, c.expected,actual)
	}
}

//JSONFileDataStub
type jsonFileStub struct {
	inputData string
}

func (j jsonFileStub) getData() []byte {
	return []byte(j.inputData)
}

//getScannerMessage returns a full scanner message
func getCorruptedScannerMessage() string {
	scannerMessage := `
[
   "rresults": {
      "event_source": "s3.amazon.com",
      "service_usage": [
        {
          "event_name": "ListObjectVersions",
          "count": 15
        },
        {
          "event_name": "ListObjects",
          "count": 224
        },
        {
          "event_name": "GetBucketEncryption",
          "count": 11
        },
        {
          "event_name": "CreateMultipartUpload",
          "count": 6
        },
        {
          "event_name": "GetObject",
          "count": 205
        },
        {
          "event_name": "GetBucketLifecycle",
          "count": 1
        },
        {
          "event_name": "ListBuckets",
          "count": 125
        },
        {
          "event_name": "GetBucketPolicy",
          "count": 19
        },
        {
          "event_name": "GetBucketVersioning",
          "count": 52
        },
        {
          "event_name": "PutObject",
          "count": 31
        }
      ]
    }
  }
]
`
	return scannerMessage
}

//getScannerMessage returns a full scanner message
func getScannerMessage() string {
	scannerMessage := `
[
  {
    "account": {
      "identifier": "999888777666",
      "name": "some account"
    },
    "description": "AWS s3 service usage scan",
    "partition": {
      "year": "2021",
      "month": "03"
    },
    "results": {
      "event_source": "s3.amazon.com",
      "service_usage": [
        {
          "event_name": "ListObjectVersions",
          "count": 15
        },
        {
          "event_name": "ListObjects",
          "count": 224
        },
        {
          "event_name": "GetBucketEncryption",
          "count": 11
        },
        {
          "event_name": "CreateMultipartUpload",
          "count": 6
        },
        {
          "event_name": "GetObject",
          "count": 205
        },
        {
          "event_name": "GetBucketLifecycle",
          "count": 1
        },
        {
          "event_name": "ListBuckets",
          "count": 125
        },
        {
          "event_name": "GetBucketPolicy",
          "count": 19
        },
        {
          "event_name": "GetBucketVersioning",
          "count": 52
        },
        {
          "event_name": "PutObject",
          "count": 31
        }
      ]
    }
  }
]
`
	return scannerMessage
}

//getTestAllowListFilteredData returns a filtered data set
func getTestAllowListFilteredData() map[string]int64 {
	filteredData := map[string]int64{
		"LookupEvents":                     10,
		"ListTags":                         1656,
		"GetEventSelectors":                12,
		"BatchGetBuilds":                   223,
		"GetLambdaFunctionRecommendations": 2343,
		"DescribeSecurityGroups":           543,
		"DescribeVpcs":                     48,
		"ListStacks":                       348,
	}

	return filteredData
}

//getTestReport returns a report in the
//form of a serialised json document
func getTestReport() *[]Report {
	jsonData := getScannerMessage()
	testStub := jsonFileStub{inputData: jsonData}
	testData := testStub.getData()
	report, _ := GenerateReport(testData)
	return report
}

func getTestSCP(scpType string, awsService string) SCP {
	allowList := getTestAllowListFilteredData()
	testSCP := GenerateSCP(scpType, awsService, allowList)
	return testSCP
}
