package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/smarty/gunit"
	"github.com/spf13/cobra"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"html/template"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type (
	goTestOutputRow struct {
		Time     string
		TestName string `json:"Test"`
		Action   string
		Package  string
		Elapsed  float64
		Output   string
	}

	testStatus struct {
		TestName           string
		Package            string
		ElapsedTime        float64
		Output             []string
		Passed             bool
		Skipped            bool
		Omitted            bool
		TestFileName       string
		TestFunctionDetail testFunctionFilePos
	}

	templateData struct {
		TestResultGroupIndicatorWidth  string
		TestResultGroupIndicatorHeight string
		TestResults                    []*testGroupData
		NumOfTestPassed                int
		NumOfTestFailed                int
		NumOfTestSkipped               int
		NumOfTests                     int
		TestDuration                   time.Duration
		ReportTitle                    string
		JsCode                         template.JS
		numOfTestsPerGroup             int
		OutputFilename                 string
		TestExecutionDate              string
		FailedTestNames                []string
	}

	testGroupData struct {
		FailureIndicator string
		SkippedIndicator string
		PackageName      string
		TestResults      []*testStatus
	}

	cmdFlags struct {
		titleFlag  string
		sizeFlag   string
		groupSize  int
		outputFlag string
		verbose    bool
	}

	goListJSONModule struct {
		Path string
		Dir  string
		Main bool
	}

	goListJSON struct {
		Dir         string
		ImportPath  string
		Name        string
		GoFiles     []string
		TestGoFiles []string
		Module      goListJSONModule
	}

	testFunctionFilePos struct {
		Line int
		Col  int
	}

	testFileDetail struct {
		FileName            string
		TestFunctionFilePos testFunctionFilePos
	}

	testFileDetailsByPackage map[string]map[string]*testFileDetail
)

func main() {
	rootCmd, _, _ := initRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initRootCommand() (*cobra.Command, *templateData, *cmdFlags) {
	flags := &cmdFlags{}
	tmplData := &templateData{}
	rootCmd := &cobra.Command{
		Use:  "go-test-report",
		Long: "Captures go test output via stdin and parses it into a single self-contained html file.",
		RunE: func(cmd *cobra.Command, args []string) (e error) {
			startTime := time.Now()
			if err := parseSizeFlag(tmplData, flags); err != nil {
				return err
			}
			tmplData.numOfTestsPerGroup = flags.groupSize
			tmplData.ReportTitle = flags.titleFlag
			tmplData.OutputFilename = flags.outputFlag
			if err := checkIfStdinIsPiped(); err != nil {
				return err
			}
			stdin := os.Stdin
			stdinScanner := bufio.NewScanner(stdin)
			testReportHTMLTemplateFile, _ := os.Create(tmplData.OutputFilename)
			reportFileWriter := bufio.NewWriter(testReportHTMLTemplateFile)
			defer func() {
				_ = stdin.Close()
				if err := reportFileWriter.Flush(); err != nil {
					e = err
				}
				if err := testReportHTMLTemplateFile.Close(); err != nil {
					e = err
				}
			}()
			startTestTime := time.Now()
			allPackageNames, allTests, failedTestNames, err := readTestDataFromStdIn(stdinScanner, flags, cmd)
			_, testsInPackages := formatAllTests(allTests)
			if err != nil {
				return errors.New(err.Error() + "\n")
			}
			elapsedTestTime := time.Since(startTestTime)
			// used to the location of test functions in test go files by package and test function name.
			testFileDetailByPackage, err := getPackageDetails(allPackageNames)
			if err != nil {
				return err
			}
			//err = generateReport(tmplData, newAllTests, failedTestNames, testFileDetailByPackage, elapsedTestTime, reportFileWriter)
			err = generateReportV2(tmplData, testsInPackages, failedTestNames, testFileDetailByPackage, elapsedTestTime, reportFileWriter)
			elapsedTime := time.Since(startTime)
			elapsedTimeMsg := []byte(fmt.Sprintf("[go-test-report] finished in %s\n", elapsedTime))
			if _, err := cmd.OutOrStdout().Write(elapsedTimeMsg); err != nil {
				return err
			}
			return nil
		},
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the version number of go-test-report",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg := fmt.Sprintf("go-test-report v%s", version)
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), msg); err != nil {
				return err
			}
			return nil
		},
	}
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().StringVarP(&flags.titleFlag,
		"title",
		"t",
		"go-test-report",
		"the title text shown in the test report")
	rootCmd.PersistentFlags().StringVarP(&flags.sizeFlag,
		"size",
		"s",
		"24",
		"the size (in pixels) of the clickable indicator for test result groups")
	rootCmd.PersistentFlags().IntVarP(&flags.groupSize,
		"groupSize",
		"g",
		20,
		"the number of tests per test group indicator")
	rootCmd.PersistentFlags().StringVarP(&flags.outputFlag,
		"output",
		"o",
		"test_report.html",
		"the HTML output file")
	rootCmd.PersistentFlags().BoolVarP(&flags.verbose,
		"verbose",
		"v",
		false,
		"while processing, show the complete output from go test ")

	return rootCmd, tmplData, flags
}

func readTestDataFromStdIn(stdinScanner *bufio.Scanner, flags *cmdFlags, cmd *cobra.Command) (allPackageNames map[string]*types.Nil, allTests map[string]*testStatus, failedTestNames []string, e error) {
	allTests = map[string]*testStatus{}
	allPackageNames = map[string]*types.Nil{}

	parentFailedTestNames := []string{}
	subFailedTestNames := []string{}

	// read from stdin and parse "go test" results
	for stdinScanner.Scan() {
		lineInput := stdinScanner.Bytes()

		if flags.verbose {
			newline := []byte("\n")
			if _, err := cmd.OutOrStdout().Write(append(lineInput, newline[0])); err != nil {
				return nil, nil, nil, err
			}
		}
		goTestOutputRow := &goTestOutputRow{}
		if err := json.Unmarshal(lineInput, goTestOutputRow); err != nil {
			return nil, nil, nil, err
		}
		goTestOutputRow.TestName = filterTestName(goTestOutputRow.TestName)

		if goTestOutputRow.TestName != "" {
			var status *testStatus
			key := goTestOutputRow.Package + "." + goTestOutputRow.TestName
			if _, exists := allTests[key]; !exists {
				status = &testStatus{
					TestName: goTestOutputRow.TestName,
					Package:  goTestOutputRow.Package,
					Output:   []string{},
				}
				allTests[key] = status
			} else {
				status = allTests[key]
			}
			if goTestOutputRow.Action == "pass" || goTestOutputRow.Action == "fail" || goTestOutputRow.Action == "skip" {
				isParentTest := !strings.Contains(goTestOutputRow.TestName, "/")
				if goTestOutputRow.Action == "fail" {
					if isParentTest {
						parentFailedTestNames = append(parentFailedTestNames, key)
					} else {
						subFailedTestNames = append(subFailedTestNames, key)
					}
				}

				if goTestOutputRow.Action == "pass" {
					status.Passed = true
					if isParentTest {
						status.Omitted = true
					}
				}
				if goTestOutputRow.Action == "skip" {
					status.Skipped = true
				}
				status.ElapsedTime = goTestOutputRow.Elapsed
			}
			allPackageNames[goTestOutputRow.Package] = nil

			status.Output = append(status.Output, goTestOutputRow.Output)
		}
	}

	failedTestNames = []string{}

	for _, parentName := range parentFailedTestNames {
		for _, subName := range subFailedTestNames {
			if strings.HasPrefix(subName, parentName) {
				allTests[parentName].Omitted = true
			}
		}
		if allTests[parentName].Omitted != true {
			failedTestNames = append(failedTestNames, parentName)
		}
	}

	failedTestNames = append(failedTestNames, subFailedTestNames...)

	return allPackageNames, allTests, failedTestNames, nil
}

func getPackageDetails(allPackageNames map[string]*types.Nil) (testFileDetailsByPackage, error) {
	var out bytes.Buffer
	var cmd *exec.Cmd
	testFileDetailByPackage := testFileDetailsByPackage{}
	stringReader := strings.NewReader("")
	for packageName := range allPackageNames {
		cmd = exec.Command("go", "list", "-json", packageName)
		out.Reset()
		stringReader.Reset("")
		cmd.Stdin = stringReader
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		goListJSON := &goListJSON{}
		if err := json.Unmarshal(out.Bytes(), goListJSON); err != nil {
			return nil, err
		}
		testFileDetailByPackage[packageName] = map[string]*testFileDetail{}
		for _, file := range goListJSON.TestGoFiles {
			sourceFilePath := fmt.Sprintf("%s/%s", goListJSON.Dir, file)
			fileSet := token.NewFileSet()
			f, err := parser.ParseFile(fileSet, sourceFilePath, nil, 0)
			if err != nil {
				return nil, err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					testFileDetail := &testFileDetail{}
					fileSetPos := fileSet.Position(n.Pos())
					folders := strings.Split(fileSetPos.String(), "/")
					fileNameWithPos := folders[len(folders)-1]
					fileDetails := strings.Split(fileNameWithPos, ":")
					lineNum, _ := strconv.Atoi(fileDetails[1])
					colNum, _ := strconv.Atoi(fileDetails[2])
					testFileDetail.FileName = fileDetails[0]
					testFileDetail.TestFunctionFilePos = testFunctionFilePos{
						Line: lineNum,
						Col:  colNum,
					}
					testFileDetailByPackage[packageName][x.Name.Name] = testFileDetail
				}
				return true
			})
		}
	}
	return testFileDetailByPackage, nil
}

type testRef struct {
	key  string
	name string
}
type byName []testRef

func (t byName) Len() int {
	return len(t)
}
func (t byName) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t byName) Less(i, j int) bool {
	return t[i].name < t[j].name
}

func generateReport(tmplData *templateData, allTests map[string]*testStatus, failedTestNames []string, testFileDetailByPackage testFileDetailsByPackage, elapsedTestTime time.Duration, reportFileWriter *bufio.Writer) error {
	// read the html template from the generated embedded asset go file
	tpl := template.New("test_report.html.template")
	testReportHTMLTemplateStr, err := hex.DecodeString(testReportHTMLTemplate)
	if err != nil {
		return err
	}
	tpl, err = tpl.Parse(string(testReportHTMLTemplateStr))
	if err != nil {
		return err
	}
	// read Javascript code from the generated embedded asset go file
	testReportJsCodeStr, err := hex.DecodeString(testReportJsCode)
	if err != nil {
		return err
	}

	tmplData.FailedTestNames = failedTestNames
	tmplData.NumOfTestPassed = 0
	tmplData.NumOfTestFailed = 0
	tmplData.NumOfTestSkipped = 0
	tmplData.JsCode = template.JS(testReportJsCodeStr)
	tgCounter := 0
	tgID := 0

	// sort the allTests map by test name (this will produce a consistent order when iterating through the map)
	var tests []testRef
	for test, status := range allTests {
		tests = append(tests, testRef{test, status.TestName})
	}
	sort.Sort(byName(tests))
	for _, test := range tests {
		status := allTests[test.key]
		if len(tmplData.TestResults) == tgID {
			tmplData.TestResults = append(tmplData.TestResults, &testGroupData{})
		}
		// add file info(name and position; line and col) associated with the test function
		testFileInfo := testFileDetailByPackage[status.Package][status.TestName]
		if testFileInfo != nil {
			status.TestFileName = testFileInfo.FileName
			status.TestFunctionDetail = testFileInfo.TestFunctionFilePos
		}
		tmplData.TestResults[tgID].TestResults = append(tmplData.TestResults[tgID].TestResults, status)
		if !status.Passed {
			if !status.Skipped {
				tmplData.TestResults[tgID].FailureIndicator = "failed"
				if !status.Omitted {
					tmplData.NumOfTestFailed++
				}
			} else {
				tmplData.TestResults[tgID].SkippedIndicator = "skipped"
				tmplData.NumOfTestSkipped++
			}
		} else {
			if !status.Omitted {
				tmplData.NumOfTestPassed++
			}
		}
		tgCounter++
		if tgCounter == tmplData.numOfTestsPerGroup {
			tgCounter = 0
			tgID++
		}
	}
	tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed + tmplData.NumOfTestSkipped
	tmplData.TestDuration = elapsedTestTime.Round(time.Millisecond)
	td := time.Now()
	tmplData.TestExecutionDate = fmt.Sprintf("%s %d, %d %02d:%02d:%02d",
		td.Month(), td.Day(), td.Year(), td.Hour(), td.Minute(), td.Second())
	if err := tpl.Execute(reportFileWriter, tmplData); err != nil {
		return err
	}
	return nil
}

func generateReportV2(tmplData *templateData, testsInPacakges map[string]map[string]*testStatus, failedTestNames []string, testFileDetailByPackage testFileDetailsByPackage, elapsedTestTime time.Duration, reportFileWriter *bufio.Writer) error {
	// read the html template from the generated embedded asset go file
	tpl := template.New("test_report.html.template")
	testReportHTMLTemplateStr, err := hex.DecodeString(testReportHTMLTemplate)
	if err != nil {
		return err
	}
	tpl, err = tpl.Parse(string(testReportHTMLTemplateStr))
	if err != nil {
		return err
	}
	// read Javascript code from the generated embedded asset go file
	testReportJsCodeStr, err := hex.DecodeString(testReportJsCode)
	if err != nil {
		return err
	}

	tmplData.FailedTestNames = failedTestNames
	tmplData.NumOfTestPassed = 0
	tmplData.NumOfTestFailed = 0
	tmplData.NumOfTestSkipped = 0
	tmplData.JsCode = template.JS(testReportJsCodeStr)
	tgID := 0

	// sort the allTests map by test name (this will produce a consistent order when iterating through the map)

	for packageName, allTests := range testsInPacakges {
		var tests []testRef
		for test, status := range allTests {
			tests = append(tests, testRef{test, status.TestName})
		}
		sort.Sort(byName(tests))
		tmplData.TestResults = append(tmplData.TestResults, &testGroupData{})
		for _, test := range tests {
			status := allTests[test.key]
			// add file info(name and position; line and col) associated with the test function
			testFileInfo := testFileDetailByPackage[status.Package][status.TestName]
			if testFileInfo != nil {
				status.TestFileName = testFileInfo.FileName
				status.TestFunctionDetail = testFileInfo.TestFunctionFilePos
			}
			tmplData.TestResults[tgID].TestResults = append(tmplData.TestResults[tgID].TestResults, status)
			if !status.Passed {
				if !status.Skipped {
					tmplData.TestResults[tgID].FailureIndicator = "failed"
					if !status.Omitted {
						tmplData.NumOfTestFailed++
					}
				} else {
					tmplData.NumOfTestSkipped++
				}
			} else {
				if !status.Omitted {
					tmplData.NumOfTestPassed++
				}
			}
		}
		tmplData.TestResults[tgID].PackageName = packageName
		tgID++
	}

	tmplData.NumOfTests = tmplData.NumOfTestPassed + tmplData.NumOfTestFailed + tmplData.NumOfTestSkipped
	tmplData.TestDuration = elapsedTestTime.Round(time.Millisecond)
	td := time.Now()
	tmplData.TestExecutionDate = fmt.Sprintf("%s %d, %d %02d:%02d:%02d",
		td.Month(), td.Day(), td.Year(), td.Hour(), td.Minute(), td.Second())
	if err := tpl.Execute(reportFileWriter, tmplData); err != nil {
		return err
	}
	return nil
}

func parseSizeFlag(tmplData *templateData, flags *cmdFlags) error {
	flags.sizeFlag = strings.ToLower(flags.sizeFlag)
	if !strings.Contains(flags.sizeFlag, "x") {
		val, err := strconv.Atoi(flags.sizeFlag)
		if err != nil {
			return err
		}
		tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", val)
		tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", val)
		return nil
	}
	if strings.Count(flags.sizeFlag, "x") > 1 {
		return errors.New(`malformed size value; only one x is allowed if specifying with and height`)
	}
	a := strings.Split(flags.sizeFlag, "x")
	valW, err := strconv.Atoi(a[0])
	if err != nil {
		return err
	}
	tmplData.TestResultGroupIndicatorWidth = fmt.Sprintf("%dpx", valW)
	valH, err := strconv.Atoi(a[1])
	if err != nil {
		return err
	}
	tmplData.TestResultGroupIndicatorHeight = fmt.Sprintf("%dpx", valH)
	return nil
}

func checkIfStdinIsPiped() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return err
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return nil
	}
	return errors.New("ERROR: missing ≪ stdin ≫ pipe")
}

type Output struct {
	isJson  bool
	line    string
	jsonObj map[string]interface{}
}
type OutputStatus struct {
	output Output
	time   time.Time
}

func (s OutputStatus) Less(other OutputStatus) bool {
	return s.time.Before(other.time)
}
func formatAllTests(allTests map[string]*testStatus) (map[string]*testStatus, map[string]map[string]*testStatus) {

	testsOutputs := make(map[string][]OutputStatus)
	testTitles := make(map[string]string)
	for key, status := range allTests {
		if _, ok := testsOutputs[key]; !ok {
			testsOutputs[key] = make([]OutputStatus, 0)
		}
		outputLine := ""
		for _, output := range status.Output {
			outputLine += output
			if !strings.HasSuffix(outputLine, "\n") {
				continue
			} else {
				jsonStr := strings.TrimSpace(outputLine)
				var jsonObj map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &jsonObj)
				o := Output{
					isJson: false,
					line:   outputLine,
				}
				out := OutputStatus{
					output: o,
					time:   time.Now(),
				}
				if err != nil {
					testsOutputs[key] = append(testsOutputs[key], out)
				} else {
					testName, foundTest := jsonObj[gunit.Test]
					packageName, foundPackage := jsonObj[gunit.Package]
					if foundTest && foundPackage {

						delete(jsonObj, gunit.Test)
						delete(jsonObj, gunit.Package)
						logTime := jsonObj["time"].(string)

						t, _ := time.Parse(time.RFC3339, logTime)
						o = Output{
							isJson:  true,
							jsonObj: jsonObj,
						}
						out = OutputStatus{
							output: o,
							time:   t,
						}
						if packageName != "" {
							newKey := packageName.(string) + "." + filterTestName(testName.(string))
							if _, ok := testsOutputs[newKey]; !ok {
								testsOutputs[newKey] = make([]OutputStatus, 0)
							}
							testsOutputs[newKey] = append(testsOutputs[newKey], out)
						} else {
							testsOutputs[key] = append(testsOutputs[key], out)
						}
					} else {
						testsOutputs[key] = append(testsOutputs[key], out)
					}
				}
				outputLine = ""
			}
		}
	}
	newAllTests := make(map[string]*testStatus)
	testsInPackages := make(map[string]map[string]*testStatus)

	genOutputs := func(key string, outputStatus []OutputStatus) []string {
		var outputs []string
		for _, item := range outputStatus {
			if item.output.isJson {
				l := ""
				level, foundLevel := item.output.jsonObj["level"].(string)
				if foundLevel {
					l = level
					delete(item.output.jsonObj, "level")
				}
				delete(item.output.jsonObj, "time")
				title, foundTitle := item.output.jsonObj[gunit.Title].(string)
				if foundTitle {
					testTitles[key] = title
				} else {
					if _, ok := item.output.jsonObj[gunit.RequestApi]; ok {
						bs, _ := json.MarshalIndent(item.output.jsonObj, "", "    ")
						outputs = append(outputs, fmt.Sprintf("---\n%s|%s ~ \n%s\n---\n", item.time, l, string(bs)))
					} else {
						bs, _ := json.Marshal(item.output.jsonObj)
						outputs = append(outputs, fmt.Sprintf("%s|%s ~ %s\n", item.time, l, string(bs)))
					}
				}
			} else {
				outputs = append(outputs, item.output.line)
			}
		}
		return outputs
	}

	for key, _ := range allTests {
		if _, ok := testsOutputs[key]; ok {
			if testsInPackages[allTests[key].Package] == nil {
				testsInPackages[allTests[key].Package] = make(map[string]*testStatus)
			}
			sort.Slice(testsOutputs[key], func(i, j int) bool {
				return OutputStatus{testsOutputs[key][i].output, testsOutputs[key][i].time}.Less(OutputStatus{testsOutputs[key][j].output, testsOutputs[key][j].time})
			})
			outputs := genOutputs(key, testsOutputs[key])
			if title, ok := testTitles[key]; ok {
				newKey := fmt.Sprintf("%s(%s)", key, title)
				newAllTests[newKey] = allTests[key]

				newAllTests[newKey].Output = outputs
				newAllTests[newKey].TestName = fmt.Sprintf("%s(%s)", newAllTests[newKey].TestName, title)
				testsInPackages[allTests[key].Package][newKey] = newAllTests[newKey]
			} else {
				newAllTests[key] = allTests[key]
				newAllTests[key].Output = outputs
				testsInPackages[allTests[key].Package][key] = newAllTests[key]
			}
		}
	}
	return newAllTests, testsInPackages
}

func filterTestName(name string) (out string) {
	out = name
	parts := strings.Split(name, "/"+gunit.FixtureParallel)
	if len(parts) == 2 {
		if parts[1] == "" {
			out = parts[0]
		} else {
			out = parts[0] + parts[1]
		}
	}
	return out
}
