package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/gear"
)

// ----- Test Helpers -----
func EqualPtr(t *testing.T, a, b interface{}) {
	assert.Equal(t, reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer())
}

type GearResponse struct {
	*http.Response
}

var DefaultClient = &http.Client{}

func RequestBy(method, url string) (*GearResponse, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := DefaultClient.Do(req)
	return &GearResponse{res}, err
}
func NewRequst(method, url string) (*http.Request, error) {
	return http.NewRequest(method, url, nil)
}
func DefaultClientDo(req *http.Request) (*GearResponse, error) {
	res, err := DefaultClient.Do(req)
	return &GearResponse{res}, err
}

func TestGearLogger(t *testing.T) {
	exit = func() {} // overwrite exit function

	t.Run("Default logger", func(t *testing.T) {
		assert := assert.New(t)

		now := time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
		logger := Default()
		assert.Equal(logger.l, DebugLevel)
		assert.Equal(logger.tf, "2006-01-02T15:04:05.999Z")
		assert.Equal(logger.lf, "%s %s %s")

		var buf bytes.Buffer

		logger.Out = &buf
		logger.Emerg("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello\n"))
		buf.Reset()

		Emerg("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello1\n"))
		buf.Reset()

		logger.Alert("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "ALERT Hello\n"))
		buf.Reset()

		Alert("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "ALERT Hello1\n"))
		buf.Reset()

		logger.Crit("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "CRIT Hello\n"))
		buf.Reset()

		Crit("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "CRIT Hello1\n"))
		buf.Reset()

		logger.Err("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "ERR Hello\n"))
		buf.Reset()

		Err("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "ERR Hello1\n"))
		buf.Reset()

		logger.Warning("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "WARNING Hello\n"))
		buf.Reset()

		Warning("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "WARNING Hello1\n"))
		buf.Reset()

		logger.Notice("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "NOTICE Hello\n"))
		buf.Reset()

		Notice("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "NOTICE Hello1\n"))
		buf.Reset()

		logger.Info("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "INFO Hello\n"))
		buf.Reset()

		Info("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "INFO Hello1\n"))
		buf.Reset()

		logger.Debug("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "DEBUG Hello\n"))
		buf.Reset()

		Debug("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "DEBUG Hello1\n"))
		buf.Reset()

		assert.Panics(func() {
			logger.Panic("Hello")
		})
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello\n"))
		buf.Reset()

		assert.Panics(func() {
			Panic("Hello1")
		})
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello1\n"))
		buf.Reset()

		logger.Fatal("Hello")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello\n"))
		buf.Reset()

		Fatal("Hello1")
		assert.Equal(buf.String(), fmt.Sprintf("%s %s", now, "EMERG Hello1\n"))
		buf.Reset()

		logger.Print("Hello")
		assert.Equal(buf.String(), "Hello")
		buf.Reset()

		Print("Hello1")
		assert.Equal(buf.String(), "Hello1")
		buf.Reset()

		logger.Printf(":%s", "Hello")
		assert.Equal(buf.String(), ":Hello")
		buf.Reset()

		Printf(":%s", "Hello1")
		assert.Equal(buf.String(), ":Hello1")
		buf.Reset()

		logger.Println("Hello")
		assert.Equal(buf.String(), "Hello\n")
		buf.Reset()

		Println("Hello1")
		assert.Equal(buf.String(), "Hello1\n")
		buf.Reset()
	})

	t.Run("logger setting", func(t *testing.T) {
		assert := assert.New(t)

		logger := Default()

		var buf bytes.Buffer
		logger.Out = &buf

		assert.Panics(func() {
			var level Level = 8
			logger.SetLevel(level)
		})
		logger.SetLevel(NoticeLevel)
		logger.Info("Hello")
		assert.Equal(buf.String(), "")
		buf.Reset()
	})
}

func TestGearLoggerMiddleware(t *testing.T) {
	t.Run("Default log", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("use native color func for windows platform")
		}
		assert := assert.New(t)

		var buf bytes.Buffer
		app := gear.New()
		logger := Default()
		logger.Out = &buf
		app.UseHandler(logger)
		app.Use(func(ctx *gear.Context) error {
			log := FromCtx(ctx)
			EqualPtr(t, log, logger.FromCtx(ctx))
			return ctx.HTML(200, "OK")
		})
		srv := app.Start()
		defer srv.Close()

		res, err := RequestBy("GET", "http://"+srv.Addr().String())
		assert.Nil(err)
		assert.Equal(200, res.StatusCode)
		assert.Equal("text/html; charset=utf-8", res.Header.Get(gear.HeaderContentType))
		time.Sleep(10 * time.Millisecond)
		log := buf.String()

		assert.Contains(log, "127.0.0.1 GET / ")
		assert.Contains(log, "\x1b[32;1m200\x1b[39;22m")
		res.Body.Close()
	})

	t.Run("simple log", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("use native color func for windows platform")
		}
		assert := assert.New(t)

		var buf bytes.Buffer
		app := gear.New()
		logger := New(&buf)
		app.UseHandler(logger)
		app.Use(func(ctx *gear.Context) error {
			log := logger.FromCtx(ctx)
			log["Data"] = []int{1, 2, 3}
			return ctx.HTML(200, "OK")
		})
		srv := app.Start()
		defer srv.Close()

		res, err := RequestBy("GET", "http://"+srv.Addr().String())
		assert.Nil(err)
		assert.Equal(200, res.StatusCode)
		assert.Equal("text/html; charset=utf-8", res.Header.Get(gear.HeaderContentType))
		time.Sleep(10 * time.Millisecond)
		log := buf.String()

		assert.Contains(log, "127.0.0.1 GET / ")
		assert.Contains(log, "\x1b[32;1m200\x1b[39;22m")
		res.Body.Close()
	})

	t.Run("custom log", func(t *testing.T) {
		assert := assert.New(t)

		var buf bytes.Buffer
		app := gear.New()

		logger := New(&buf)
		logger.SetInitLog(func(log Log, ctx *gear.Context) {
			log["IP"] = ctx.IP()
			log["Method"] = ctx.Method
			log["URL"] = ctx.Req.URL.String()
			log["Start"] = time.Now()
			log["UserAgent"] = ctx.Get(gear.HeaderUserAgent)
		})
		logger.SetConsumeLog(func(log Log, _ *Logger) {
			end := time.Now()
			log["Time"] = end.Sub(log["Start"].(time.Time)) / 1e6
			delete(log, "Start")
			switch res, err := json.Marshal(log); err == nil {
			case true:
				logger.Output(end, InfoLevel, string(res))
			default:
				logger.Output(end, WarningLevel, err.Error())
			}
		})

		app.UseHandler(logger)
		app.Use(func(ctx *gear.Context) error {
			log := logger.FromCtx(ctx)
			log["Data"] = []int{1, 2, 3}
			return ctx.HTML(200, "OK")
		})
		srv := app.Start()
		defer srv.Close()

		res, err := RequestBy("GET", "http://"+srv.Addr().String())
		assert.Nil(err)
		assert.Equal(200, res.StatusCode)
		assert.Equal("text/html; charset=utf-8", res.Header.Get(gear.HeaderContentType))
		time.Sleep(10 * time.Millisecond)
		log := buf.String()
		assert.Contains(log, time.Now().UTC().Format(time.RFC3339)[0:19])
		assert.Contains(log, " INFO ")
		assert.Contains(log, `"Data":[1,2,3]`)
		assert.Contains(log, `"Method":"GET"`)
		assert.Contains(log, `"Status":200`)
		assert.Contains(log, `"UserAgent":`)
		res.Body.Close()
	})

	t.Run("Work with panic", func(t *testing.T) {
		assert := assert.New(t)

		var buf bytes.Buffer
		var errbuf bytes.Buffer

		app := gear.New()
		app.Set("AppLogger", log.New(&errbuf, "TEST: ", 0))

		logger := New(&buf)
		logger.SetInitLog(func(log Log, ctx *gear.Context) {
			log["IP"] = ctx.IP()
			log["Method"] = ctx.Method
			log["URL"] = ctx.Req.URL.String()
			log["Start"] = time.Now()
			log["UserAgent"] = ctx.Get(gear.HeaderUserAgent)
		})
		logger.SetConsumeLog(func(log Log, _ *Logger) {
			end := time.Now()
			log["Time"] = end.Sub(log["Start"].(time.Time)) / 1e6
			delete(log, "Start")
			switch res, err := json.Marshal(log); err == nil {
			case true:
				logger.Output(end, InfoLevel, string(res))
			default:
				logger.Output(end, WarningLevel, err.Error())
			}
		})

		app.UseHandler(logger)
		app.Use(func(ctx *gear.Context) (err error) {
			log := logger.FromCtx(ctx)
			log["Data"] = map[string]interface{}{"a": 0}
			panic("Some error")
		})
		srv := app.Start()
		defer srv.Close()

		res, err := RequestBy("POST", "http://"+srv.Addr().String())
		assert.Nil(err)
		assert.Equal(500, res.StatusCode)
		assert.Equal("text/plain; charset=utf-8", res.Header.Get(gear.HeaderContentType))
		time.Sleep(10 * time.Millisecond)
		log := buf.String()
		assert.Contains(log, time.Now().UTC().Format(time.RFC3339)[0:19])
		assert.Contains(log, " INFO ")
		assert.Contains(log, `"Data":{"a":0}`)
		assert.Contains(log, `"Method":"POST"`)
		assert.Contains(log, `"Status":500`)
		assert.Contains(log, `"UserAgent":`)
		assert.Contains(errbuf.String(), "Some error")
		res.Body.Close()
	})

	t.Run("Color", func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(ColorGreen, colorStatus(200))
		assert.Equal(ColorGreen, colorStatus(204))
		assert.Equal(ColorCyan, colorStatus(304))
		assert.Equal(ColorYellow, colorStatus(404))
		assert.Equal(ColorRed, colorStatus(504))
	})
}