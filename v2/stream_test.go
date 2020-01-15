package xlsx

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	. "gopkg.in/check.v1"
)

const (
	TestsShouldMakeRealFiles = false
)

type StreamSuite struct{}

var _ = Suite(&StreamSuite{})

func (s *StreamSuite) TestTestsShouldMakeRealFilesShouldBeFalse(t *C) {
	if TestsShouldMakeRealFiles {
		t.Fatal("TestsShouldMakeRealFiles should only be true for local debugging. Don't forget to switch back before commiting.")
	}
}

func TestXlsxStreamWrite(t *testing.T) {
	// When shouldMakeRealFiles is set to true this test will make actual XLSX files in the file system.
	// This is useful to ensure files open in Excel, Numbers, Google Docs, etc.
	// In case of issues you can use "Open XML SDK 2.5" to diagnose issues in generated XLSX files:
	// https://www.microsoft.com/en-us/download/details.aspx?id=30425
	testCases := []struct {
		testName      string
		sheetNames    []string
		workbookData  [][][]string
		headerTypes   [][]*CellType
		expectedError error
	}{
		{
			testName: "One Sheet",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
			},
			headerTypes: [][]*CellType{
				{nil, CellTypeString.Ptr(), nil, CellTypeString.Ptr()},
			},
		},
		{
			testName: "One Column",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					{"Token"},
					{"123"},
				},
			},
		},
		{
			testName: "Several Sheets, with different numbers of columns and rows",
			sheetNames: []string{
				"Sheet 1", "Sheet 2", "Sheet3",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU", "Stock"},
					{"456", "Salsa", "200", "0346", "1"},
					{"789", "Burritos", "400", "754", "3"},
				},
				{
					{"Token", "Name", "Price"},
					{"9853", "Guacamole", "500"},
					{"2357", "Margarita", "700"},
				},
			},
		},
		{
			testName: "Two Sheets with same the name",
			sheetNames: []string{
				"Sheet 1", "Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU", "Stock"},
					{"456", "Salsa", "200", "0346", "1"},
					{"789", "Burritos", "400", "754", "3"},
				},
			},
			expectedError: fmt.Errorf("duplicate sheet name '%s'.", "Sheet 1"),
		},
		{
			testName: "One Sheet Registered, tries to write to two",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU"},
					{"456", "Salsa", "200", "0346"},
				},
			},
			expectedError: AlreadyOnLastSheetError,
		},
		{
			testName: "One Sheet, too many columns in row 1",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123", "asdf"},
				},
			},
			expectedError: WrongNumberOfRowsError,
		},
		{
			testName: "One Sheet, too few columns in row 1",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300"},
				},
			},
			expectedError: WrongNumberOfRowsError,
		},
		{
			testName: "Lots of Sheets, only writes rows to one, only writes headers to one, should not error and should still create a valid file",
			sheetNames: []string{
				"Sheet 1", "Sheet 2", "Sheet 3", "Sheet 4", "Sheet 5", "Sheet 6",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
				{{"Id", "Unit Cost"}},
				{{}},
				{{}},
				{{}},
			},
		},
		{
			testName: "Two Sheets, only writes to one, should not error and should still create a valid file",
			sheetNames: []string{
				"Sheet 1", "Sheet 2",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
			},
		},
		{
			testName: "Larger Sheet",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
				},
			},
		},
		{
			testName: "UTF-8 Characters. This XLSX File loads correctly with Excel, Numbers, and Google Docs. It also passes Microsoft's Office File Format Validator.",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					// String courtesy of https://github.com/minimaxir/big-list-of-naughty-strings/
					// Header row contains the tags that I am filtering on
					{"Token", endSheetDataTag, "Price", fmt.Sprintf(dimensionTag, "A1:D1")},
					// Japanese and emojis
					{"123", "パーティーへ行かないか", "300", "🍕🐵 🙈 🙉 🙊"},
					// XML encoder/parser test strings
					{"123", `<?xml version="1.0" encoding="ISO-8859-1"?>`, "300", `<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE foo [ <!ELEMENT foo ANY ><!ENTITY xxe SYSTEM "file:///etc/passwd" >]><foo>&xxe;</foo>`},
					// Upside down text and Right to Left Arabic text
					{"123", `˙ɐnbᴉlɐ ɐuƃɐɯ ǝɹolop ʇǝ ǝɹoqɐl ʇn ʇunpᴉpᴉɔuᴉ ɹodɯǝʇ poɯsnᴉǝ op pǝs 'ʇᴉlǝ ƃuᴉɔsᴉdᴉpɐ ɹnʇǝʇɔǝsuoɔ 'ʇǝɯɐ ʇᴉs ɹolop ɯnsdᴉ ɯǝɹo˥
					00˙Ɩ$-`, "300", `ﷺ`},
					{"123", "Taco", "300", "0000000123"},
				},
			},
		},
	}
	for i, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			var filePath string
			var buffer bytes.Buffer
			if TestsShouldMakeRealFiles {
				filePath = fmt.Sprintf("Workbook%d.xlsx", i)
			}
			err := writeStreamFile(filePath, &buffer, testCase.sheetNames, testCase.workbookData, testCase.headerTypes, TestsShouldMakeRealFiles)
			switch {
			case err != nil && testCase.expectedError == nil:
				t.Fatalf("Unexpected error: %v", err.Error())
			case err == nil && testCase.expectedError != nil:
				t.Fatalf("Error is nil, but expected error was: %v", testCase.expectedError)
			case err != nil && testCase.expectedError != nil && err.Error() != testCase.expectedError.Error():
				t.Fatalf("Error differs from expected error. Error: %v, Expected Error: %v ", err, testCase.expectedError)
			}
			if testCase.expectedError != nil {
				return
			}
			// read the file back with the xlsx package
			var bufReader *bytes.Reader
			var size int64
			if !TestsShouldMakeRealFiles {
				bufReader = bytes.NewReader(buffer.Bytes())
				size = bufReader.Size()
			}
			actualSheetNames, actualWorkbookData, _ := readXLSXFile(t, filePath, bufReader, size, TestsShouldMakeRealFiles)
			// check if data was able to be read correctly
			if !reflect.DeepEqual(actualSheetNames, testCase.sheetNames) {
				t.Fatal("Expected sheet names to be equal")
			}
			if !reflect.DeepEqual(actualWorkbookData, testCase.workbookData) {
				t.Fatal("Expected workbook data to be equal")
			}

		})
	}
}

func TestXlsxStreamWriteWithDefaultCellType(t *testing.T) {
	// When shouldMakeRealFiles is set to true this test will make actual XLSX files in the file system.
	// This is useful to ensure files open in Excel, Numbers, Google Docs, etc.
	// In case of issues you can use "Open XML SDK 2.5" to diagnose issues in generated XLSX files:
	// https://www.microsoft.com/en-us/download/details.aspx?id=30425
	testCases := []struct {
		testName             string
		sheetNames           []string
		workbookData         [][][]string
		expectedWorkbookData [][][]string
		headerTypes          [][]*StreamingCellMetadata
		expectedError        error
	}{
		{
			testName: "One Sheet",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300.0", "0000000123"},
					{"123", "Taco", "string", "0000000123"},
				},
			},
			expectedWorkbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300.00", "0000000123"},
					{"123", "Taco", "string", "0000000123"},
				},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultDecimalStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr()},
			},
		},
		{
			testName: "One Column",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					{"Token"},
					{"1234"},
				},
			},
			expectedWorkbookData: [][][]string{
				{
					{"Token"},
					{"1234.00"},
				},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultDecimalStreamingCellMetadata.Ptr()},
			},
		},
		{
			testName: "Several Sheets, with different numbers of columns and rows",
			sheetNames: []string{
				"Sheet 1", "Sheet 2", "Sheet3",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU", "Stock"},
					{"456", "Salsa", "200", "0346", "1"},
					{"789", "Burritos", "400", "754", "3"},
				},
				{
					{"Token", "Name", "Price"},
					{"9853", "Guacamole", "500"},
					{"2357", "Margarita", "700"},
				},
			},
			expectedWorkbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300.00", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU", "Stock"},
					{"456", "Salsa", "200.00", "0346", "1"},
					{"789", "Burritos", "400.00", "754", "3"},
				},
				{
					{"Token", "Name", "Price"},
					{"9853", "Guacamole", "500"},
					{"2357", "Margarita", "700"},
				},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultIntegerStreamingCellMetadata.Ptr(), nil, DefaultDecimalStreamingCellMetadata.Ptr(), nil},
				{DefaultIntegerStreamingCellMetadata.Ptr(), nil, DefaultDecimalStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultIntegerStreamingCellMetadata.Ptr()},
				{nil, nil, nil},
			},
		},
		{
			testName: "Two Sheets with the same name",
			sheetNames: []string{
				"Sheet 1", "Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU", "Stock"},
					{"456", "Salsa", "200", "0346", "1"},
					{"789", "Burritos", "400", "754", "3"},
				},
			},
			expectedError: fmt.Errorf("duplicate sheet name '%s'.", "Sheet 1"),
		},
		{
			testName: "One Sheet Registered, tries to write to two",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{
					{"Token", "Name", "Price", "SKU"},
					{"456", "Salsa", "200", "0346"},
				},
			},
			expectedError: AlreadyOnLastSheetError,
		},
		{
			testName: "One Sheet, too many columns in row 1",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123", "asdf"},
				},
			},
			expectedError: WrongNumberOfRowsError,
		},
		{
			testName: "One Sheet, too few columns in row 1",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300"},
				},
			},
			expectedError: WrongNumberOfRowsError,
		},
		{
			testName: "Lots of Sheets, only writes rows to one, only writes headers to one, should not error and should still create a valid file",
			sheetNames: []string{
				"Sheet 1", "Sheet 2", "Sheet 3", "Sheet 4", "Sheet 5", "Sheet 6",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
				{{"Id", "Unit Cost"}},
				{{}},
				{{}},
				{{}},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultIntegerStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultIntegerStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr()},
				{nil},
				{nil, nil},
				{nil},
				{nil},
				{nil},
			},
			expectedWorkbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
				{{"Id", "Unit Cost"}},
				{{}},
				{{}},
				{{}},
			},
		},
		{
			testName: "Two Sheets, only writes to one, should not error and should still create a valid file",
			sheetNames: []string{
				"Sheet 1", "Sheet 2",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
			},
			expectedWorkbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123"},
				},
				{{}},
			},

			headerTypes: [][]*StreamingCellMetadata{
				{DefaultDateStreamingCellMetadata.Ptr(), DefaultDateStreamingCellMetadata.Ptr(), DefaultDateStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr()},
				{nil},
			},
		},
		{
			testName: "Larger Sheet",
			sheetNames: []string{
				"Sheet 1",
			},
			workbookData: [][][]string{
				{
					{"Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU", "Token", "Name", "Price", "SKU"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
					{"123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123", "123", "Taco", "300", "0000000123"},
					{"456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346", "456", "Salsa", "200", "0346"},
					{"789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754", "789", "Burritos", "400", "754"},
				},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultIntegerStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultIntegerStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr()},
			},
		},
		{
			testName: "UTF-8 Characters. This XLSX File loads correctly with Excel, Numbers, and Google Docs. It also passes Microsoft's Office File Format Validator.",
			sheetNames: []string{
				"Sheet1",
			},
			workbookData: [][][]string{
				{
					// String courtesy of https://github.com/minimaxir/big-list-of-naughty-strings/
					// Header row contains the tags that I am filtering on
					{"Token", endSheetDataTag, "Price", fmt.Sprintf(dimensionTag, "A1:D1")},
					// Japanese and emojis
					{"123", "パーティーへ行かないか", "300", "🍕🐵 🙈 🙉 🙊"},
					// XML encoder/parser test strings
					{"123", `<?xml version="1.0" encoding="ISO-8859-1"?>`, "300", `<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE foo [ <!ELEMENT foo ANY ><!ENTITY xxe SYSTEM "file:///etc/passwd" >]><foo>&xxe;</foo>`},
					// Upside down text and Right to Left Arabic text
					{"123", `˙ɐnbᴉlɐ ɐuƃɐɯ ǝɹolop ʇǝ ǝɹoqɐl ʇn ʇunpᴉpᴉɔuᴉ ɹodɯǝʇ poɯsnᴉǝ op pǝs 'ʇᴉlǝ ƃuᴉɔsᴉdᴉpɐ ɹnʇǝʇɔǝsuoɔ 'ʇǝɯɐ ʇᴉs ɹolop ɯnsdᴉ ɯǝɹo˥
						00˙Ɩ$-`, "300", `ﷺ`},
					{"123", "Taco", "300", "0000000123"},
				},
			},
			headerTypes: [][]*StreamingCellMetadata{
				{DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr(), DefaultStringStreamingCellMetadata.Ptr()},
			},
		},
	}
	for i, testCase := range testCases {
		// if testCase.testName != "Lots of Sheets, only writes rows to one, only writes headers to one, should not error and should still create a valid file" {
		// 	continue
		// }
		t.Run(testCase.testName, func(t *testing.T) {

			var filePath string
			var buffer bytes.Buffer
			if TestsShouldMakeRealFiles {
				filePath = fmt.Sprintf("WorkbookTyped%d.xlsx", i)
			}
			err := writeStreamFileWithDefaultMetadata(filePath, &buffer, testCase.sheetNames, testCase.workbookData, testCase.headerTypes, TestsShouldMakeRealFiles)
			switch {
			case err == nil && testCase.expectedError != nil:
				t.Fatalf("Expected an error, but nil was returned\n")
			case err != nil && testCase.expectedError == nil:
				t.Fatalf("Unexpected error: %q", err.Error())
			case err != testCase.expectedError && err.Error() != testCase.expectedError.Error():
				t.Fatalf("Error differs from expected error. Error: %v, Expected Error: %v ", err, testCase.expectedError)
			case err != nil:
				// We got an error we expected
				return
			}

			// read the file back with the xlsx package
			var bufReader *bytes.Reader
			var size int64
			if !TestsShouldMakeRealFiles {
				bufReader = bytes.NewReader(buffer.Bytes())
				size = bufReader.Size()
			}
			actualSheetNames, actualWorkbookData, workbookCellTypes := readXLSXFile(t, filePath, bufReader, size, TestsShouldMakeRealFiles)
			verifyCellTypesInColumnMatchHeaderType(t, workbookCellTypes, testCase.headerTypes, testCase.workbookData)
			// check if data was able to be read correctly
			if !reflect.DeepEqual(actualSheetNames, testCase.sheetNames) {
				t.Fatal("Expected sheet names to be equal")
			}
			if testCase.expectedWorkbookData == nil {
				testCase.expectedWorkbookData = testCase.workbookData
			}
			if !reflect.DeepEqual(actualWorkbookData, testCase.expectedWorkbookData) {
				t.Log("expected: \n")
				t.Logf("%s\n", testCase.expectedWorkbookData)
				t.Log("\n")
				t.Log("result: \n")
				t.Logf("%s\n", actualWorkbookData)
				t.Log("\n")
				t.Fatal("Expected workbook data to be equal")
			}
		})
	}
}

// Ensures that the cell type of all cells in each column across all sheets matches the provided header types
// in each corresponding sheet
func verifyCellTypesInColumnMatchHeaderType(t *testing.T, workbookCellTypes [][][]CellType, headerMetadata [][]*StreamingCellMetadata, workbookData [][][]string) {

	numSheets := len(workbookCellTypes)
	numHeaders := len(headerMetadata)
	if numSheets != numHeaders {
		t.Fatalf("Number of sheets in workbook: %d not equal to number of sheet headers: %d", numSheets, numHeaders)
	}

	for sheetI, headers := range headerMetadata {
		var sanitizedHeaders []CellType
		for _, header := range headers {
			if header == (*StreamingCellMetadata)(nil) || header.cellType == CellTypeString {
				sanitizedHeaders = append(sanitizedHeaders, CellTypeInline)
			} else {
				sanitizedHeaders = append(sanitizedHeaders, header.cellType)
			}
		}

		sheet := workbookCellTypes[sheetI]
		// Skip header row
		for rowI, row := range sheet[1:] {
			if len(row) != len(headers) {
				t.Fatalf("Number of cells in row: %d not equal number of headers; %d", len(row), len(headers))
			}
			for colI, cellType := range row {
				headerTypeForCol := sanitizedHeaders[colI]
				if cellType != headerTypeForCol.fallbackTo(workbookData[sheetI][rowI+1][colI], CellTypeInline) {
					t.Fatalf("Cell type %d in row: %d and col: %d does not match header type: %d for this col in sheet: %d",
						cellType, rowI, colI, headerTypeForCol, sheetI)
				}
			}
		}
	}

}

// The purpose of TestXlsxStyleBehavior is to ensure that initMaxStyleId has the correct starting value
// and that the logic in AddSheet() that predicts Style IDs is correct.
func TestXlsxStyleBehavior(t *testing.T) {
	file := NewFile()
	sheet, err := file.AddSheet("Sheet 1")
	if err != nil {
		t.Fatal(err)
	}
	row := sheet.AddRow()
	rowData := []string{"testing", "1", "2", "3"}
	if count := row.WriteSlice(&rowData, -1); count != len(rowData) {
		t.Fatal("not enough cells written")
	}
	parts, err := file.MarshallParts()
	styleSheet, ok := parts["xl/styles.xml"]
	if !ok {
		t.Fatal("no style sheet")
	}
	// Created an XLSX file with only the default style.
	// This means the library adds a style by default, but no others are created
	if !strings.Contains(styleSheet, fmt.Sprintf(`<cellXfs count="%d">`, initMaxStyleId)) {
		t.Fatal("Expected sheet to have one style")
	}

	file = NewFile()
	sheet, err = file.AddSheet("Sheet 1")
	if err != nil {
		t.Fatal(err)
	}
	row = sheet.AddRow()
	rowData = []string{"testing", "1", "2", "3", "4"}
	if count := row.WriteSlice(&rowData, -1); count != len(rowData) {
		t.Fatal("not enough cells written")
	}
	sheet.SetType(0, 4, CellTypeString)
	sheet.SetType(3, 3, CellTypeNumeric)
	parts, err = file.MarshallParts()
	styleSheet, ok = parts["xl/styles.xml"]
	if !ok {
		t.Fatal("no style sheet")
	}
	// Created an XLSX file with two distinct cell types, which
	// should create two new styles.  The same cell type was added
	// three times, this should be coalesced into the same style
	// rather than recreating the style. This XLSX stream library
	// depends on this behaviour when predicting the next style
	// id.
	if !strings.Contains(styleSheet, fmt.Sprintf(`<cellXfs count="%d">`, initMaxStyleId+2)) {
		t.Fatal("Expected sheet to have four styles")
	}
}

// writeStreamFile will write the file using this stream package
func writeStreamFile(filePath string, fileBuffer io.Writer, sheetNames []string, workbookData [][][]string, headerTypes [][]*CellType, shouldMakeRealFiles bool) error {
	var file *StreamFileBuilder
	var err error
	if shouldMakeRealFiles {
		file, err = NewStreamFileBuilderForPath(filePath)
		if err != nil {
			return err
		}
	} else {
		file = NewStreamFileBuilder(fileBuffer)
	}
	for i, sheetName := range sheetNames {
		var sheetHeaderTypes []*CellType
		if i < len(headerTypes) {
			sheetHeaderTypes = headerTypes[i]
		}
		err := file.AddSheet(sheetName, sheetHeaderTypes)
		if err != nil {
			return err
		}
	}
	streamFile, err := file.Build()
	if err != nil {
		return err
	}
	for i, sheetData := range workbookData {
		if i != 0 {
			err = streamFile.NextSheet()
			if err != nil {
				return err
			}
		}
		for _, row := range sheetData {
			err = streamFile.Write(row)
			if err != nil {
				return err
			}
		}
	}
	err = streamFile.Close()
	if err != nil {
		return err
	}
	return nil
}

// writeStreamFileWithDefaultMetadata is the same thing as writeStreamFile but with headerMetadata instead of headerTypes
func writeStreamFileWithDefaultMetadata(filePath string, fileBuffer io.Writer, sheetNames []string, workbookData [][][]string, headerMetadata [][]*StreamingCellMetadata, shouldMakeRealFiles bool) error {
	var file *StreamFileBuilder
	var err error
	if shouldMakeRealFiles {
		file, err = NewStreamFileBuilderForPath(filePath)
		if err != nil {
			return err
		}
	} else {
		file = NewStreamFileBuilder(fileBuffer)
	}

	for i, sheetName := range sheetNames {
		var sheetHeaderTypes []*StreamingCellMetadata
		if i < len(headerMetadata) {
			sheetHeaderTypes = headerMetadata[i]
		}
		err := file.AddSheetWithDefaultColumnMetadata(sheetName, sheetHeaderTypes)
		if err != nil {
			return err
		}
	}
	streamFile, err := file.Build()
	if err != nil {
		return err
	}
	for i, sheetData := range workbookData {
		if i != 0 {
			err = streamFile.NextSheet()
			if err != nil {
				return err
			}
		}

		for _, row := range sheetData {
			err = streamFile.WriteWithColumnDefaultMetadata(row)
			if err != nil {
				return err
			}
		}
	}
	err = streamFile.Close()
	if err != nil {
		return err
	}
	return nil
}

// readXLSXFile will read the file using the xlsx package.
func readXLSXFile(t *testing.T, filePath string, fileBuffer io.ReaderAt, size int64, shouldMakeRealFiles bool) ([]string, [][][]string, [][][]CellType) {
	var readFile *File
	var err error
	if shouldMakeRealFiles {
		readFile, err = OpenFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		readFile, err = OpenReaderAt(fileBuffer, size)
		if err != nil {
			t.Fatal(err)
		}
	}

	var actualWorkbookData [][][]string
	var workbookCellTypes [][][]CellType
	var sheetNames []string
	for _, sheet := range readFile.Sheets {
		sheetData := [][]string{}
		sheetCellTypes := [][]CellType{}
		for _, row := range sheet.Rows {
			data := []string{}
			cellTypes := []CellType{}
			for _, cell := range row.Cells {
				str, err := cell.FormattedValue()
				if err != nil {
					t.Fatal(err)
				}
				data = append(data, str)
				cellTypes = append(cellTypes, cell.Type())
			}
			sheetData = append(sheetData, data)
			sheetCellTypes = append(sheetCellTypes, cellTypes)
		}
		sheetNames = append(sheetNames, sheet.Name)
		actualWorkbookData = append(actualWorkbookData, sheetData)
		workbookCellTypes = append(workbookCellTypes, sheetCellTypes)
	}
	return sheetNames, actualWorkbookData, workbookCellTypes
}

func checkForAutoFilterTag(filePath string, fileBuffer io.ReaderAt, size int64, shouldMakeRealFiles bool) (bool, error) {
	var readFile *File
	var err error
	if shouldMakeRealFiles {
		readFile, err = OpenFile(filePath)
		if err != nil {
			return false, err
		}
	} else {
		readFile, err = OpenReaderAt(fileBuffer, size)
		if err != nil {
			return false, err
		}
	}

	for _, sheet := range readFile.Sheets {
		if sheet.AutoFilter == nil {
			return false, nil
		}
	}
	return true, nil
}

func TestAddAutoFilters(t *testing.T) {
	c := qt.New(t)
	sheetNames := []string{
		"Sheet1",
	}
	workbookData := [][][]string{
		{
			{"Filter 1", "Filter 2"},
			{"123", "125"},
			{"123", "125"},
			{"123", "125"},
			{"125", "123"},
			{"125", "123"},
			{"125", "123"},
		},
	}
	var headerTypes [][]*CellType

	var file *StreamFileBuilder
	var err error
	filePath := "Workbook_autoFilters.xlsx"
	buffer := bytes.NewBuffer(nil)

	if TestsShouldMakeRealFiles {
		file, err = NewStreamFileBuilderForPath(filePath)
		if err != nil {
			c.Fatal(err)
		}
	} else {
		file = NewStreamFileBuilder(buffer)
	}

	for i, sheetName := range sheetNames {
		var sheetHeaderTypes []*CellType
		if i < len(headerTypes) {
			sheetHeaderTypes = headerTypes[i]
		}
		err := file.AddSheetWithAutoFilters(sheetName, sheetHeaderTypes)
		if err != nil {
			c.Fatal(err)
		}
	}
	streamFile, err := file.Build()
	if err != nil {
		c.Fatal(err)
	}
	for i, sheetData := range workbookData {
		if i != 0 {
			err = streamFile.NextSheet()
			if err != nil {
				c.Fatal(err)
			}
		}
		for _, row := range sheetData {
			err = streamFile.Write(row)
			if err != nil {
				c.Fatal(err)
			}
		}
	}
	err = streamFile.Close()
	if err != nil {
		c.Fatal(err)
	}

	// read the file back with the xlsx package
	var bufReader *bytes.Reader
	var size int64
	if !TestsShouldMakeRealFiles {
		bufReader = bytes.NewReader(buffer.Bytes())
		size = bufReader.Size()
	}
	actualSheetNames, actualWorkbookData, _ := readXLSXFile(t, filePath, bufReader, size, TestsShouldMakeRealFiles)
	// check if data was able to be read correctly
	c.Assert(actualSheetNames, qt.DeepEquals, sheetNames)
	c.Assert(actualWorkbookData, qt.DeepEquals, workbookData)

	result, err := checkForAutoFilterTag(filePath, bufReader, size, TestsShouldMakeRealFiles)
	if err != nil {
		c.Fatal(err)
	}
	if result == false {
		c.Fatal("No autoFilter added")
	}
}

func (s *StreamSuite) TestAddSheetErrorsAfterBuild(t *C) {
	file := NewStreamFileBuilder(bytes.NewBuffer(nil))

	err := file.AddSheet("Sheet1", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = file.AddSheet("Sheet2", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Build()
	if err != nil {
		t.Fatal(err)
	}
	err = file.AddSheet("Sheet3", nil)
	if err != BuiltStreamFileBuilderError {
		t.Fatal(err)
	}
}

func (s *StreamSuite) TestBuildErrorsAfterBuild(t *C) {
	file := NewStreamFileBuilder(bytes.NewBuffer(nil))

	err := file.AddSheet("Sheet1", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = file.AddSheet("Sheet2", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = file.Build()
	if err != nil {
		t.Fatal(err)
	}
	_, err = file.Build()
	if err != BuiltStreamFileBuilderError {
		t.Fatal(err)
	}
}

func TestCloseWithNothingWrittenToSheets(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	file := NewStreamFileBuilder(buffer)

	sheetNames := []string{"Sheet1", "Sheet2"}
	expectedWorkbookData := [][][]string{{}, {}}
	err := file.AddSheet(sheetNames[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	err = file.AddSheet(sheetNames[1], nil)
	if err != nil {
		t.Fatal(err)
	}

	stream, err := file.Build()
	if err != nil {
		t.Fatal(err)
	}
	err = stream.Close()
	if err != nil {
		t.Fatal(err)
	}
	bufReader := bytes.NewReader(buffer.Bytes())
	size := bufReader.Size()

	actualSheetNames, actualWorkbookData, _ := readXLSXFile(t, "", bufReader, size, false)
	// check if data was able to be read correctly
	if !reflect.DeepEqual(actualSheetNames, sheetNames) {
		t.Fatal("Expected sheet names to be equal")
	}
	if !reflect.DeepEqual(actualWorkbookData, expectedWorkbookData) {
		t.Logf("Expected:\n%s\n\n", expectedWorkbookData)
		t.Logf("Actual:\n%s\n\n", actualWorkbookData)
		t.Fatal("Expected workbook data to be equal")
	}
}

func TestMergeCells(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	fileBuilder := NewStreamFileBuilder(buffer)
	cellTypes := []*CellType{nil, nil, nil, nil, nil}
	if err := fileBuilder.AddSheet("Sheet1", cellTypes); err != nil {
		t.Fatal(err)
	}

	streamFile, err := fileBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}

	records := [][]string{
		{"Привет", "Hola", "Hi", "Hallo", "Bonjour"},
		{"Дорогой", "Querido", "Dear", "Lieber", "Cher"},
		{"Друг", "Amigo", "Friend", "Freund", "Ami"},
	}
	if err = streamFile.WriteAll(records); err != nil {
		t.Fatal(err)
	}

	streamFile.AddMergeCells(1, 1, 2, 3)
	if streamFile.currentSheet.mergeCells[0] != "B2:D3" {
		t.Error("Incorrect merge cell ref")
	}

	if err = streamFile.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := OpenBinary(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	row := file.Sheets[0].Rows[1]
	cell := row.Cells[1]
	// Two cells are added horizontally and one vertically.
	if cell.HMerge != 2 || cell.VMerge != 1 {
		fmt.Println(cell.HMerge, cell.VMerge)
		t.Error("Incorrect merge cell values")
	}
}
