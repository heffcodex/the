package tzap

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	httpRequestMaskHeaders = map[string]struct{}{
		"Api-Key":             {},
		"Api-Token":           {},
		"Authorization":       {},
		"Cookie":              {},
		"Key":                 {},
		"Proxy-Authorization": {},
		"Token":               {},
		"X-Api-Key":           {},
		"X-Api-Token":         {},
		"X-Key":               {},
		"X-Request-Key":       {},
		"X-Request-Token":     {},
		"X-Token":             {},
	}
	httpHeaderMask = func(vs ...string) []string {
		res := make([]string, len(vs))

		for i, v := range vs {
			res[i] = "..*" + strconv.Itoa(len(v)) + "*.."
		}

		return res
	}

	_ zapcore.ObjectMarshaler = (*httpRequestMarshaler)(nil)
)

type httpRequestMarshaler struct {
	Method        string
	URL           string
	Proto         string
	ContentLength int64
	Headers       map[string][]string
}

func (m *httpRequestMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("method", m.Method)
	enc.AddString("url", m.URL)
	enc.AddString("proto", m.Proto)
	enc.AddInt64("contentLength", m.ContentLength)

	_ = enc.AddObject("headers", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, vs := range m.Headers {
			_ = enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
				for _, v := range vs {
					enc.AppendString(v)
				}

				return nil
			}))
		}

		return nil
	}))

	return nil
}

func HTTPRequest(r *http.Request) zap.Field {
	headers := make(map[string][]string, len(r.Header))

	for k, vs := range r.Header {
		if _, mask := httpRequestMaskHeaders[k]; mask {
			headers[k] = httpHeaderMask(vs...)
		} else {
			headers[k] = slices.Clone(vs)
		}
	}

	return zap.Field{
		Key:  KeyHTTPRequest,
		Type: zapcore.ObjectMarshalerType,
		Interface: &httpRequestMarshaler{
			Method:        r.Method,
			URL:           r.URL.String(),
			Proto:         r.Proto,
			ContentLength: r.ContentLength,
			Headers:       headers,
		},
	}
}

func FastHTTPRequest(r *fasthttp.Request) zap.Field {
	headers := make(map[string][]string, r.Header.Len())

	r.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)

		if _, mask := httpRequestMaskHeaders[k]; mask {
			headers[k] = append(headers[k], httpHeaderMask(v)...)
		} else {
			headers[k] = append(headers[k], v)
		}
	})

	return zap.Field{
		Key:  KeyHTTPRequest,
		Type: zapcore.ObjectMarshalerType,
		Interface: &httpRequestMarshaler{
			Method:        string(r.Header.Method()),
			URL:           r.URI().String(),
			Proto:         string(r.Header.Protocol()),
			ContentLength: int64(r.Header.ContentLength()),
			Headers:       headers,
		},
	}
}
