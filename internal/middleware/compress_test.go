package middleware_test

// func TestGzipCompression(t *testing.T) {

// 	t.Run("sends_gzip", func(t *testing.T) {
// 		ctrl := gomock.NewController(t)
// 		defer ctrl.Finish()
// 		m := mocks.NewMockURLServiceInterface(ctrl)
// 		m.EXPECT().AddShortURL(gomock.Any()).AnyTimes()
// 		m.EXPECT().MakeShortURLByID(gomock.Any()).AnyTimes()

// 		buf := bytes.NewBuffer(nil)
// 		zb := gzip.NewWriter(buf)
// 		_, err := zb.Write([]byte("http://test.ru"))
// 		require.NoError(t, err)
// 		err = zb.Close()
// 		require.NoError(t, err)

// 		handler := middleware.GzipMiddleware(handler.RootHandler(m))

// 		req := httptest.NewRequest(http.MethodPost, "/", buf)
// 		req.Header.Set("Content-Encoding", "gzip")
// 		req.Header.Set("Content-Type", "text/plain")
// 		req.Header.Set("Accept-Encoding", "")

// 		rw := httptest.NewRecorder()

// 		r := chi.NewRouter()
// 		r.Post("/", handler)

// 		r.ServeHTTP(rw, req)

// 		resp := rw.Result()
// 		defer resp.Body.Close()

// 		require.NoError(t, err)
// 		require.Equal(t, http.StatusCreated, resp.StatusCode)

// 	})
// }
