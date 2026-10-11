package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/yigger/jiezhang-backend/internal/bootstrap"
	"github.com/yigger/jiezhang-backend/internal/config"
	"github.com/yigger/jiezhang-backend/internal/service/billimport"
)

type paths []string

func (p *paths) String() string        { return strings.Join(*p, ", ") }
func (p *paths) Set(path string) error { *p = append(*p, path); return nil }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("bill-import", flag.ContinueOnError)
	flags.SetOutput(errOut)
	var files paths
	flags.Var(&files, "file", "CSV/XLSX file; repeat for multiple files")
	user := flags.Int64("user-id", 0, "operator user ID (required)")
	book := flags.Int64("book-id", 0, "target account book ID (required)")
	source := flags.String("source", "", "alipay or wechat; default detects from headers")
	mappingPath := flags.String("mapping", "", "optional JSON mapping file")
	reportPath := flags.String("report", "", "optional JSON preview/result report; keep outside source control")
	apply := flags.Bool("apply", false, "apply historical import; default previews without writes")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if len(flags.Args()) != 0 || *user <= 0 || *book <= 0 || len(files) == 0 || len(files) > 20 || (*source != "" && *source != "alipay" && *source != "wechat") {
		fmt.Fprintln(errOut, "需要 --file、正数 --user-id 和 --book-id；来源仅支持 alipay/wechat。")
		return 2
	}
	input := billimport.Input{UserID: *user, AccountBookID: *book, Apply: *apply}
	if *mappingPath != "" {
		file, err := os.Open(*mappingPath)
		if err != nil {
			fmt.Fprintln(errOut, "无法读取映射配置。")
			return 2
		}
		decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&input.Mapping)
		if err == nil {
			var extra any
			if decoder.Decode(&extra) != io.EOF {
				err = errors.New("trailing JSON")
			}
		}
		_ = file.Close()
		if err != nil {
			fmt.Fprintln(errOut, "映射配置必须是有效 JSON，字段见 README。")
			return 2
		}
	}
	var opened []*os.File
	defer func() {
		for _, f := range opened {
			_ = f.Close()
		}
	}()
	for _, path := range files {
		format := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		if format != "csv" && format != "xlsx" {
			fmt.Fprintln(errOut, "仅支持已解压的 .csv/.xlsx 文件。")
			return 2
		}
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(errOut, "无法打开账单文件。")
			return 2
		}
		opened = append(opened, f)
		input.Files = append(input.Files, billimport.File{Reader: f, Format: format, Source: *source})
	}
	var reportFile *os.File
	if *reportPath != "" {
		var err error
		reportFile, err = os.OpenFile(*reportPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			fmt.Fprintln(errOut, "无法创建报告（路径须不存在且可写）。")
			return 2
		}
		defer reportFile.Close()
	}
	service, closeDB, err := bootstrap.NewBillImporter(config.Load().MySQLDSN)
	if err != nil {
		fmt.Fprintln(errOut, err.Error())
		return 1
	}
	defer closeDB()
	report, err := service.Import(ctx, input)
	if reportFile != nil && report.Rows != nil {
		encoder := json.NewEncoder(reportFile)
		encoder.SetIndent("", "  ")
		writeErr := encoder.Encode(report)
		if writeErr == nil {
			writeErr = reportFile.Sync()
		}
		if writeErr != nil {
			fmt.Fprintln(errOut, "无法保存报告（路径须不存在且可写）。")
			if report.Applied {
				fmt.Fprintln(errOut, "账单已经提交；重复执行会检查来源订单，不会重复入账。")
			}
			return 1
		}
	}
	if report.Rows != nil {
		fmt.Fprintf(out, "可导入 %d；已导入 %d；重复 %d；跳过 %d；错误 %d；拟新建分类/资产 %d。\n", report.Ready, report.Imported, report.Duplicate, report.Skipped, report.Errors, len(report.Creations))
	}
	if err != nil {
		if errors.Is(err, billimport.ErrInvalid) {
			fmt.Fprintln(errOut, err.Error())
		} else if errors.Is(err, context.Canceled) {
			fmt.Fprintln(errOut, "导入已取消；未提交的事务已回滚。")
		} else {
			fmt.Fprintln(errOut, "数据库操作失败；检查连接权限及 20261011_statement_import.sql 迁移。")
		}
		return 1
	}
	if report.Errors > 0 {
		fmt.Fprintln(errOut, "预览存在错误；使用 --report 查看行号和原因，完善映射后重试。")
		return 1
	}
	if !report.Applied {
		fmt.Fprintln(out, "仅预览，未写数据库。检查报告后添加 --apply 执行历史补录；现有余额保持不变。")
	}
	return 0
}
