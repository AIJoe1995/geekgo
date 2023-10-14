package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"geekgo/week5/webook/domain"
	"geekgo/week5/webook/service"
	svcmocks "geekgo/week5/webook/service/mocks"
	ijwt "geekgo/week5/webook/web/jwt"
	jwtmocks "geekgo/week5/webook/web/jwt/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler)
		req      LoginSMSReq
		wantCode int
		wantBody Result
	}{
		{
			name: "短信登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				svc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15012345678", "123456").Return(true, nil)
				svc.EXPECT().FindOrCreate(gomock.Any(), "15012345678").Return(domain.User{
					Id:    1,
					Phone: "15012345678",
				}, nil)
				jwtHdl.EXPECT().SetLoginToken(gomock.Any(), int64(1)).Return(nil)
				return svc, codesvc, jwtHdl
			},
			req: LoginSMSReq{
				Phone: "15012345678",
				Code:  "123456",
			},
			wantCode: 200,
			wantBody: Result{
				Msg: "验证码校验通过",
			},
		},
		{
			name: "短信验证时出现系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				svc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15012345678", "123456").Return(false, errors.New("mock codesvc error"))

				return svc, codesvc, jwtHdl
			},
			req: LoginSMSReq{
				Phone: "15012345678",
				Code:  "123456",
			},
			wantCode: 200,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "短信验证失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				svc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15012345678", "123456").Return(false, nil)

				return svc, codesvc, jwtHdl
			},
			req: LoginSMSReq{
				Phone: "15012345678",
				Code:  "123456",
			},
			wantCode: 200,
			wantBody: Result{
				Code: 4,
				Msg:  "验证码有误",
			},
		},
		{
			name: "短信登录-验证码验证通过-数据库访问失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				svc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15012345678", "123456").Return(true, nil)
				svc.EXPECT().FindOrCreate(gomock.Any(), "15012345678").Return(domain.User{}, errors.New("mock db error"))

				return svc, codesvc, nil
			},
			req: LoginSMSReq{
				Phone: "15012345678",
				Code:  "123456",
			},
			wantCode: 200,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "短信登录-jwtsetlogincode失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				svc := svcmocks.NewMockUserService(ctrl)
				codesvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)

				codesvc.EXPECT().Verify(gomock.Any(), "login", "15012345678", "123456").Return(true, nil)
				svc.EXPECT().FindOrCreate(gomock.Any(), "15012345678").Return(domain.User{
					Id:    1,
					Phone: "15012345678",
				}, nil)
				jwtHdl.EXPECT().SetLoginToken(gomock.Any(), int64(1)).Return(errors.New("mock jwt error"))
				return svc, codesvc, jwtHdl
			},
			req: LoginSMSReq{
				Phone: "15012345678",
				Code:  "123456",
			},
			wantCode: 200,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, codesvc, jwtHdl := tc.mock(ctrl)
			uh := NewUserHandler(codesvc, svc, jwtHdl)

			b, err := json.Marshal(tc.req)
			require.NoError(t, err)
			reqBody := bytes.NewBuffer(b)
			// 构造请求
			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", reqBody)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server := gin.Default()
			uh.RegisterRoutes(server)
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)

			var webRes Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, webRes)

		})
	}
}
