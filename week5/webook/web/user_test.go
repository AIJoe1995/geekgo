package web

import (
	"bytes"
	"context"

	"encoding/json"
	"geekgo/week5/webook/domain"
	"geekgo/week5/webook/service"
	svcmocks "geekgo/week5/webook/service/mocks"
	"geekgo/week5/webook/web/jwt"
	jwtmocks "geekgo/week5/webook/web/jwt/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler)

		reqBody  string
		biz      string
		phone    string
		code     string
		wantCode int
		wantBody Result
	}{
		{
			name: "验证码校验通过",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15212345678", "123456").Return(true, nil)
				usersvc.EXPECT().FindOrCreate(gomock.Any(), "15212345678").Return(domain.User{
					Id:    1,
					Phone: "15212345678",
				}, nil)
				jwtHdl.EXPECT().SetLoginToken(context.Background(), 1).Return(nil)

				return usersvc, codesvc, jwtHdl
			},
			reqBody: `
{
"phone": "15212345678", 
"code": "123456"
}`,
			biz:      "login",
			phone:    "15212345678",
			code:     "123456",
			wantCode: 200,
			wantBody: Result{Msg: "验证码校验通过"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name,
			func(t *testing.T) {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				usersvc, codesvc, jwtHdl := tc.mock(ctrl)

				// 测试userhandler.LoginSMS
				userHdl := NewUserHandler(codesvc, usersvc, jwtHdl) // codeSvc svc jwtHdl 需要mock实现

				server := gin.Default()

				userHdl.RegisterRoutes(server)

				req, err := http.NewRequest(http.MethodPost,
					"/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")

				resp := httptest.NewRecorder()
				//server.Use(func(c *gin.Context) {
				//	c.Set("user", jwt.UserClaims{})
				//}) // jwt怎么用？

				server.ServeHTTP(resp, req)
				assert.Equal(t, tc.wantCode, resp.Code)
				var webRes Result
				err = json.NewDecoder(resp.Body).Decode(&webRes)
				require.NoError(t, err)
				assert.Equal(t, tc.wantBody, webRes)

			})
	}
}
