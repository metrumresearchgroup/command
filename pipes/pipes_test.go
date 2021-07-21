package pipes_test

/*
func TestNew(t *testing.T) {
	type args struct {
		options []pipes.Option
	}
	tests := []struct {
		name string
		args args
		want pipes.Pipes
	}{
		{
			name: "no options",
			args: args{
				options: nil,
			},
			want: pipes.Pipes{},
		},
		{
			name: "default pipes",
			want: pipes.Pipes{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			},
		},
		{
			name: "wired pipes",
			args: args{
				[]pipes.Option{
					pipes.WithStdout(os.Stdout),
					pipes.WithStderr(os.Stderr),
					pipes.WithStdin(os.Stdin),
				},
			},
			want: pipes.Pipes{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("pipes output matches inputs", func(t *testing.T) {
				if got := pipes.New(tt.args.options...); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("New() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}
*/
