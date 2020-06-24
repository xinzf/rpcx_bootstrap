package bootstrap

import (
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	PanicLevel LogLevel = "panic"
	FatalLevel LogLevel = "fatal"
)

func (l LogLevel) transform() zapcore.Level {
	mp := make(map[LogLevel]zapcore.Level)
	{
		mp[DebugLevel] = zapcore.DebugLevel
		mp[InfoLevel] = zapcore.InfoLevel
		mp[WarnLevel] = zapcore.WarnLevel
		mp[ErrorLevel] = zapcore.ErrorLevel
		mp[PanicLevel] = zapcore.PanicLevel
		mp[FatalLevel] = zap.FatalLevel
	}

	return mp[l]
}

type LogType string

const (
	LogJson LogType = "json"
	LogText LogType = "text"
)

type loggerConfig struct {
	Filename string   `mapstructure:"filename"` //日志保存路径
	LogLevel LogLevel `mapstructure:"level"`    //日志记录级别
	MaxSize  int      `mapstructure:"max_size"` //日志分割的尺寸 MB
	MaxAge   int      `mapstructure:"max_age"`  //分割日志保存的时间 day
	LogType  LogType  `mapstructure:"log_type"` //日志类型,普通 或 json
}

var encoderConfig = zapcore.EncoderConfig{
	// Keys can be anything except the empty string.
	TimeKey:       "time",
	LevelKey:      "level",
	NameKey:       "flag",
	CallerKey:     "file",
	MessageKey:    "msg",
	StacktraceKey: "stack",
	LineEnding:    zapcore.DefaultLineEnding,
	EncodeLevel:   zapcore.CapitalLevelEncoder,
	//EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006/01/02 15:04:05"))
	},
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

//默认参数
const (
	defaultLogFilename   string   = "./log/default.log" //日志保存路径 //需要设置程序当前运行路径
	defaultLogLevel      LogLevel = DebugLevel          //日志记录级别
	defaultLogMaxSize    int      = 512                 //日志分割的尺寸 MB
	defaultLogMaxAge     int      = 30                  //分割日志保存的时间 day
	defaultLogStacktrace LogLevel = PanicLevel          //记录堆栈的级别
	//defaultLogIsStdOut   bool     = true                //是否标准输出console输出
	defaultLogProjectKey string = "service" //
	defaultLogType              = LogText
)

var Logger *logger

type logger struct {
	lg   *zap.SugaredLogger
	atom zap.AtomicLevel
}

func (l *logger) init() error {
	if Config.Logger.Filename == "" {
		Config.Logger.Filename = defaultLogFilename
	}
	if Config.Logger.LogLevel == "" {
		Config.Logger.LogLevel = defaultLogLevel
	}
	if Config.Logger.MaxSize == 0 {
		Config.Logger.MaxSize = defaultLogMaxSize
	}
	if Config.Logger.MaxAge == 0 {
		Config.Logger.MaxAge = defaultLogMaxAge
	}
	if Config.Logger.LogType == "" {
		Config.Logger.LogType = defaultLogType
	}

	var writers = []zapcore.WriteSyncer{}
	writers = append(writers, os.Stdout)
	osfileout := zapcore.AddSync(&lumberjack.Logger{
		Filename:   Config.Logger.Filename,
		MaxAge:     Config.Logger.MaxAge,
		MaxBackups: 3,
		MaxSize:    Config.Logger.MaxSize,
		LocalTime:  true,
		Compress:   false,
	})

	writers = append(writers, osfileout)
	w := zapcore.NewMultiWriteSyncer(writers...)

	atom := zap.NewAtomicLevel()
	atom.SetLevel(Config.Logger.LogLevel.transform())

	var enc zapcore.Encoder
	if Config.Logger.LogType == LogText {
		enc = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(enc, w, atom)

	lg := zap.New(
		core,
		zap.AddStacktrace(defaultLogStacktrace.transform()),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	lg = lg.With(zap.String(defaultLogProjectKey, Config.Server.Name))

	Logger = &logger{
		lg:   lg.Sugar(),
		atom: atom,
	}
	return nil
}

func (l *logger) Sync() {
	l.lg.Sync()
}

func (l *logger) SetLogLevel(level LogLevel) {
	l.atom.SetLevel(level.transform())
}
func (l *logger) Debug(keysAndValues ...interface{}) {
	l.lg.Debugw("", logger{}.coupArray(keysAndValues)...)
}
func (l *logger) Info(keysAndValues ...interface{}) {
	l.lg.Infow("", logger{}.coupArray(keysAndValues)...)
}
func (l *logger) Warn(keysAndValues ...interface{}) {
	l.lg.Warnw("", logger{}.coupArray(keysAndValues)...)
}
func (l *logger) Error(keysAndValues ...interface{}) {
	l.lg.Errorw("", logger{}.coupArray(keysAndValues)...)
}
func (l *logger) Panic(keysAndValues ...interface{}) {
	l.lg.Panicw("", logger{}.coupArray(keysAndValues)...)
}
func (l *logger) Fatal(keysAndValues ...interface{}) {
	l.lg.Fatalw("", logger{}.coupArray(keysAndValues)...)
}

func (l *logger) Dump(keysAndValues ...interface{}) {
	arr := logger{}.coupArray(keysAndValues)
	for k, v := range arr {
		if k%2 == 0 {
			arr[k] = v
		} else {
			arr[k] = strings.Replace(spew.Sdump(v), "\n", "", -1)
		}
	}
	l.lg.Debugw("Dump", arr...)
}

func (l *logger) Print(v ...interface{}) {
	l.Info(v)
}

//拼接完整的数组
func (logger) coupArray(kv []interface{}) []interface{} {
	if len(kv)%2 != 0 {
		kv = append(kv, kv[len(kv)-1])
		kv[len(kv)-2] = "default"
	}
	return kv
}
