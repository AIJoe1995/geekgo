package tencent

// https://cloud.tencent.com/document/product/382/43199

import (
	"context"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   sms.Client
	appId    string
	signName string
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, phone []string) error {
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(s.appId)
	request.SignName = common.StringPtr(s.signName)
	request.TemplateId = common.StringPtr(tpl)
	request.TemplateParamSet = common.StringPtrs(args)
	request.PhoneNumberSet = common.StringPtrs(phone)
	response, err := s.client.SendSms(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("An API error has returned: %s", err)
	}

	//// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
	//if err != nil {
	//	panic(err)
	//}

	for _, status := range response.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, 原因：%s", *status.Code, *status.Message)
		}
	}
	//b, _ := json.Marshal(response.Response)
	return nil

}
