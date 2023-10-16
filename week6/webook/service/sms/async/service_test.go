package async

import (
	"go.uber.org/mock/gomock"
	"testing"
	"week6/webook/domain"
	"week6/webook/repository"
	smsrepomocks "week6/webook/repository/mocks"
	"week6/webook/service/sms"
	smsmocks "week6/webook/service/sms/mocks"
)

// 需要准备数据库？

func TestAsyncSMSService_AsyncSendSMS(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository)
	}{
		{
			name: "异步发送短信成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository) {
				svc := smsmocks.NewMockService(ctrl)
				repo := smsrepomocks.NewMockSMSRepository(ctrl)
				repo.EXPECT().PreemptWaitingSMS(gomock.Any()).Return(domain.SMS{
					Id:      1,
					TplId:   "1",
					Args:    []string{"123456"},
					Numbers: []string{"15012345678"},
				}, nil)
				svc.EXPECT().Send(gomock.Any(), "1", "123456", "15012345678").Return(nil)
				return svc, repo
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			_, _ = tc.mock(ctrl)
		})
	}
}
