package mcp

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xuri/excelize/v2"
	"strings"
)

func Exec() {
	// 创建MCP服务器
	s := server.NewMCPServer("Excel Assistant", "1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// 注册Excel读取资源模板
	excelResource := mcp.NewResourceTemplate(
		"excel://{file}/sheet/{sheet}/range/{range}",
		"Excel Data",
		mcp.WithTemplateDescription("Read data from Excel spreadsheet"),
		mcp.WithTemplateMIMEType("application/json"),
	)
	s.AddResourceTemplate(excelResource, excelResourceHandler)

	mcp.NewTool("read_excel",
		mcp.WithDescription("Read data from Excel file"),
		mcp.WithString("file", mcp.Required(),
			mcp.Description("Path to Excel file"),
			mcp.Pattern(`.*\.xlsx$`)),
		mcp.WithString("sheet", mcp.Required(),
			mcp.Description("Worksheet name")),
		mcp.WithString("range", mcp.Required(),
			mcp.Description("Cell range to read")),
		mcp.WithString("output", mcp.Required(),
			mcp.Description("Output format")),
	)
	// 注册Excel工具
	excelTools := []struct {
		name        string
		description string
		handler     server.ToolHandlerFunc
	}{
		{
			"read_excel",
			"Read data from Excel file",
			readExcelHandler,
		},
		{
			"write_excel",
			"Write data to Excel file",
			writeExcelHandler,
		},
		//{
		//	"add_row",
		//	"Add new row to Excel sheet",
		//	addRowHandler,
		//},
	}

	for _, tool := range excelTools {
		t := mcp.NewTool(tool.name,
			mcp.WithDescription(tool.description),
			mcp.WithString("file", mcp.Required(),
				mcp.Description("Path to Excel file"),
				mcp.Pattern(`.*\.xlsx$`)),
			mcp.WithString("sheet", mcp.Required(),
				mcp.Description("Worksheet name")),
		)
		s.AddTool(t, tool.handler)
	}

	// 启动服务器
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// Excel资源处理器
func excelResourceHandler(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	params := parseExcelURI(req.Params.URI)
	if params == nil {
		return nil, fmt.Errorf("invalid excel URI format")
	}

	f, err := excelize.OpenFile(params.file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(params.sheet)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	headers := rows[0]
	for _, row := range rows[1:] {
		item := make(map[string]interface{})
		for i, cell := range row {
			if i >= len(headers) {
				break
			}
			item[headers[i]] = cell
		}
		result = append(result, item)
	}

	// 修正返回类型为ResourceContents
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     fmt.Sprintf("%v", result),
		},
	}, nil
}

// 读取Excel工具处理
func readExcelHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	file := args["file"].(string)
	sheet := args["sheet"].(string)

	f, err := excelize.OpenFile(file)
	if err != nil {
		return mcp.NewToolResultError("Failed to open file"), nil
	}
	defer f.Close()

	// 获取所有单元格内容
	rows, err := f.GetRows(sheet)
	if err != nil {
		return mcp.NewToolResultError("Failed to read sheet"), nil
	}

	// 使用 NewToolResultText 代替 NewToolResultJSON
	return mcp.NewToolResultText(fmt.Sprintf("%v", rows)), nil
}

// 写入Excel工具处理
func writeExcelHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.Params.Arguments
	file := args["file"].(string)
	sheet := args["sheet"].(string)
	data := args["data"].(map[string]interface{})

	f, err := excelize.OpenFile(file)
	if err != nil {
		f = excelize.NewFile()
		f.NewSheet(sheet)
	}

	// 获取当前行数
	rows, _ := f.GetRows(sheet)
	rowNum := len(rows) + 1

	// 写入数据
	col := 'A'
	for _, value := range data {
		cell := fmt.Sprintf("%c%d", col, rowNum)
		f.SetCellValue(sheet, cell, value)
		col++
	}

	if err := f.SaveAs(file); err != nil {
		return mcp.NewToolResultError("Failed to save file"), nil
	}

	return mcp.NewToolResultText("Data written successfully"), nil
}

// Excel URI参数结构
type excelParams struct {
	file  string
	sheet string
	Range string
}

// 解析Excel URI参数
func parseExcelURI(uri string) *excelParams {
	// 解析格式：excel://{file}/sheet/{sheet}/range/{range}
	parts := strings.Split(uri, "/")
	if len(parts) < 7 {
		return nil
	}
	return &excelParams{
		file:  parts[2],
		sheet: parts[4],
		Range: parts[6],
	}
}
