package conf

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	beego "github.com/astaxie/beego/logs"
	"github.com/sirupsen/logrus"
)

// GOPATH contains the gopath of the env
var GOPATH string

// LogDir is the log directory for logs.json
var LogDir string

// LogFile name
var LogFile string

func init() {
	GOPATH = os.Getenv("GOPATH")
	if GOPATH == "" {
		GOPATH = defaultGOPATH()
	}

	appPath := GOPATH + "/src/github.com/hjochman/docker-registry-manager/app"
	LogDir = appPath + "/logs/"
	LogFile = LogDir + "/log.json"

	// create log dir if needed
	if _, err := os.Stat(LogDir); os.IsNotExist(err) {
		if err = os.MkdirAll(LogDir, 0755); err != nil {
			logrus.Fatal(err)
		}
	}

	// Create the log file if needed
	if _, err := os.Stat(LogFile); os.IsNotExist(err) {
		if _, err = os.Create(LogFile); err != nil {
			logrus.Fatal(err)
		}
	}

	// Setup logrus with the json events log
	f, _ := os.OpenFile(LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	logrus.AddHook(NewFileHook(f, &logrus.JSONFormatter{}))

	// log to std out for docker logs
	logrus.SetOutput(os.Stdout)
	logrus.AddHook(ContextHook{})

	// Connect the beego logs to the other logs
	beego.Register("docker-registry-manager", NewBeegoHook)
	beego.SetLogger("docker-registry-manager", "")
	beego.EnableFuncCallDepth(true)
}

type registryLogger struct {
	Level int `json:"level"`
}

func NewBeegoHook() beego.Logger { return &registryLogger{Level: 0} }

// Init returns nothing since theres no additional config optios needed
func (rl *registryLogger) Init(jsonconfig string) error { return nil }

// Destroy is a empty method
func (rl *registryLogger) Destroy() {}

// Flush is a empty method
func (rl *registryLogger) Flush() {}

// WriteMsg will write the msg and level into es
func (rl *registryLogger) WriteMsg(when time.Time, msg string, level int) error {
	if level < int(logrus.GetLevel()) {
		return nil
	}

	switch level {
	// beego is reverse order
	case 0:
		logrus.Panic(msg)
	case 1:
		logrus.Panic(msg)
	case 2:
		logrus.Panic(msg)
	case 3:
		logrus.Error(msg)
	case 4:
		logrus.Warn(msg)
	case 5:
		logrus.Info(msg)
	case 6:
		logrus.Info(msg)
	case 7:
		logrus.Debug(msg)
	}
	return nil
}

type ContextHook struct{}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {

			if strings.Contains(name, "registryLogger") {
				// Remove the prefix beego attaches
				r := regexp.MustCompile(`\[[A-Z]\] \[(.*):(.*)\] (.*)`)
				message := r.FindAllStringSubmatch(entry.Message, 1)

				// add the caller as a separate field
				entry.Data["file"] = message[0][1]
				line, _ := strconv.Atoi(message[0][2])
				entry.Data["line"] = line
				entry.Message = message[0][3]
				entry.Data["source"] = "beego"
				break
			} else {
				file, line := fu.FileLine(pc[i] - 1)
				entry.Data["file"] = path.Base(file)
				entry.Data["line"] = line
				entry.Data["source"] = "app"
				break
			}
		}

	}
	return nil
}

type FileHook struct {
	writer    io.Writer
	formatter logrus.Formatter
}

func NewFileHook(w io.Writer, f logrus.Formatter) logrus.Hook {
	return FileHook{
		writer:    w,
		formatter: f,
	}
}

func (h FileHook) Fire(e *logrus.Entry) error {
	b, err := h.formatter.Format(e)
	if err != nil {
		return err
	}
	_, err = h.writer.Write(b)
	return err
}

// Levels returns all logrus levels.
func (h FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func defaultGOPATH() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	if home := os.Getenv(env); home != "" {
		def := filepath.Join(home, "go")
		if filepath.Clean(def) == filepath.Clean(runtime.GOROOT()) {
			// Don't set the default GOPATH to GOROOT,
			// as that will trigger warnings from the go tool.
			return ""
		}
		return def
	}
	return ""
}
