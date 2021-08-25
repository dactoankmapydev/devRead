package handle_log

import (
	"time"

	"github.com/natefinch/lumberjack"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:   "message",
		TimeKey:      "time",
		LevelKey:     "level",
		CallerKey:    "caller",
		EncodeLevel:  CustomLevelEncoder,         //Format cách hiển thị level log
		EncodeTime:   SyslogTimeEncoder,          //Format hiển thị thời điểm log
		EncodeCaller: zapcore.ShortCallerEncoder, //Format dòng code bắt đầu log
	})
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func logErrorWriter() (zapcore.WriteSyncer, error) {
	logErrorPath := "./error.log"
	_, _, _ = zap.Open(logErrorPath)

	return zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(&lumberjack.Logger{
			Filename: logErrorPath,
			MaxSize:  20, // megabytes
			MaxAge:   3,  // days
		}),
	), nil
}

func logInfoWriter() (zapcore.WriteSyncer, error) {
	logInfoPath := "./info.log"
	_, _, _ = zap.Open(logInfoPath)

	return zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(&lumberjack.Logger{
			Filename: logInfoPath,
			MaxSize:  20,
			MaxAge:   3,
		}),
	), nil
}

// Write log to file by level log and console
func WriteLog() (*zap.Logger, error) {
	highWriteSyncer, errorWriter := logErrorWriter()
	if errorWriter != nil {
		return nil, errorWriter
	}
	lowWriteSyncer, errorInfoWriter := logInfoWriter()
	if errorInfoWriter != nil {
		return nil, errorInfoWriter
	}

	encoder := getEncoder()

	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	lowCore := zapcore.NewCore(encoder, lowWriteSyncer, lowPriority)
	highCore := zapcore.NewCore(encoder, highWriteSyncer, highPriority)

	logger := zap.New(zapcore.NewTee(lowCore, highCore), zap.AddCaller())
	return logger, nil
}