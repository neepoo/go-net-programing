package chapter13

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var encoderCfg = zapcore.EncoderConfig{
	MessageKey: "msg",
	LevelKey:   "level",
	//TimeKey:          "",
	NameKey:   "name",
	CallerKey: "caller",
	//FunctionKey:      "",
	//StacktraceKey:    "",
	//LineEnding:       "",
	EncodeLevel: zapcore.LowercaseLevelEncoder,
	//EncodeTime:       nil,
	//EncodeDuration:   nil,
	EncodeCaller: zapcore.ShortCallerEncoder,
	//EncodeName:       nil,
	//ConsoleSeparator: "",
}

func Example_zapJSon() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		),
		zap.AddCaller(),
		zap.Fields(
			// zap.String("version", runtime.Version()),
			zap.String("version", "go1.17.3"),
		),
	)

	defer zl.Sync()

	example := zl.Named("example")
	example.Debug("test debug message")
	example.Info("test info message")
	example.Sync()
	//a1 := example.Named("a1")
	//a1.Info("a1 test info message")

	// Output:
	// {"level":"debug","name":"example","caller":"chapter13/zap_test.go:43","msg":"test debug message","version":"go1.17.3"}
	// {"level":"info","name":"example","caller":"chapter13/zap_test.go:44","msg":"test info message","version":"go1.17.3"}
}

func Example_zapConsole() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	console := zl.Named("[console]")

	console.Info("this is logged by the logger")
	console.Debug("this is below the logger's threshold and won't log")
	console.Error("this is also logged by the logger")
	// Output:
	// info	[console]	this is logged by the logger
	// error	[console]	this is also logged by the logger
}

func Example_zapInfoFileDebugConsole() {
	logFile := new(bytes.Buffer)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(zapcore.AddSync(logFile)),
		zapcore.InfoLevel)
	zl := zap.New(
		core,
	)
	defer func() { _ = zl.Sync() }()
	zl.Debug("this is below the logger's threshold and won't log")
	zl.Error("this is logged by the logger")

	zl = zl.WithOptions(
		zap.WrapCore(
			func(core zapcore.Core) zapcore.Core {
				// 这里的这个core就是line78的core
				ucEncoderCfg := encoderCfg
				ucEncoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
				// Lastly, you use the
				//zapcore.NewTee function, which is like the io.MultiWriter function, to return
				//a zap.Core that writes to multiple cores
				return zapcore.NewTee(
					core,
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(ucEncoderCfg),
						zapcore.Lock(os.Stdout),
						zapcore.DebugLevel),
				)
			},
		),
		//zap.Fields(zap.String("greeting", "hello")),
	)

	fmt.Println("standard output:")
	zl.Debug("this is only logged as condole encoding")
	zl.Info("this is logged as console encoding and json")
	fmt.Print("\nlog file contents:\n", logFile.String())
	// Output:
	// standard output:
	// DEBUG	this is only logged as condole encoding
	// INFO	this is logged as console encoding and json
	//
	// log file contents:
	// {"level":"error","msg":"this is logged by the logger"}
	// {"level":"info","msg":"this is logged as console encoding and json"}
}

func Example_zapSampling() {
	zl := zap.New(
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.Lock(os.Stdout),
				zapcore.DebugLevel),
			time.Second, // 采样间隔
			1,           // 复制的日志记录的上限阈值
			3,           // 第n个复制的日志会恢复记录
		),
	)
	defer zl.Sync()
	// you are logging the first log entry, and then every third duplicate
	//log entry that the logger receives in a one-second interval. Once the interval
	//elapses, the logger starts over and logs the first entry, then every third dupli-
	//cate for the remainder of the one-second interval.
	for i := 0; i < 10; i++ {
		if i == 5 {
			//  ensure that the sample logger starts logging anew dur-
			//ing the next one-second interval
			time.Sleep(time.Second)
		}
		zl.Debug(fmt.Sprintf("%d", i))
		zl.Debug("debug message")
	}
	// Output:
	// {"level":"debug","msg":"0"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"1"}
	// {"level":"debug","msg":"2"}
	// {"level":"debug","msg":"3"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"4"}
	// {"level":"debug","msg":"5"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"6"}
	// {"level":"debug","msg":"7"}
	// {"level":"debug","msg":"8"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"9"}
}

func Example_zapDynamicDebugging() {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// will watch this semaphore file
	debugLevelFile := filepath.Join(tempDir, "level.debug")
	atomicLevel := zap.NewAtomicLevel()

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atomicLevel,
		),
	)

	defer zl.Sync()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zl.Fatal("creating watcher error", zap.Error(err))
	}

	defer watcher.Close()

	err = watcher.Add(tempDir)

	ready := make(chan struct{})

	go func() {
		defer close(ready)

		originalLevel := atomicLevel.Level()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name == debugLevelFile {
					switch {
					case event.Op&fsnotify.Create == fsnotify.Create:
						atomicLevel.SetLevel(zapcore.DebugLevel)
						ready <- struct{}{}

					case event.Op&fsnotify.Remove == fsnotify.Remove:
						atomicLevel.SetLevel(originalLevel)
						ready <- struct{}{}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zl.Error(err.Error())
			}
		}
	}()

	// shouldn't output
	zl.Debug("this is below the logger's threshold")

	df, err := os.Create(debugLevelFile)
	if err != nil {
		zl.Fatal(err.Error())
	}
	err = df.Close()
	if err != nil {
		zl.Fatal(err.Error())
	}
	<-ready
	// now debug level

	// should output
	zl.Debug("this is now at the logger's threshold")
	err = os.Remove(debugLevelFile)
	if err != nil {
		zl.Fatal(err.Error())
	}
	<-ready
	// info level

	// shouldn't output
	zl.Debug("this is below the logger's threshold again")
	// output
	zl.Info("this is at the logger's current threshold")
	// Output:
	// {"level":"debug","msg":"this is now at the logger's threshold"}
	// {"level":"info","msg":"this is at the logger's current threshold"}
}

// TestZapLogRotation 使用第三方库来留存日志
func TestZapLogRotation(t *testing.T) {
	temp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(temp)

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(
				&lumberjack.Logger{
					Filename:   filepath.Join(temp, "debug.log"),
					Compress:   true,
					LocalTime:  true,
					MaxAge:     7,
					MaxBackups: 5,
					MaxSize:    100, // MByte
				},
			),
			zapcore.DebugLevel,
		),
	)

	defer zl.Sync()

	zl.Debug("debug message written to the log file")
}
