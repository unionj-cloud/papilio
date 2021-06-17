package ddhttp

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/unionj-cloud/go-doudou/pathutils"
	"github.com/unionj-cloud/go-doudou/stringutils"
	"github.com/unionj-cloud/go-doudou/svc/config"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Srv interface {
	// Run the service
	Run()
	// Register routes
	AddRoute(route ...Route)
	// Use middleware
	AddMiddleware(mwf ...func(http.Handler) http.Handler)
}

func newServer(router http.Handler) *http.Server {
	host := os.Getenv("SRV_HOST")
	port := config.SvcPort.Load()
	write, err := time.ParseDuration(os.Getenv("SRV_WRITETIMEOUT"))
	if err != nil {
		logrus.Warnf("Parse %s %s as time.Duration failed: %s, use default 15s instead.\n", "SRV_WRITETIMEOUT",
			os.Getenv("SRV_WRITETIMEOUT"), err.Error())
		write = 15 * time.Second
	}

	read, err := time.ParseDuration(os.Getenv("SRV_READTIMEOUT"))
	if err != nil {
		logrus.Warnf("Parse %s %s as time.Duration failed: %s, use default 15s instead.\n", "SRV_READTIMEOUT",
			os.Getenv("SRV_READTIMEOUT"), err.Error())
		read = 15 * time.Second
	}

	idle, err := time.ParseDuration(os.Getenv("SRV_IDLETIMEOUT"))
	if err != nil {
		logrus.Warnf("Parse %s %s as time.Duration failed: %s, use default 60s instead.\n", "SRV_IDLETIMEOUT",
			os.Getenv("SRV_IDLETIMEOUT"), err.Error())
		idle = 60 * time.Second
	}

	server := &http.Server{
		Addr: strings.Join([]string{host, port}, ":"),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: write,
		ReadTimeout:  read,
		IdleTimeout:  idle,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		logrus.Infof("Http server is listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			logrus.Println(err)
		}
	}()

	return server
}

func configureLogger(logger *logrus.Logger, logptr *string, level logrus.Level) *os.File {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(level)

	if logptr != nil {
		var (
			err error
			f   *os.File
		)
		logpath := *logptr
		logpath, err = pathutils.FixPath(logpath, "")
		if err != nil {
			logger.Errorln(fmt.Sprintf("%+v\n", err))
		}
		if stringutils.IsNotEmpty(logpath) {
			err = os.MkdirAll(logpath, os.ModePerm)
			if err != nil {
				logger.Errorln(err)
				return nil
			}
		}
		f, err = os.OpenFile(filepath.Join(logpath, "app.log"), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logger.Errorf("error opening file: %v\n", err)
			return nil
		}
		mw := io.MultiWriter(os.Stdout, f)
		logger.SetOutput(mw)
		return f
	}

	return nil
}

func printRoutes(routes []Route) {
	logrus.Infoln("================ Registered Routes ================")
	data := [][]string{}
	for _, r := range routes {
		data = append(data, []string{r.Name, r.Method, r.Pattern})
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Name", "Method", "Pattern"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
	rows := strings.Split(strings.TrimSpace(tableString.String()), "\n")
	for _, row := range rows {
		logrus.Infoln(row)
	}
	logrus.Infoln("===================================================")
}
