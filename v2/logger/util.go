package logger

import (
	"io"
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/segmentio/encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapLogger(level zapcore.Level, writers ...io.Writer) (l *zap.Logger) {
	zapWriters := make([]zapcore.WriteSyncer, 0)
	for _, writer := range writers {
		if writer == nil {
			continue
		}
		zapWriters = append(zapWriters, zapcore.AddSync(writer))
	}

	core := zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(zapWriters...), zapcore.Level(level))
	l = zap.New(core)
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "xtime",
		MessageKey:     "x",
		EncodeDuration: millisDurationEncoder,
		EncodeTime:     timeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.999"))
}

func millisDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Nanoseconds() / 1000000)
}

func formatLogs(msg string, mask bool, fields ...Field) (logRecord []zap.Field) {
	logRecord = append(logRecord, zap.String("message", msg))
	for _, field := range fields {
		logRecord = append(logRecord, formatLog(field.Key, field.Val, mask))
	}
	return
}

func formatLog(key string, msg interface{}, mask bool) (logRecord zap.Field) {
	if msg == nil {
		logRecord = zap.Any(key, struct{}{})
		return
	}

	p, ok := msg.(proto.Message)
	if ok {
		b, err := json.Marshal(p)
		if err != nil {
			logRecord = zap.Any(key, p.String())
			return
		}

		var data interface{}
		if err = json.Unmarshal(b, &data); err != nil {
			logRecord = zap.Any(key, p.String())
			return
		}

		logRecord = zap.Any(key, data)
		return
	}

	if str, ok := msg.(string); ok {
		var data interface{}
		if err := json.Unmarshal([]byte(str), &data); err != nil {
			logRecord = zap.String(key, str)
			return
		}

		logRecord = zap.Any(key, data)
		return
	}

	if !mask {
		logRecord = zap.Any(key, msg)
		return
	}

	switch reflect.ValueOf(msg).Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Struct, reflect.Map:
		msgMasking := masking(msg)
		if convert, ok := msgMasking.(reflect.Value); ok {
			logRecord = zap.Any(key, convert.Interface())
			return
		}
	}

	logRecord = zap.Any(key, msg)
	return
}

func masking(data interface{}) interface{} {
	original := reflect.ValueOf(data)
	altered := reflect.New(original.Type()).Elem()

	switch original.Kind() {
	case reflect.Ptr:
		if !isNil(original) {
			elem := original.Elem()
			switch elem.Kind() {
			case reflect.Struct, reflect.Interface, reflect.Ptr:
				altered.Set(masking(elem.Interface()).(reflect.Value).Addr())
			case reflect.Slice:
				ptr := reflect.New(elem.Type())
				ptr.Elem().Set(maskSlice(elem))
				altered.Set(ptr)
			case reflect.Map:
				ptr := reflect.New(elem.Type())
				ptr.Elem().Set(maskMap(elem))
				altered.Set(ptr)
			default:
				altered.Set(elem.Addr())
			}
		}
	case reflect.Slice:
		altered = maskSlice(original)
	case reflect.Map:
		altered = maskMap(original)
	case reflect.Struct:
		if original.Type() == TypeTime {
			altered.Set(original)
			return altered
		}
		for i := 0; i < original.NumField(); i++ {
			field := original.Field(i)
			switch field.Kind() {
			case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
				if altered.Field(i).CanSet() && !isNil(field) {
					_, hasMaskTag := original.Type().Field(i).Tag.Lookup(maskTag)
					if hasMaskTag {
						altered.Field(i).Set(maskValue(field))
					} else if field.Type() == TypeSliceOfBytes {
						altered.Field(i).Set(original.Field(i))
					} else {
						altered.Field(i).Set(masking(field.Interface()).(reflect.Value))
					}
				}
			default:
				if _, ok := original.Type().Field(i).Tag.Lookup(maskTag); ok {
					if original.Field(i).Kind() == reflect.String && original.Field(i).Len() > 0 {
						altered.Field(i).SetString(stringMask)
					} else if altered.Field(i).CanSet() {
						altered.Field(i).Set(original.Field(i))
					}
				} else if altered.Field(i).CanSet() {
					altered.Field(i).Set(original.Field(i))
				}
			}
		}
	default:
		altered.Set(original)
	}

	return altered
}

func maskSlice(elem reflect.Value) reflect.Value {
	altered := reflect.MakeSlice(elem.Type(), elem.Len(), elem.Len())
	for i := 0; i < elem.Len(); i++ {
		value := elem.Index(i)
		switch value.Kind() {
		case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
			if !isNil(value) {
				altered.Index(i).Set(masking(value.Interface()).(reflect.Value))
			}
		default:
			altered.Index(i).Set(value)
		}
	}
	return altered
}

func maskMap(elem reflect.Value) reflect.Value {
	altered := reflect.MakeMapWithSize(elem.Type(), len(elem.MapKeys()))
	mapRange := elem.MapRange()
	for mapRange.Next() {
		switch mapRange.Value().Kind() {
		case reflect.Struct, reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr:
			if !isNil(mapRange.Value()) {
				altered.SetMapIndex(
					mapRange.Key(),
					masking(mapRange.Value().Interface()).(reflect.Value),
				)
			}
		default:
			altered.SetMapIndex(mapRange.Key(), mapRange.Value())
		}
	}
	return altered
}

func isNil(elem reflect.Value) bool {
	return elem.Interface() == nil ||
		(reflect.ValueOf(elem.Interface()).Kind() == reflect.Ptr && reflect.ValueOf(elem.Interface()).IsNil())
}

func maskValue(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Map &&
		v.Type().Key().Kind() == reflect.String &&
		v.Type().Elem().Kind() == reflect.String {
		masked := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, k := range v.MapKeys() {
			masked.SetMapIndex(k, reflect.ValueOf(stringMask))
		}
		return masked
	}
	return reflect.Zero(v.Type())
}
