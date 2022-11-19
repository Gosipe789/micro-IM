package tg

import "testing"

func TestTG_SendMsg(t1 *testing.T) {
	type fields struct {
		Url string
	}
	type args struct {
		parameter map[string]interface{}
		body      interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				Url: "https://api.telegram.org/bot5780070107:AAFtrN_OJtOzD8J1E5DyjXza-U40vB_3QGA/sendMessage",
			},
			args: args{
				parameter: map[string]interface{}{
					"chat_id": "-833718329",
					"text":    "监测到已激活账号 \n地址：TLbeeBoiRdarUbdYd4NyRWMFhNDvM6Bym1\n密钥：95eee42346078726f5969fb7640fd8c1eec04dc4e4e45f94f0a54e92480d62a8",
				},
				body: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TG{
				Url: tt.fields.Url,
			}
			if err := t.SendMsg(tt.args.parameter, tt.args.body); (err != nil) != tt.wantErr {
				t1.Errorf("SendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
