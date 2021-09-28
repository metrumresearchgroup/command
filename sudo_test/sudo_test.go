package sudo_test

import (
	"os/user"
	"testing"

	"github.com/metrumresearchgroup/wrapt"

	. "github.com/metrumresearchgroup/command"
)

func TestImpersonate_sudo(tt *testing.T) {
	t := wrapt.WrapT(tt)

	cmd := New("ls", "/")

	u, err := user.Current()
	t.R.NoError(err)

	t.R.NoError(cmd.Impersonate(u.Name, false))

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(string(out))
}
