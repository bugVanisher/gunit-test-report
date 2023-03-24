package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
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
	//data := `{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","PackageName":"go-test-report","Test":"TestFunc1"}
	//{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"go-test-report","Test":"TestFunc1","Output":"=== RUN   TestFunc1\n"}
	//{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","PackageName":"go-test-report","Test":"TestFunc1","Output":"--- PASS: TestFunc1 (1.25s)\n"}
	//{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","PackageName":"go-test-report","Test":"TestFunc1","Elapsed":1.25}
	//{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","PackageName":"package2","Test":"TestFunc2"}
	//{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"package2","Test":"TestFunc2","Output":"=== RUN   TestFunc2\n"}
	//{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","PackageName":"package2","Test":"TestFunc2","Output":"--- PASS: TestFunc2 (0.25s)\n"}
	//{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","PackageName":"package2","Test":"TestFunc2","Elapsed":0.25}
	//{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","PackageName":"go-test-report","Test":"TestFunc3"}
	//{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"go-test-report","Test":"TestFunc3","Output":"=== RUN   TestFunc3\n"}
	//{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"go-test-report","Test":"TestFunc3","Output":"sample output\n"}
	//{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","PackageName":"go-test-report","Test":"TestFunc3","Output":"--- FAIL: TestFunc3 (0.00s)\n"}
	//{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"fail","PackageName":"go-test-report","Test":"TestFunc3","Elapsed":0}
	//{"Time":"2021-11-10T21:28:34.882842+08:00","Action":"output","PackageName":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test3Report10Success\",\"time\":\"2021-11-10T21:28:34+08:00\",\"message\":\"resp:[{ID:90762 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159 PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159?txSecret=01365d8fd909f78fbbfe90f8c171993c\u0026txTime=618FBD65\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl: TransferAddr:10.144.25.67:8080 Status:1 StartTime:1636550885902 DispatchTimes:0 CTime:1636550885902 MTime:1636550907033 DomainID:1 QualityLevelID:10 TaskQualityLevelID:0 TaskType:2 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90764 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.s"}
	//{"Time":"2021-11-10T21:28:34.883377+08:00","Action":"output","PackageName":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"hopee.com/live/33872_id-test-1636550819514658-1636551159_hd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_hd?txSecret=1144f8a67dfe64ae3f36d8e5a7ce507d\u0026txTime=618FBD7D\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl:hd TransferAddr:- Status:1 StartTime:1636550909745 DispatchTimes:1 CTime:1636550909747 MTime:1636550909747 DomainID:1 QualityLevelID:10 TaskQualityLevelID:10 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90765 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_sd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_sd?txSecret=625c6a314de35e6a80f6e2ff7eeb3e3e\u0026txTime=618FBD7D\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=T"}
	//{"Time":"2021-11-10T21:28:34.88985+08:00","Action":"output","PackageName":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"XCLOUD\u0026session_id=1636551159 TranscodeTpl:sd TransferAddr:- Status:1 StartTime:1636550909745 DispatchTimes:1 CTime:1636550909749 MTime:1636550909749 DomainID:1 QualityLevelID:10 TaskQualityLevelID:20 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90763 AppID:LS RoomID:LS:1636550819514658 SessionID:LS:1636551159 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_flu PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636550819514658-1636551159_flu?txSecret=57332a1862c318d74e899e9891cf4051\u0026txTime=618FBD69\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636551159 TranscodeTpl:flu TransferAddr:- Status:1 StartTime:1636550889754 DispatchTimes:1 CTime:1636550889756 MTime:1636550907033 DomainID:1 QualityLevelID:10 TaskQualityLevelID:30 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomain"}
	//{"Time":"2021-11-10T21:28:34.895624+08:00","Action":"output","PackageName":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"ID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0}]\"}\n"}
	//{"Time":"2021-11-10T21:28:34.895651+08:00","Action":"output","PackageName":"command-line-arguments","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-10T21:28:34+08:00\",\"message\":\"try to delete roomId:1636550819514658\"}\n"}
	//`
	data := `{"Time":"2021-11-11T15:25:45.650928+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode"}
{"Time":"2021-11-11T15:25:45.650964+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode","Output":"=== RUN   TestFullModeRoomTransCode\n"}
{"Time":"2021-11-11T15:25:45.651359+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode","Output":"=== PAUSE TestFullModeRoomTransCode\n"}
{"Time":"2021-11-11T15:25:45.651373+08:00","Action":"pause","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode"}
{"Time":"2021-11-11T15:25:45.651392+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode"}
{"Time":"2021-11-11T15:25:45.651398+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"=== RUN   TestRoomNoTransCode\n"}
{"Time":"2021-11-11T15:25:45.651654+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"=== PAUSE TestRoomNoTransCode\n"}
{"Time":"2021-11-11T15:25:45.651669+08:00","Action":"pause","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode"}
{"Time":"2021-11-11T15:25:45.651677+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode"}
{"Time":"2021-11-11T15:25:45.651684+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"=== RUN   TestRoomReport2TransCode\n"}
{"Time":"2021-11-11T15:25:45.652142+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"=== PAUSE TestRoomReport2TransCode\n"}
{"Time":"2021-11-11T15:25:45.652186+08:00","Action":"pause","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode"}
{"Time":"2021-11-11T15:25:45.652209+08:00","Action":"cont","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode"}
{"Time":"2021-11-11T15:25:45.652219+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode","Output":"=== CONT  TestFullModeRoomTransCode\n"}
{"Time":"2021-11-11T15:25:45.652232+08:00","Action":"cont","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode"}
{"Time":"2021-11-11T15:25:45.65224+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"=== CONT  TestRoomNoTransCode\n"}
{"Time":"2021-11-11T15:25:45.652417+08:00","Action":"cont","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode"}
{"Time":"2021-11-11T15:25:45.652438+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"=== CONT  TestRoomReport2TransCode\n"}
{"Time":"2021-11-11T15:25:45.652797+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.652751 I | get roomId: 1636615545652973\n"}
{"Time":"2021-11-11T15:25:45.652854+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.652801 I | get roomId: 1636615545652103\n"}
{"Time":"2021-11-11T15:25:45.65291+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.652874 I | get roomId: 1636615545652302\n"}
{"Time":"2021-11-11T15:25:45.694869+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.693\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:90\tzk services node\t{\"serviceName\": \"livetech-configer-test-id\", \"cost\": 40, \"services\": [{\"name\":\"livetech-configer-test-id\",\"version\":\"\",\"metadata\":{},\"endpoints\":[],\"nodes\":[{\"id\":\"/services/livetech-configer-test-id/10.144.48.118:8080\",\"address\":\"10.144.48.118:8080\",\"metadata\":null}]}]}\n"}
{"Time":"2021-11-11T15:25:45.694952+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.694\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:130\tset watch\t{\"serviceName\": \"livetech-configer-test-id\"}\n"}
{"Time":"2021-11-11T15:25:45.751649+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.751\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/watcher.go:55\twatch watchDir ChildrenW\t{\"event\": {\"Type\":0,\"State\":0,\"Path\":\"\",\"Err\":null,\"Server\":\"\"}, \"childrenPaths\": [\"10.144.48.118:8080\"], \"stat\": {\"Czxid\":4295015266,\"Mzxid\":4295015266,\"Ctime\":1606895424101,\"Mtime\":1606895424101,\"Version\":0,\"Cversion\":153,\"Aversion\":0,\"EphemeralOwner\":0,\"DataLength\":0,\"NumChildren\":1,\"Pzxid\":25782234203}, \"zk\": \"10.129.100.231:2181\"}\n"}
{"Time":"2021-11-11T15:25:45.823753+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"CreateRooms return {Success: 1, FailRooms: []}\n"}
{"Time":"2021-11-11T15:25:45.823799+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"CreateRooms return {Success: 1, FailRooms: []}\n"}
{"Time":"2021-11-11T15:25:45.849238+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"CreateRooms return {Success: 1, FailRooms: []}\n"}
{"Time":"2021-11-11T15:25:45.864704+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:25:45+08:00\",\"message\":\"AddTranscodeWhiteUser Success roomId:1636615545652103\"}\n"}
{"Time":"2021-11-11T15:25:45.890456+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.890\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:90\tzk services node\t{\"serviceName\": \"livetech-streamapi-test-id\", \"cost\": 41, \"services\": [{\"name\":\"livetech-streamapi-test-id\",\"version\":\"\",\"metadata\":{},\"endpoints\":[],\"nodes\":[{\"id\":\"/services/livetech-streamapi-test-id/10.144.53.102:8080\",\"address\":\"10.144.53.102:8080\",\"metadata\":null}]}]}\n"}
{"Time":"2021-11-11T15:25:45.890503+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.890\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:130\tset watch\t{\"serviceName\": \"livetech-streamapi-test-id\"}\n"}
{"Time":"2021-11-11T15:25:45.933757+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"2021-11-11 15:25:45.933\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/watcher.go:55\twatch watchDir ChildrenW\t{\"event\": {\"Type\":0,\"State\":0,\"Path\":\"\",\"Err\":null,\"Server\":\"\"}, \"childrenPaths\": [\"10.144.53.102:8080\"], \"stat\": {\"Czxid\":12885074552,\"Mzxid\":12885074552,\"Ctime\":1608030391719,\"Mtime\":1608030391719,\"Version\":0,\"Cversion\":259,\"Aversion\":0,\"EphemeralOwner\":0,\"DataLength\":0,\"NumChildren\":1,\"Pzxid\":25782594109}, \"zk\": \"10.129.100.231:2181\"}\n"}
{"Time":"2021-11-11T15:25:46.151243+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"{\"level\":\"info\",\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:25:46+08:00\",\"message\":\"AddTranscodeWhiteUser Success roomId:1636615545652302\"}\n"}
{"Time":"2021-11-11T15:25:48.893174+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"{\"Test\":\"TestRoomNoTransCode\",\"time\":\"2021-11-11T15:25:48+08:00\",\"message\":\"get push url list err:{\\\"id\\\":\\\"go.micro.client\\\",\\\"code\\\":408,\\\"detail\\\":\\\"context deadline exceeded, callee_addr:10.144.53.102:8080, cost:3043ms\\\",\\\"status\\\":\\\"Request Timeout\\\"}\"}\n"}
{"Time":"2021-11-11T15:25:48.893232+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"{\"Test\":\"TestRoomNoTransCode\",\"time\":\"2021-11-11T15:25:48+08:00\",\"message\":\"try to delete roomId:1636615545652973\"}\n"}
{"Time":"2021-11-11T15:25:49.218301+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"resp: {}\r\n"}
{"Time":"2021-11-11T15:25:49.218344+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"--- FAIL: TestRoomNoTransCode (3.57s)\n"}
{"Time":"2021-11-11T15:25:49.218354+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"    livetech.go:51: failFast ...\n"}
{"Time":"2021-11-11T15:26:45.693682+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"2021-11-11 15:26:45.693\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:90\tzk services node\t{\"serviceName\": \"livetech-streamapi-test-id\", \"cost\": 41, \"services\": [{\"name\":\"livetech-streamapi-test-id\",\"version\":\"\",\"metadata\":{},\"endpoints\":[],\"nodes\":[{\"id\":\"/services/livetech-streamapi-test-id/10.144.53.102:8080\",\"address\":\"10.144.53.102:8080\",\"metadata\":null}]}]}\n"}
{"Time":"2021-11-11T15:26:45.734418+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"2021-11-11 15:26:45.734\tinfo\t/Users/yuhe/git/livetech_api_autotest/vendor/git.garena.com/shopee/feed/microkit/registry/cache.go:90\tzk services node\t{\"serviceName\": \"livetech-configer-test-id\", \"cost\": 40, \"services\": [{\"name\":\"livetech-configer-test-id\",\"version\":\"\",\"metadata\":{},\"endpoints\":[],\"nodes\":[{\"id\":\"/services/livetech-configer-test-id/10.144.48.118:8080\",\"address\":\"10.144.48.118:8080\",\"metadata\":null}]}]}\n"}
{"Time":"2021-11-11T15:26:46.008316+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:26:46+08:00\",\"message\":\"args:[/usr/local/bin/ffmpeg -stream_loop -1 -re -i /Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv -s 360x640 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652103-1636615928?txSecret=7107f2d54806dd5a721a7d369ee683c4\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636615928]\"}\n"}
{"Time":"2021-11-11T15:26:46.206447+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:26:46+08:00\",\"message\":\"args:[/usr/local/bin/ffmpeg -stream_loop -1 -re -i /Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv -s 360x640 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652302-1636616005?txSecret=7a56631eb766fd9f4d1962e37e9dfb5a\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636616005]\"}\n"}
{"Time":"2021-11-11T15:26:49.016058+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:26:49+08:00\",\"message\":\"check timeout... start process success?\"}\n"}
{"Time":"2021-11-11T15:26:49.102005+08:00","Action":"fail","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomNoTransCode","Elapsed":3.57}
{"Time":"2021-11-11T15:26:49.10205+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success"}
{"Time":"2021-11-11T15:26:49.102057+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"=== RUN   TestRoomReport2TransCode/Test1Report30Success\n"}
{"Time":"2021-11-11T15:26:49.102112+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"2021-11-11 15:26:49.102042 I | in testcase now ...\n"}
{"Time":"2021-11-11T15:26:49.213141+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:26:49+08:00\",\"message\":\"check timeout... start process success?\"}\n"}
{"Time":"2021-11-11T15:26:49.362773+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success"}
{"Time":"2021-11-11T15:26:49.362811+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"=== RUN   TestFullModeRoomTransCode/Test1Transcode30Success\n"}
{"Time":"2021-11-11T15:26:49.362827+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"2021-11-11 15:26:49.362782 I | in testcase now ...\n"}
{"Time":"2021-11-11T15:26:49.362835+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"{\"Test\":\"TestFullModeRoomTransCode/Test1Transcode30Success\",\"time\":\"2021-11-11T15:26:49+08:00\",\"message\":\"transcodenum: 1\"}\n"}
{"Time":"2021-11-11T15:26:49.599001+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"{\"Test\":\"TestRoomReport2TransCode/Test1Report30Success\",\"time\":\"2021-11-11T15:26:49+08:00\",\"message\":\"transcodenum: 1\"}\n"}
{"Time":"2021-11-11T15:26:59.903628+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test1Report30Success\",\"time\":\"2021-11-11T15:26:59+08:00\",\"message\":\"need  1 tasks , and there are 1 \"}\n"}
{"Time":"2021-11-11T15:26:59.991224+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success"}
{"Time":"2021-11-11T15:26:59.991271+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"=== RUN   TestRoomReport2TransCode/Test2Report30Success\n"}
{"Time":"2021-11-11T15:26:59.991284+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"2021-11-11 15:26:59.991236 I | in testcase now ...\n"}
{"Time":"2021-11-11T15:27:00.034085+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"{\"Test\":\"TestRoomReport2TransCode/Test2Report30Success\",\"time\":\"2021-11-11T15:27:00+08:00\",\"message\":\"transcodenum: 1\"}\n"}
{"Time":"2021-11-11T15:27:09.684435+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"{\"level\":\"info\",\"Test\":\"TestFullModeRoomTransCode/Test1Transcode30Success\",\"time\":\"2021-11-11T15:27:09+08:00\",\"message\":\"need  1 tasks , and there are 1 \"}\n"}
{"Time":"2021-11-11T15:27:09.776553+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess"}
{"Time":"2021-11-11T15:27:09.776591+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"=== RUN   TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess\n"}
{"Time":"2021-11-11T15:27:09.776623+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"2021-11-11 15:27:09.776567 I | in testcase now ...\n"}
{"Time":"2021-11-11T15:27:09.776803+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:09+08:00\",\"message\":\"stop \u0026{FFmpeg:/usr/local/bin/ffmpeg input:/Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv inputOpts:[-stream_loop -1 -re] output:rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652302-1636616005?txSecret=7a56631eb766fd9f4d1962e37e9dfb5a\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636616005 outputOpts:[-s 360x640 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac] child:0xc00199cc60 Running:true TestName:TestFullModeRoomTransCode t:0xc000cee100}\"}\n"}
{"Time":"2021-11-11T15:27:09.776834+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:09+08:00\",\"message\":\"args:[/usr/local/bin/ffmpeg -stream_loop -1 -re -i /Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv -s 720x1080 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652302-1636616005?txSecret=7a56631eb766fd9f4d1962e37e9dfb5a\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636616005]\"}\n"}
{"Time":"2021-11-11T15:27:10.116522+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test2Report30Success\",\"time\":\"2021-11-11T15:27:10+08:00\",\"message\":\"need  1 tasks , and there are 1 \"}\n"}
{"Time":"2021-11-11T15:27:10.411553+08:00","Action":"run","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success"}
{"Time":"2021-11-11T15:27:10.41159+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"=== RUN   TestRoomReport2TransCode/Test3Report10Success\n"}
{"Time":"2021-11-11T15:27:10.411639+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"2021-11-11 15:27:10.411579 I | in testcase now ...\n"}
{"Time":"2021-11-11T15:27:10.768306+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestRoomReport2TransCode/Test3Report10Success\",\"time\":\"2021-11-11T15:27:10+08:00\",\"message\":\"transcodenum: 3\"}\n"}
{"Time":"2021-11-11T15:27:12.78111+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:12+08:00\",\"message\":\"check timeout... start process success?\"}\n"}
{"Time":"2021-11-11T15:27:12.781155+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess\",\"time\":\"2021-11-11T15:27:12+08:00\",\"message\":\"transcodenum: 3\"}\n"}
{"Time":"2021-11-11T15:27:21.103269+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test3Report10Success\",\"time\":\"2021-11-11T15:27:21+08:00\",\"message\":\"need  3 tasks , and there are 3 \"}\n"}
{"Time":"2021-11-11T15:27:21.19093+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode/Test3Report10Success\",\"time\":\"2021-11-11T15:27:21+08:00\",\"message\":\"resp:[{ID:90886 AppID:LS RoomID:LS:1636615545652103 SessionID:LS:1636615928 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928 PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928?txSecret=83c93494c8edaa601917e8a68a252c3a\u0026txTime=6190BA3B\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636615928 TranscodeTpl: TransferAddr:10.144.25.67:8080 Status:1 StartTime:1636615611317 DispatchTimes:0 CTime:1636615611317 MTime:1636615632705 DomainID:1 QualityLevelID:10 TaskQualityLevelID:0 TaskType:2 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90890 AppID:LS RoomID:LS:1636615545652103 SessionID:LS:1636615928 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.s"}
{"Time":"2021-11-11T15:27:21.19109+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"hopee.com/live/33872_id-test-1636615545652103-1636615928_hd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928_hd?txSecret=ae596403f7bf998274c1fb82105c1e32\u0026txTime=6190BA53\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636615928 TranscodeTpl:hd TransferAddr:- Status:1 StartTime:1636615635729 DispatchTimes:1 CTime:1636615635736 MTime:1636615635736 DomainID:1 QualityLevelID:10 TaskQualityLevelID:10 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90891 AppID:LS RoomID:LS:1636615545652103 SessionID:LS:1636615928 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928_sd PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928_sd?txSecret=077108a28e19a020cf69e9b17a13c7ec\u0026txTime=6190BA53\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=T"}
{"Time":"2021-11-11T15:27:21.191938+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"XCLOUD\u0026session_id=1636615928 TranscodeTpl:sd TransferAddr:- Status:1 StartTime:1636615635729 DispatchTimes:1 CTime:1636615635740 MTime:1636615635740 DomainID:1 QualityLevelID:10 TaskQualityLevelID:20 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomainID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0} {ID:90888 AppID:LS RoomID:LS:1636615545652103 SessionID:LS:1636615928 CdnID:TXCLOUD FlowRatio:1 PullURL:rtmp://rtmp-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928_flu PushURL:rtmp://push-tx.livestream.shopee.com/live/33872_id-test-1636615545652103-1636615928_flu?txSecret=add66f80d068f787431fe7c008ddeaec\u0026txTime=6190BA3F\u0026pushDomain=push-tx.livestream.shopee.com\u0026cdnID=TXCLOUD\u0026session_id=1636615928 TranscodeTpl:flu TransferAddr:- Status:1 StartTime:1636615615718 DispatchTimes:1 CTime:1636615615720 MTime:1636615632705 DomainID:1 QualityLevelID:10 TaskQualityLevelID:30 TaskType:0 ClientIP: CostMode:1 SrcCdnID:TXCLOUD SrcDomain"}
{"Time":"2021-11-11T15:27:21.192726+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"ID:1 TaskSrc:1 EndTime:0 BrokenTimes:0 BrokenDuration:0 Bitrate:0 Fps:0 Delay:0 PacketLossRate:0 HealthDegree:0}]\"}\n"}
{"Time":"2021-11-11T15:27:21.193401+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:27:21+08:00\",\"message\":\"try to delete roomId:1636615545652103\"}\n"}
{"Time":"2021-11-11T15:27:21.370599+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"resp: {}\r\n"}
{"Time":"2021-11-11T15:27:21.411652+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:27:21+08:00\",\"message\":\"DeleteTranscodeWhiteUser Success roomId:1636615545652103\"}\n"}
{"Time":"2021-11-11T15:27:21.45571+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestRoomReport2TransCode\",\"time\":\"2021-11-11T15:27:21+08:00\",\"message\":\"stop \u0026{FFmpeg:/usr/local/bin/ffmpeg input:/Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv inputOpts:[-stream_loop -1 -re] output:rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652103-1636615928?txSecret=7107f2d54806dd5a721a7d369ee683c4\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636615928 outputOpts:[-s 360x640 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac] child:0xc001470420 Running:true TestName:TestRoomReport2TransCode t:0xc000ceed00}\"}\n"}
{"Time":"2021-11-11T15:27:21.455761+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Output":"--- PASS: TestRoomReport2TransCode (95.81s)\n"}
{"Time":"2021-11-11T15:27:21.455853+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"    --- PASS: TestRoomReport2TransCode/Test1Report30Success (10.89s)\n"}
{"Time":"2021-11-11T15:27:21.455881+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"        test_case.go:71: Test definition:\n"}
{"Time":"2021-11-11T15:27:21.455896+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Output":"            \n"}
{"Time":"2021-11-11T15:27:21.45635+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test1Report30Success","Elapsed":10.89}
{"Time":"2021-11-11T15:27:21.456363+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"    --- PASS: TestRoomReport2TransCode/Test2Report30Success (10.42s)\n"}
{"Time":"2021-11-11T15:27:21.45637+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"        test_case.go:71: Test definition:\n"}
{"Time":"2021-11-11T15:27:21.456377+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Output":"            \n"}
{"Time":"2021-11-11T15:27:21.456783+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test2Report30Success","Elapsed":10.42}
{"Time":"2021-11-11T15:27:21.456794+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"    --- PASS: TestRoomReport2TransCode/Test3Report10Success (10.78s)\n"}
{"Time":"2021-11-11T15:27:21.456801+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"        test_case.go:71: Test definition:\n"}
{"Time":"2021-11-11T15:27:21.456809+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"            \n"}
{"Time":"2021-11-11T15:27:33.093819+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess\",\"time\":\"2021-11-11T15:27:33+08:00\",\"message\":\"need  3 tasks , and there are 3 \"}\n"}
{"Time":"2021-11-11T15:27:33.180078+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:33+08:00\",\"message\":\"try to delete roomId:1636615545652302\"}\n"}
{"Time":"2021-11-11T15:27:33.27087+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"resp: {}\r\n"}
{"Time":"2021-11-11T15:27:33.311607+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"level\":\"info\",\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:33+08:00\",\"message\":\"DeleteTranscodeWhiteUser Success roomId:1636615545652302\"}\n"}
{"Time":"2021-11-11T15:27:33.362319+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Output":"{\"Test\":\"TestFullModeRoomTransCode\",\"time\":\"2021-11-11T15:27:33+08:00\",\"message\":\"stop \u0026{FFmpeg:/usr/local/bin/ffmpeg input:/Users/yuhe/git/livetech_api_autotest/resource/video/korea_girls_vertical.flv inputOpts:[-stream_loop -1 -re] output:rtmp://push-tx.lvb.test.shopee.co.id/live/33872_id-test-1636615545652302-1636616005?txSecret=7a56631eb766fd9f4d1962e37e9dfb5a\u0026txTime=6190BA38\u0026pushDomain=push-tx.lvb.test.shopee.co.id\u0026cdnID=TXCLOUD\u0026session_id=1636616005 outputOpts:[-s 720x1080 -f flv -flvflags no_duration_filesize -vcodec libx264 -acodec aac] child:0xc0013ee3f0 Running:true TestName:TestFullModeRoomTransCode t:0xc000cee100}\"}\n"}
{"Time":"2021-11-11T15:27:33.362373+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode/Test3Report10Success","Elapsed":10.78}
{"Time":"2021-11-11T15:27:33.362471+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestRoomReport2TransCode","Elapsed":95.81}
{"Time":"2021-11-11T15:27:33.362484+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode","Output":"--- PASS: TestFullModeRoomTransCode (107.72s)\n"}
{"Time":"2021-11-11T15:27:33.3625+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"    --- PASS: TestFullModeRoomTransCode/Test1Transcode30Success (20.41s)\n"}
{"Time":"2021-11-11T15:27:33.362512+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"        test_case.go:71: Test definition:\n"}
{"Time":"2021-11-11T15:27:33.363035+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Output":"            \n"}
{"Time":"2021-11-11T15:27:33.363054+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test1Transcode30Success","Elapsed":20.41}
{"Time":"2021-11-11T15:27:33.363063+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"    --- PASS: TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess (23.40s)\n"}
{"Time":"2021-11-11T15:27:33.363227+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"        test_case.go:71: Test definition:\n"}
{"Time":"2021-11-11T15:27:33.363239+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Output":"            \n"}
{"Time":"2021-11-11T15:27:33.363247+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode/Test2UpgradeBitrate2HighLevelSuccess","Elapsed":23.4}
{"Time":"2021-11-11T15:27:33.363253+08:00","Action":"pass","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Test":"TestFullModeRoomTransCode","Elapsed":107.72}
{"Time":"2021-11-11T15:27:33.363552+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Output":"FAIL\n"}
{"Time":"2021-11-11T15:27:33.384935+08:00","Action":"output","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Output":"FAIL\tgit.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel\t109.622s\n"}
{"Time":"2021-11-11T15:27:33.384987+08:00","Action":"fail","PackageName":"git.garena.com/shopee/live-streaming/livetech_qa/livetech_api_autotest/testcases/streamapi/ReportQualityLevel","Elapsed":109.622}
`
	stdinScanner := bufio.NewScanner(strings.NewReader(data))
	cmd := &cobra.Command{}
	allPackageNames, allTests, _, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
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
		Passed:             false,
		Omitted:            true,
		TestFileName:       "",
		TestFunctionDetail: testFunctionFilePos{},
	}
	allTests["Parent/TestFunc1"] = &testStatus{
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
	testReportHTMLTemplateFile, _ := os.Create("test.html")
	reportFileWriter := bufio.NewWriter(testReportHTMLTemplateFile)
	defer func() {
		if err := reportFileWriter.Flush(); err != nil {
		}
		if err := testReportHTMLTemplateFile.Close(); err != nil {
		}
	}()
	err := generateReport(tmplData, allTests, nil, testFileDetailsByPackage, elapsedTestTime, reportFileWriter)
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

func TestGenerateReportV2(t *testing.T) {
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
		Passed:             false,
		Omitted:            true,
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
		Package:            "package2",
		ElapsedTime:        0,
		Output:             nil,
		Passed:             true,
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
	testsInPackages := make(map[string]map[string]*testStatus)
	testsInPackages["package2"] = make(map[string]*testStatus)
	testsInPackages["package2"]["TestFunc3"] = allTests["TestFunc3"]
	testsInPackages["package2"]["TestFunc2"] = allTests["TestFunc2"]
	testsInPackages["go-test-report"] = make(map[string]*testStatus)
	testsInPackages["go-test-report"]["TestFunc1"] = allTests["TestFunc1"]
	testsInPackages["go-test-report"]["TestFunc4"] = allTests["TestFunc4"]

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
	testFileDetailsByPackage["package2"]["TestFunc3"] = &testFileDetail{
		FileName: "sample_file_3.go",
		TestFunctionFilePos: testFunctionFilePos{
			Line: 99,
			Col:  17,
		},
	}
	testFileDetailsByPackage["go-test-report"]["TestFunc4"] = &testFileDetail{
		TestFunctionFilePos: testFunctionFilePos{
			Line: 0,
			Col:  0,
		},
	}
	elapsedTestTime := 3 * time.Second
	testReportHTMLTemplateFile, _ := os.Create("test.html")
	reportFileWriter := bufio.NewWriter(testReportHTMLTemplateFile)
	defer func() {
		if err := reportFileWriter.Flush(); err != nil {
		}
		if err := testReportHTMLTemplateFile.Close(); err != nil {
		}
	}()
	err := generateReportV2(tmplData, testsInPackages, nil, testFileDetailsByPackage, elapsedTestTime, reportFileWriter)
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
	data := `{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","PackageName":"foo","Test":"Test"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"foo","Test":"Test","Output":"=== RUN   Test\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","PackageName":"foo","Test":"Test","Output":"--- PASS: Test (1.5s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"pass","PackageName":"foo","Test":"Test","Elapsed":1.5}
{"Time":"2020-07-10T01:24:44.269511-05:00","Action":"run","PackageName":"bar","Test":"Test"}
{"Time":"2020-07-10T01:24:44.270071-05:00","Action":"output","PackageName":"bar","Test":"Test","Output":"=== RUN   Test\n"}
{"Time":"2020-07-10T01:24:44.270295-05:00","Action":"output","PackageName":"bar","Test":"Test","Output":"--- FAIL: Test (0.5s)\n"}
{"Time":"2020-07-10T01:24:44.270311-05:00","Action":"fail","PackageName":"bar","Test":"Test","Elapsed":0.5}
`
	stdinScanner := bufio.NewScanner(strings.NewReader(data))
	cmd := &cobra.Command{}
	allPackageNames, allTests, _, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
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

func TestHexEncode(t *testing.T) {
	// 读取模板文件
	templateData, err := ioutil.ReadFile("test_report.html.template")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 读取JS文件
	jsData, err := ioutil.ReadFile("test_report.js")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将文件内容编码为字符串
	templateString := string(templateData)
	jsString := string(jsData)

	hstr := hex.EncodeToString([]byte(templateString))
	jstr := hex.EncodeToString([]byte(jsString))

	fmt.Println(hstr)
	fmt.Println(jstr)
}
