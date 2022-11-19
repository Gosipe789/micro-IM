package req

import "testing"

func TestHttpPost(t *testing.T) {
	type args struct {
		url       string
		parameter map[string]interface{}
		data      interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				url: "https://api.telegram.org/bot5780070107:AAFtrN_OJtOzD8J1E5DyjXza-U40vB_3QGA/sendMessage",
				parameter: map[string]interface{}{
					"chat_id": "-833718329",
					"text":    "监测到 \nTUwQj4nEjAYFT8zHSJhb4isUYkfRJGCGPo\n向\nTQM2zUnDszVBYoRG9YNoCo6j4cYVWnCGPo\n转账金额\n100000.00 USDT",
				},
				data: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HttpPost(tt.args.url, tt.args.parameter, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("HttpPost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
