package logit

import (
	"errors"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestBuildDispatchCore(t *testing.T) {
	type args struct {
		ruleName       string
		filename       string
		dispatchRules  []ZapDispatch
		writerBuilder  WriterBuilder
		encoderBuilder EncoderBuilder
		opts           []ZapWriterOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ruleName: "1hour",
				filename: "testdata/dispatch_core_1.log",
				dispatchRules: []ZapDispatch{
					{
						FileSuffix: "wf",
						Levels: []zapcore.Level{
							zapcore.WarnLevel,
							zapcore.ErrorLevel,
						},
						EncoderBuilder: DefaultEncoder,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				ruleName: "1hour",
				filename: "testdata/dispatch_core_1.log",
				dispatchRules: []ZapDispatch{
					{
						FileSuffix: "wf",
						Levels: []zapcore.Level{
							zapcore.WarnLevel,
							zapcore.ErrorLevel,
						},
					},
				},
				writerBuilder: func(ruleName, filename string, opts ...ZapWriterOptions) (zapcore.WriteSyncer, error) {
					return nil, errors.New("test error")
				},
			},

			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, closeFuncs, err := BuildDispatchCore(tt.args.ruleName, tt.args.filename, tt.args.dispatchRules, tt.args.writerBuilder, tt.args.encoderBuilder, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDispatchCore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if closeFuncs != nil {
				closeFuncs()
			}
		})
	}
}
