/*
 * Copyright 2015 Xuyuan Pang
 * Author: Xuyuan Pang
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package encoding

import (
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Xuyuanp/hador"
	"github.com/smartystreets/goconvey/convey"
)

func newSimpleHandler(content string) hador.HandlerFunc {
	return func(ctx *hador.Context) {
		ctx.Response.Write([]byte(content))
	}
}

type simpleResponeWriter struct{}

func (srw *simpleResponeWriter) Header() http.Header           { return nil }
func (srw *simpleResponeWriter) Write(buf []byte) (int, error) { return 0, nil }
func (srw *simpleResponeWriter) WriteHeader(code int)          {}

func TestAcceptEncoding(t *testing.T) {
	convey.Convey("Test AcceptEncoding", t, func() {
		var (
			c1 AcceptEncoding = "compress, gzip"
			c2 AcceptEncoding
			c3 AcceptEncoding = "*"
			c4 AcceptEncoding = "compress;q=0.5, gzip;q=1.0"
			c5 AcceptEncoding = "*;q=0, compress;q=0.3"
			c6 AcceptEncoding = "*;q=0, gzip;q=0.2"
			c7 AcceptEncoding = "gzip;q=0"
		)
		convey.So(c1.Accept(encodingTypeGZip), convey.ShouldBeTrue)
		convey.So(c2.Accept(encodingTypeGZip), convey.ShouldBeTrue)
		convey.So(c2.Accept(encodingTypeDeflate), convey.ShouldBeFalse)
		convey.So(c3.Accept(encodingTypeGZip), convey.ShouldBeTrue)
		convey.So(c4.Accept(encodingTypeGZip), convey.ShouldBeTrue)
		convey.So(c5.Accept(encodingTypeGZip), convey.ShouldBeFalse)
		convey.So(c6.Accept(encodingTypeGZip), convey.ShouldBeTrue)
		convey.So(c7.Accept(encodingTypeGZip), convey.ShouldBeFalse)
	})
}

func TestGZipFilter(t *testing.T) {
	convey.Convey("Test GZipFilter", t, func() {
		h := hador.New()
		h.Before(GZipFilter(true))
		h.Get("/foo", newSimpleHandler("gzip"))

		convey.Convey("Test accept", func() {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/foo", nil)
			convey.So(err, convey.ShouldBeNil)

			h.ServeHTTP(resp, req)

			convey.So(resp.Code, convey.ShouldEqual, http.StatusOK)
			reader, err := gzip.NewReader(resp.Body)
			convey.So(err, convey.ShouldBeNil)
			body, err := ioutil.ReadAll(reader)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(body), convey.ShouldEqual, "gzip")
			convey.So(resp.Header().Get(headerContentEncoding),
				convey.ShouldEqual,
				encodingTypeGZip)
		})
		convey.Convey("Test not accept", func() {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/foo", nil)
			convey.So(err, convey.ShouldBeNil)
			req.Header.Set(headerAcceptEncoding, encodingTypeDeflate)

			h.ServeHTTP(resp, req)

			convey.So(resp.Code, convey.ShouldEqual, http.StatusNotAcceptable)
		})
	})
}

func TestDeflateFilter(t *testing.T) {
	convey.Convey("Test GZipFilter", t, func() {
		h := hador.New()
		h.Before(DeflateFilter(flate.DefaultCompression, true))
		h.Get("/foo", newSimpleHandler("deflate"))

		convey.Convey("Test accept", func() {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/foo", nil)
			req.Header.Set(headerAcceptEncoding, encodingTypeDeflate)
			convey.So(err, convey.ShouldBeNil)

			h.ServeHTTP(resp, req)

			convey.So(resp.Code, convey.ShouldEqual, http.StatusOK)
			reader := flate.NewReader(resp.Body)
			convey.So(err, convey.ShouldBeNil)
			body, err := ioutil.ReadAll(reader)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(body), convey.ShouldEqual, "deflate")
			convey.So(resp.Header().Get(headerContentEncoding), convey.ShouldEqual, encodingTypeDeflate)
		})
		convey.Convey("Test not accept", func() {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/foo", nil)
			convey.So(err, convey.ShouldBeNil)

			h.ServeHTTP(resp, req)

			convey.So(resp.Code, convey.ShouldEqual, http.StatusNotAcceptable)
		})
	})
}
