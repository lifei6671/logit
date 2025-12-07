package logit

import (
	"context"
	"testing"
)

func TestRemoveField(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "RemoveField",
			args: args{
				ctx: func() context.Context {
					ctx := NewContext(context.Background())
					AddInfo(ctx, String("key", "text"))
					return ctx
				}(),
				key: "key",
			},
		}, {
			name: "RemoveField_Meta",
			args: args{
				ctx: func() context.Context {
					ctx := NewContext(context.Background())
					AddMetaField(ctx, String("key1", "text"))
					return ctx
				}(),
				key: "key1",
			},
		},
		{
			name: "RemoveField_NormalOrder",
			args: args{
				ctx: func() context.Context {
					ctx := NewContext(context.Background())
					AddField(ctx, String("key2", "text"))
					return ctx
				}(),
				key: "key2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveField(tt.args.ctx, tt.args.key)
			_, has := FindField(tt.args.ctx, tt.args.key)
			if has {
				t.Errorf("RemoveField() has = %v, want = %v", has, false)
			}
		})
	}
}

func TestFindMetaField(t *testing.T) {
	type args struct {
		ctx  context.Context
		key  string
		wait struct {
			key string
			has bool
		}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "FindMetaField",
			args: args{
				ctx: func() context.Context {
					ctx := NewContext(context.Background())
					AddMetaField(ctx, String("key", "text"))
					return WithContext(ctx)
				}(),
				key: "key",
				wait: struct {
					key string
					has bool
				}{
					key: "key",
					has: true,
				},
			},
		},
		{
			name: "FindMetaField_Fail",
			args: args{
				ctx: func() context.Context {
					ctx := WithContext(context.Background())
					AddMetaField(ctx, String("key", "text"))
					AddMetaField(ctx, String("key", "text"))
					return ctx
				}(),
				key: "key1",
				wait: struct {
					key string
					has bool
				}{
					key: "",
					has: false,
				},
			},
		},
		{
			name: "FindMetaField_Empty_Ctx",
			args: args{
				ctx: func() context.Context {
					ctx := context.Background()
					AddField(ctx, String("key", "text"))
					AddWarn(ctx, String("key", "text"))
					AddDebug(ctx, String("key", "text"))
					AddError(ctx, String("key", "text"))
					AddFatal(ctx, String("key", "text"))
					AddInfo(ctx, String("key", "text"))
					return ctx
				}(),
				key: "key",
				wait: struct {
					key string
					has bool
				}{key: "", has: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, has := FindMetaField(tt.args.ctx, tt.args.key)

			if tt.args.wait.has && !has {
				t.Errorf("FindMetaField() has = %v, want = %v", has, true)
			}
			if !tt.args.wait.has && has {
				t.Errorf("FindMetaField() has = %v, want = %v", has, false)
			}
			if tt.args.wait.key != field.Key {
				t.Errorf("FindMetaField() key = %v, want = %v", field.Key, tt.args.wait.key)
			}
		})
	}
}
