package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestVersionCommand(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"version"})
	rootCmdErr := rootCmd.Execute()
	assertions.Nil(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal(fmt.Sprintf("go-test-report v%s\n", version), string(output))
}

func TestTitleFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title", "Sample Test Report"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("Sample Test Report", tmplData.ReportTitle)
	assertions.NotEmpty(output)
}

func TestTitleFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--title"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --title`)
}

func TestSizeFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, flags := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size", "24"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("24", flags.sizeFlag)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorWidth)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorHeight)
	assertions.NotEmpty(output)
}

func TestSizeFlagWithFullDimensions(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, flags := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size", "24x16"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("24x16", flags.sizeFlag)
	assertions.Equal("24px", tmplData.TestResultGroupIndicatorWidth)
	assertions.Equal("16px", tmplData.TestResultGroupIndicatorHeight)
	assertions.NotEmpty(output)
}

func TestSizeFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--size"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --size`)
}

func TestGroupSizeFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--groupSize", "32"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal(32, tmplData.numOfTestsPerGroup)
	assertions.NotEmpty(output)
}

func TestGroupSizeFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--groupSize"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --groupSize`)
}

func TestGroupOutputFlag(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, tmplData, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--output", "test_file.html"})
	rootCmdErr := rootCmd.Execute()
	assertions.Error(rootCmdErr)
	output, readErr := ioutil.ReadAll(buffer)
	assertions.Nil(readErr)
	assertions.Equal("test_file.html", tmplData.OutputFilename)
	assertions.NotEmpty(output)
}

func TestGroupOutputFlagIfMissingValue(t *testing.T) {
	assertions := assert.New(t)
	buffer := bytes.NewBufferString("")
	rootCmd, _, _ := initRootCommand()
	rootCmd.SetOut(buffer)
	rootCmd.SetArgs([]string{"--output"})
	rootCmdErr := rootCmd.Execute()
	assertions.NotNil(rootCmdErr)
	assertions.Equal(rootCmdErr.Error(), `flag needs an argument: --output`)
}

func TestReadTestDataFromStdIn(t *testing.T) {
	assertions := assert.New(t)
	flags := &cmdFlags{}
	data := `{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"go-test-report","Test":"TestFunc1"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc1","Output":"=== RUN   TestFunc1\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc1","Output":"--- PASS: TestFunc1 (1.25s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","Package":"go-test-report","Test":"TestFunc1","Elapsed":1.25}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"package2","Test":"TestFunc2"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"package2","Test":"TestFunc2","Output":"=== RUN   TestFunc2\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"package2","Test":"TestFunc2","Output":"--- PASS: TestFunc2 (0.25s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","Package":"package2","Test":"TestFunc2","Elapsed":0.25}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"go-test-report","Test":"TestFunc3"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc3","Output":"=== RUN   TestFunc3\n"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc3","Output":"sample output\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"go-test-report","Test":"TestFunc3","Output":"--- FAIL: TestFunc3 (0.00s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"fail","Package":"go-test-report","Test":"TestFunc3","Elapsed":0}
{"Time":"2021-11-10T21:28:34.882842+08:00","Action":"output","Package":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test3Report10Success\",\"time\":\"2021-11-10T21:28:34+08:00\",\"message\":\"resp:[{ID:90762 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159 PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159?txSecret=01365d8fd909f78fbbfe90f8c171993c\u0026txTime=618FBD65\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl: TransferAddr:10.144.25.67:8080 Status:1 StartTime:1636550885902 DispatchTimes:0 CTime:1636550885902 MTime:1636550907033 DomainID:1 QualityLevelID:10 TaskQualityLevelID:0 TaskType:2 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90764 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.s"}
{"Time":"2021-11-10T21:28:34.883377+08:00","Action":"output","Package":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"hopee.com/live/33872_id-test-1636550819514658-1636551159_hd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_hd?txSecret=1144f8a67dfe64ae3f36d8e5a7ce507d\u0026txTime=618FBD7D\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl:hd TransferAddr:- Status:1 StartTime:1636550909745 DispatchTimes:1 CTime:1636550909747 MTime:1636550909747 DomainID:1 QualityLevelID:10 TaskQualityLevelID:10 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90765 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_sd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_sd?txSecret=625c6a314de35e6a80f6e2ff7eeb3e3e\u0026txTime=618FBD7D\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=T"}
{"Time":"2021-11-10T21:28:34.88985+08:00","Action":"output","Package":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"XCLOUD\u0026session_id=1636551159 TranscodeTpl:sd TransferAddr:- Status:1 StartTime:1636550909745 DispatchTimes:1 CTime:1636550909749 MTime:1636550909749 DomainID:1 QualityLevelID:10 TaskQualityLevelID:20 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90763 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_flu PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_flu?txSecret=57332a1862c318d74e899e9891cf4051\u0026txTime=618FBD69\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl:flu TransferAddr:- Status:1 StartTime:1636550889754 DispatchTimes:1 CTime:1636550889756 MTime:1636550907033 DomainID:1 QualityLevelID:10 TaskQualityLevelID:30 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomain"}
{"Time":"2021-11-10T21:28:34.895624+08:00","Action":"output","Package":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"ID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0}]\"}\n"}
{"Time":"2021-11-10T21:28:34.895651+08:00","Action":"output","Package":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-10T21:28:34+08:00\",\"message\":\"try to delete roomId:1636550819514658\"}\n"}
`
	stdinScanner := bufio.NewScanner(strings.NewReader(data))
	cmd := &cobra.Command{}
	allPackageNames, allTests, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
	formatAllTests(allTests)
	assertions.Nil(err)
	assertions.Len(allPackageNames, 3)
	assertions.Contains(allPackageNames, "go-test-report")
	assertions.Contains(allPackageNames, "package2")
	assertions.Len(allTests, 4)
	assertions.Contains(allTests, "go-test-report.TestFunc1")
	assertions.Contains(allTests, "package2.TestFunc2")
	assertions.Contains(allTests, "go-test-report.TestFunc3")

	val := allTests["go-test-report.TestFunc1"]
	assertions.True(val.Passed)
	assertions.Equal("TestFunc1", val.TestName)
	assertions.Equal(1.25, val.ElapsedTime)
	assertions.Len(val.Output, 2)
	assertions.Equal("=== RUN   TestFunc1\n", val.Output[0])
	assertions.Equal("--- PASS: TestFunc1 (1.25s)", val.Output[1])
	assertions.Equal(0, val.TestFunctionDetail.Line)
	assertions.Equal(0, val.TestFunctionDetail.Col)

	val = allTests["package2.TestFunc2"]
	assertions.True(val.Passed)
	assertions.Equal("TestFunc2", val.TestName)
	assertions.Equal(0.25, val.ElapsedTime)
	assertions.Len(val.Output, 2)
	assertions.Equal("=== RUN   TestFunc2\n", val.Output[0])
	assertions.Equal("--- PASS: TestFunc2 (0.25s)", val.Output[1])
	assertions.Equal(0, val.TestFunctionDetail.Line)
	assertions.Equal(0, val.TestFunctionDetail.Col)

	val = allTests["go-test-report.TestFunc3"]
	assertions.False(val.Passed)
	assertions.Equal("TestFunc3", val.TestName)
	assertions.Equal(0.00, val.ElapsedTime)
	assertions.Len(val.Output, 3)
	assertions.Equal("=== RUN   TestFunc3\n", val.Output[0])
	assertions.Equal("--- FAIL: TestFunc3 (0.00s)\n", val.Output[2])
	assertions.Equal(0, val.TestFunctionDetail.Line)
	assertions.Equal(0, val.TestFunctionDetail.Col)
}

func TestGenerateReport(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{
		TestResultGroupIndicatorWidth:  "20px",
		TestResultGroupIndicatorHeight: "16px",
		ReportTitle:                    "test-title",
		numOfTestsPerGroup:             2,
		OutputFilename:                 "test-output-report.html",
	}
	allTests := map[string]*testStatus{}
	allTests["TestFunc1"] = &testStatus{
		TestName:           "TestFunc1",
		Package:            "go-test-report",
		ElapsedTime:        0,
		Output:             nil,
		Passed:             true,
		TestFileName:       "",
		TestFunctionDetail: testFunctionFilePos{},
	}
	allTests["TestFunc2"] = &testStatus{
		TestName:     "TestFunc2",
		Package:      "package2",
		ElapsedTime:  0,
		Output:       nil,
		Passed:       true,
		TestFileName: "",

		TestFunctionDetail: testFunctionFilePos{},
	}
	allTests["TestFunc3"] = &testStatus{
		TestName:           "TestFunc3",
		Package:            "go-test-report",
		ElapsedTime:        0,
		Output:             nil,
		Passed:             false,
		TestFileName:       "",
		TestFunctionDetail: testFunctionFilePos{},
	}
	allTests["TestFunc4"] = &testStatus{
		TestName:           "TestFunc4",
		Package:            "go-test-report",
		ElapsedTime:        0,
		Output:             nil,
		Passed:             false,
		Skipped:            true,
		TestFileName:       "",
		TestFunctionDetail: testFunctionFilePos{},
	}
	testFileDetailsByPackage := testFileDetailsByPackage{}
	testFileDetailsByPackage["go-test-report"] = map[string]*testFileDetail{}
	testFileDetailsByPackage["go-test-report"]["TestFunc1"] = &testFileDetail{
		FileName: "sample_file_1.go",
		TestFunctionFilePos: testFunctionFilePos{
			Line: 101,
			Col:  1,
		},
	}
	testFileDetailsByPackage["package2"] = map[string]*testFileDetail{}
	testFileDetailsByPackage["package2"]["TestFunc2"] = &testFileDetail{
		FileName: "sample_file_2.go",
		TestFunctionFilePos: testFunctionFilePos{
			Line: 784,
			Col:  17,
		},
	}
	testFileDetailsByPackage["go-test-report"]["TestFunc3"] = &testFileDetail{
		TestFunctionFilePos: testFunctionFilePos{
			Line: 0,
			Col:  0,
		},
	}
	elapsedTestTime := 3 * time.Second
	writer := bufio.NewWriter(&bytes.Buffer{})
	err := generateReport(tmplData, allTests, testFileDetailsByPackage, elapsedTestTime, writer)
	assertions.Nil(err)
	assertions.Equal(2, tmplData.NumOfTestPassed)
	assertions.Equal(1, tmplData.NumOfTestFailed)
	assertions.Equal(1, tmplData.NumOfTestSkipped)
	assertions.Equal(4, tmplData.NumOfTests)

	assertions.Equal("TestFunc1", tmplData.TestResults[0].TestResults[0].TestName)
	assertions.Equal("go-test-report", tmplData.TestResults[0].TestResults[0].Package)
	assertions.Equal(true, tmplData.TestResults[0].TestResults[0].Passed)
	assertions.Equal("sample_file_1.go", tmplData.TestResults[0].TestResults[0].TestFileName)
	assertions.Equal(1, tmplData.TestResults[0].TestResults[0].TestFunctionDetail.Col)
	assertions.Equal(101, tmplData.TestResults[0].TestResults[0].TestFunctionDetail.Line)

	assertions.Equal("TestFunc2", tmplData.TestResults[0].TestResults[1].TestName)
	assertions.Equal("package2", tmplData.TestResults[0].TestResults[1].Package)
	assertions.Equal(true, tmplData.TestResults[0].TestResults[1].Passed)
	assertions.Equal("sample_file_2.go", tmplData.TestResults[0].TestResults[1].TestFileName)
	assertions.Equal(17, tmplData.TestResults[0].TestResults[1].TestFunctionDetail.Col)
	assertions.Equal(784, tmplData.TestResults[0].TestResults[1].TestFunctionDetail.Line)

	assertions.Equal("TestFunc3", tmplData.TestResults[1].TestResults[0].TestName)
	assertions.Equal("go-test-report", tmplData.TestResults[1].TestResults[0].Package)
	assertions.Equal(false, tmplData.TestResults[1].TestResults[0].Passed)
	assertions.Empty(tmplData.TestResults[1].TestResults[0].TestFileName)
	assertions.Equal(0, tmplData.TestResults[1].TestResults[0].TestFunctionDetail.Col)
	assertions.Equal(0, tmplData.TestResults[1].TestResults[0].TestFunctionDetail.Line)
}

func TestSameTestName(t *testing.T) {
	assertions := assert.New(t)
	flags := &cmdFlags{}
	data := `{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"foo","Test":"Test"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"foo","Test":"Test","Output":"=== RUN   Test\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"foo","Test":"Test","Output":"--- PASS: Test (1.5s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","Package":"foo","Test":"Test","Elapsed":1.5}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","Package":"bar","Test":"Test"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","Package":"bar","Test":"Test","Output":"=== RUN   Test\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","Package":"bar","Test":"Test","Output":"--- FAIL: Test (0.5s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"fail","Package":"bar","Test":"Test","Elapsed":0.5}
`
	stdinScanner := bufio.NewScanner(strings.NewReader(data))
	cmd := &cobra.Command{}
	allPackageNames, allTests, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
	assertions.Nil(err)
	assertions.Len(allPackageNames, 2)
	assertions.Contains(allPackageNames, "foo")
	assertions.Contains(allPackageNames, "bar")
	assertions.Len(allTests, 2)
}

func TestParseSizeFlagIfValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "x",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "": invalid syntax`)

}

func TestParseSizeFlagIfWidthValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "Bx27",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "b": invalid syntax`)
}

func TestParseSizeFlagIfHeightValueIsNotInteger(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "10xA",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `strconv.Atoi: parsing "a": invalid syntax`)
}

func TestParseSizeFlagIfMalformedSize(t *testing.T) {
	assertions := assert.New(t)
	tmplData := &templateData{}
	flags := &cmdFlags{
		sizeFlag: "10xx19",
	}
	err := parseSizeFlag(tmplData, flags)
	assertions.Error(err)
	assertions.Equal(err.Error(), `malformed size value; only one x is allowed if specifying with and height`)
}
