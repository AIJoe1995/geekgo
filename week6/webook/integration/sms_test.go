package integration

import (
	"errors"
	"fmt"
	"github.com/ecodeclub/ekit/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
	"time"
	"week6/webook/integration/startup"
	"week6/webook/repository/dao"
	"week6/webook/service/sms"
	smsmocks "week6/webook/service/sms/mocks"
)

type AsyncSMSTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *AsyncSMSTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
}

func (s *AsyncSMSTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE table `sms`")
}

func (s *AsyncSMSTestSuite) TestAsyncSend() {
	// before 向数据库中插入待发送的短信
	// 发送之后 检查发送结果和数据库这条短信的标记
	now := time.Now()
	fmt.Println(now.UnixMilli())
	testCases := []struct {
		name string
		// 虽然是集成测试，但是我们也不想真的发短信，所以用 mock
		mock func(ctrl *gomock.Controller) sms.Service
		// 准备数据
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		//{
		//	name: "empty test",
		//	mock: func(ctrl *gomock.Controller) sms.Service {
		//		svc := smsmocks.NewMockService(ctrl)
		//		svc.EXPECT().Send(gomock.Any(), "123",
		//			[]string{"123456"}, []string{"15212345678"}).Return(nil)
		//		return svc
		//	},
		//	before: func(t *testing.T) {
		//
		//	},
		//	after: func(t *testing.T) {
		//
		//	},
		//},
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "123",
					[]string{"123456"}, []string{"15212345678"}).Return(nil)
				return svc
			},
			before: func(t *testing.T) {
				// 准备一条数据
				err := s.db.Create(&dao.SMS{
					Id: 1,
					Config: sqlx.JsonColumn[dao.SmsConfig]{
						Val: dao.SmsConfig{
							TplId:   "123",
							Args:    []string{"123456"},
							Numbers: []string{"15212345678"},
						},
						Valid: true,
					},
					RetryMax: 3,
					Status:   0,
					Ctime:    now.Add(-time.Minute * 2).UnixMilli(),
					Utime:    now.Add(-time.Minute * 2).UnixMilli(),
				}).Error
				assert.NoError(t, err)
				time.Sleep(time.Minute * 2)
			},
			after: func(t *testing.T) {
				// 验证数据
				var as dao.SMS
				err := s.db.Where("id=?", 1).First(&as).Error
				assert.NoError(t, err)
				assert.Equal(t, uint8(2), as.Status)
			},
		},
		{
			name: "发送失败，标记为失败",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "123",
					[]string{"123456"}, []string{"15212345678"}).
					Return(errors.New("模拟失败"))
				return svc
			},
			before: func(t *testing.T) {
				// 准备一条数据
				err := s.db.Create(&dao.SMS{
					Id: 1,
					Config: sqlx.JsonColumn[dao.SmsConfig]{
						Val: dao.SmsConfig{
							TplId:   "123",
							Args:    []string{"123456"},
							Numbers: []string{"15212345678"},
						},
						Valid: true,
					},
					RetryMax: 3,
					RetryCnt: 2,
					Status:   0,
					Ctime:    now.Add(-time.Minute * 2).UnixMilli(),
					Utime:    now.Add(-time.Minute * 2).UnixMilli(),
				}).Error
				assert.NoError(t, err)

			},
			after: func(t *testing.T) {
				// 验证数据
				var as dao.SMS
				err := s.db.Where("id=?", 1).First(&as).Error
				assert.NoError(t, err)
				assert.Equal(t, uint8(1), as.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := startup.InitAsyncSmsService(tc.mock(ctrl))
			tc.before(t)
			time.Sleep(time.Minute)
			defer tc.after(t)
			svc.AsyncSendSMS()
		})
	}
}

func TestAsyncSmsService(t *testing.T) {
	suite.Run(t, &AsyncSMSTestSuite{})
}
